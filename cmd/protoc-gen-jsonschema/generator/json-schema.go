// Copyright 2021 Google LLC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package generator

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/google/gnostic/jsonschema"
)

var (
	typeString  = "string"
	typeNumber  = "number"
	typeInteger = "integer"
	typeBoolean = "boolean"
	typeObject  = "object"
	typeArray   = "array"

	formatDate     = "date"
	formatDateTime = "date-time"
	formatEnum     = "enum"
	formatBytes    = "bytes"
)

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
}

type Configuration struct {
	BaseURL  *string
	Version  *string
	Naming   *string
	EnumType *string
}

// JSONSchemaGenerator holds internal state needed to generate the JSON Schema documents for a transcoded Protocol Buffer service.
type JSONSchemaGenerator struct {
	conf   Configuration
	plugin *protogen.Plugin

	linterRulePattern *regexp.Regexp
}

// NewJSONSchemaGenerator creates a new generator for a protoc plugin invocation.
func NewJSONSchemaGenerator(plugin *protogen.Plugin, conf Configuration) *JSONSchemaGenerator {
	baseURL := *conf.BaseURL
	if len(baseURL) > 0 && baseURL[len(baseURL)-1:] != "/" {
		baseURL += "/"
	}
	conf.BaseURL = &baseURL

	return &JSONSchemaGenerator{
		conf:   conf,
		plugin: plugin,

		linterRulePattern: regexp.MustCompile(`\(-- .* --\)`),
	}
}

// Run runs the generator.
func (g *JSONSchemaGenerator) Run() error {
	for _, file := range g.plugin.Files {
		if file.Generate {
			schemas := g.buildSchemasFromMessages(file.Messages)
			for _, schema := range schemas {
				outputFile := g.plugin.NewGeneratedFile(fmt.Sprintf("%s.json", schema.Name), "")
				outputFile.Write([]byte(schema.Value.JSONString()))
			}
		}
	}

	return nil
}

// filterCommentString removes line breaks and linter rules from comments.
func (g *JSONSchemaGenerator) filterCommentString(c protogen.Comments, removeNewLines bool) string {
	comment := string(c)
	if removeNewLines {
		comment = strings.Replace(comment, "\n", "", -1)
	}
	comment = g.linterRulePattern.ReplaceAllString(comment, "")
	return strings.TrimSpace(comment)
}

func (g *JSONSchemaGenerator) formatMessageNameString(name string) string {
	if *g.conf.Naming == "proto" {
		return name
	}

	if len(name) > 1 {
		return strings.ToUpper(name[0:1]) + name[1:]
	}

	if len(name) == 1 {
		return strings.ToLower(name)
	}

	return name
}

func (g *JSONSchemaGenerator) formatFieldName(field *protogen.Field) string {
	if *g.conf.Naming == "proto" {
		return string(field.Desc.Name())
	}

	return field.Desc.JSONName()
}

// messageDefinitionName builds the full schema definition name of a message.
func messageDefinitionName(desc protoreflect.MessageDescriptor) string {
	name := string(desc.Name())

	pkg := string(desc.ParentFile().Package())
	parentName := desc.Parent().FullName()
	if len(parentName) > len(pkg) {
		parentName = parentName[len(pkg)+1:]
		name = fmt.Sprintf("%s.%s", parentName, name)
	}

	return strings.Replace(name, ".", "_", -1)
}

func (g *JSONSchemaGenerator) schemaOrReferenceForType(desc protoreflect.MessageDescriptor) *jsonschema.Schema {
	// Create the full typeName
	typeName := fmt.Sprintf(".%s.%s", desc.ParentFile().Package(), desc.Name())

	switch typeName {

	case ".google.protobuf.Timestamp":
		// Timestamps are serialized as strings
		return &jsonschema.Schema{Type: &jsonschema.StringOrStringArray{String: &typeString}, Format: &formatDateTime}

	case ".google.type.Date":
		// Dates are serialized as strings
		return &jsonschema.Schema{Type: &jsonschema.StringOrStringArray{String: &typeString}, Format: &formatDate}

	case ".google.type.DateTime":
		// DateTimes are serialized as strings
		return &jsonschema.Schema{Type: &jsonschema.StringOrStringArray{String: &typeString}, Format: &formatDateTime}

	case ".google.protobuf.Struct":
		// Struct is equivalent to a JSON object
		return &jsonschema.Schema{Type: &jsonschema.StringOrStringArray{String: &typeObject}}

	case ".google.protobuf.Value":
		// Value is equivalent to any JSON value except null
		return &jsonschema.Schema{
			Type: &jsonschema.StringOrStringArray{
				StringArray: &[]string{typeString, typeNumber, typeInteger, typeBoolean, typeObject, typeArray},
			},
		}

	case ".google.protobuf.Empty":
		// Empty is close to JSON undefined than null, so ignore this field
		return nil
	}

	typeName = messageDefinitionName(desc)
	ref := "#/definitions/" + g.formatMessageNameString(typeName)
	return &jsonschema.Schema{Ref: &ref}
}

func (g *JSONSchemaGenerator) schemaOrReferenceForField(field protoreflect.FieldDescriptor, definitions *[]*jsonschema.NamedSchema) *jsonschema.Schema {
	if field.IsMap() {
		typ := "object"
		return &jsonschema.Schema{
			Type: &jsonschema.StringOrStringArray{String: &typ},
			AdditionalProperties: &jsonschema.SchemaOrBoolean{
				Schema: g.schemaOrReferenceForField(field.MapValue(), definitions),
			},
		}
	}

	var kindSchema *jsonschema.Schema

	kind := field.Kind()

	switch kind {

	case protoreflect.MessageKind:
		kindSchema = g.schemaOrReferenceForType(field.Message())
		if kindSchema == nil {
			return nil
		}

		if kindSchema.Ref != nil {
			if !refInDefinitions(*kindSchema.Ref, definitions) {
				ref := strings.Replace(*kindSchema.Ref, "#/definitions/", *g.conf.BaseURL, 1)
				ref += ".json"
				kindSchema.Ref = &ref
			}
		}

	case protoreflect.StringKind:
		kindSchema = &jsonschema.Schema{Type: &jsonschema.StringOrStringArray{String: &typeString}}

	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Uint32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Uint64Kind,
		protoreflect.Sfixed32Kind, protoreflect.Fixed32Kind, protoreflect.Sfixed64Kind,
		protoreflect.Fixed64Kind:
		format := kind.String()
		kindSchema = &jsonschema.Schema{Type: &jsonschema.StringOrStringArray{String: &typeInteger}, Format: &format}

	case protoreflect.EnumKind:
		kindSchema = &jsonschema.Schema{Format: &formatEnum}
		if g.conf.EnumType != nil && *g.conf.EnumType == typeString {
			kindSchema.Type = &jsonschema.StringOrStringArray{String: &typeString}
			kindSchema.Enumeration = &[]jsonschema.SchemaEnumValue{}
			for i := 0; i < field.Enum().Values().Len(); i++ {
				name := string(field.Enum().Values().Get(i).Name())
				*kindSchema.Enumeration = append(*kindSchema.Enumeration, jsonschema.SchemaEnumValue{String: &name})
			}
		} else {
			kindSchema.Type = &jsonschema.StringOrStringArray{String: &typeInteger}
		}

	case protoreflect.BoolKind:
		kindSchema = &jsonschema.Schema{Type: &jsonschema.StringOrStringArray{String: &typeBoolean}}

	case protoreflect.FloatKind, protoreflect.DoubleKind:
		format := kind.String()
		kindSchema = &jsonschema.Schema{Type: &jsonschema.StringOrStringArray{String: &typeNumber}, Format: &format}

	case protoreflect.BytesKind:
		kindSchema = &jsonschema.Schema{Type: &jsonschema.StringOrStringArray{String: &typeString}, Format: &formatBytes}

	default:
		log.Printf("(TODO) Unsupported field type: %+v", field.Message().FullName())
	}

	if field.IsList() {
		typ := "array"
		return &jsonschema.Schema{
			Type: &jsonschema.StringOrStringArray{String: &typ},
			Items: &jsonschema.SchemaOrSchemaArray{
				Schema: kindSchema,
			},
		}
	}

	return kindSchema
}

// buildSchemasFromMessages creates a schema for each message.
func (g *JSONSchemaGenerator) buildSchemasFromMessages(messages []*protogen.Message) []*jsonschema.NamedSchema {
	schemas := []*jsonschema.NamedSchema{}

	// For each message, generate a schema.
	for _, message := range messages {
		schemaName := string(message.Desc.Name())
		typ := "object"
		id := fmt.Sprintf("%s%s.json", *g.conf.BaseURL, schemaName)

		schema := &jsonschema.NamedSchema{
			Name: schemaName,
			Value: &jsonschema.Schema{
				Schema:     g.conf.Version,
				ID:         &id,
				Type:       &jsonschema.StringOrStringArray{String: &typ},
				Title:      &schemaName,
				Properties: &[]*jsonschema.NamedSchema{},
			},
		}

		description := g.filterCommentString(message.Comments.Leading, true)
		if description != "" {
			schema.Value.Description = &description
		}

		// Any embedded messages will be created as definitions
		if message.Messages != nil {
			if schema.Value.Definitions == nil {
				schema.Value.Definitions = &[]*jsonschema.NamedSchema{}
			}

			for _, subMessage := range message.Messages {
				subSchemas := g.buildSchemasFromMessages([]*protogen.Message{subMessage})
				if len(subSchemas) != 1 {
					continue
				}
				subSchema := subSchemas[0]
				subSchema.Value.ID = nil
				subSchema.Value.Schema = nil
				subSchema.Name = messageDefinitionName(subMessage.Desc)

				if subSchema.Value.Definitions != nil {
					*schema.Value.Definitions = append(*schema.Value.Definitions, *subSchema.Value.Definitions...)
					subSchema.Value.Definitions = nil
				}

				*schema.Value.Definitions = append(*schema.Value.Definitions, subSchemas...)
			}
		}

		if message.Desc.IsMapEntry() {
			continue
		}

		for _, field := range message.Fields {
			// The field is either described by a reference or a schema.
			fieldSchema := g.schemaOrReferenceForField(field.Desc, schema.Value.Definitions)
			if fieldSchema == nil {
				continue
			}

			// Handle readonly and writeonly properties, if the schema version can handle it.
			if getSchemaVersion(schema.Value) >= "07" {
				t := true
				// Check the field annotations to see if this is a readonly field.
				extension := proto.GetExtension(field.Desc.Options(), annotations.E_FieldBehavior)
				if extension != nil {
					switch v := extension.(type) {
					case []annotations.FieldBehavior:
						for _, vv := range v {
							if vv == annotations.FieldBehavior_OUTPUT_ONLY {
								fieldSchema.ReadOnly = &t
							} else if vv == annotations.FieldBehavior_INPUT_ONLY {
								fieldSchema.WriteOnly = &t
							}
						}
					default:
						log.Printf("unsupported extension type %T", extension)
					}
				}
			}

			fieldName := g.formatFieldName(field)
			// Do not add title for ref values
			if fieldSchema.Ref == nil {
				fieldSchema.Title = &fieldName
			}

			// Get the field description from the comments.
			description := g.filterCommentString(field.Comments.Leading, true)
			if description != "" {
				// Note: Description will be ignored if $ref is set, but is still useful
				fieldSchema.Description = &description
			}

			*schema.Value.Properties = append(
				*schema.Value.Properties,
				&jsonschema.NamedSchema{
					Name:  fieldName,
					Value: fieldSchema,
				},
			)
		}

		schemas = append(schemas, schema)
	}

	return schemas
}

var reSchemaVersion = regexp.MustCompile(`https*://json-schema.org/draft[/-]([^/]+)/schema`)

func getSchemaVersion(schema *jsonschema.Schema) string {
	schemaSchema := *schema.Schema
	matches := reSchemaVersion.FindStringSubmatch(schemaSchema)
	if len(matches) == 2 {
		return matches[1]
	}
	return ""
}

func refInDefinitions(ref string, definitions *[]*jsonschema.NamedSchema) bool {
	if definitions == nil {
		return false
	}
	ref = strings.TrimPrefix(ref, "#/definitions/")
	for _, def := range *definitions {
		if ref == def.Name {
			return true
		}
	}
	return false
}
