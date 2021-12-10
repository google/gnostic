// Copyright 2020 Google LLC. All Rights Reserved.
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
	"net/url"
	"regexp"
	"strings"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/google/gnostic/jsonschema"
)

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
}

type Configuration struct {
	BaseURL *string
	Version *string
	Naming  *string
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

// schemaReferenceForTypeName returns a JSON Schema definitions reference.
func (g *JSONSchemaGenerator) schemaReferenceForTypeName(typeName string) string {
	parts := strings.Split(typeName, ".")
	lastPart := parts[len(parts)-1]
	return "#/definitions/" + g.formatMessageNameString(lastPart)
}

// fullMessageTypeName builds the full type name of a message.
func fullMessageTypeName(message protoreflect.MessageDescriptor) string {
	name := string(message.Name())
	return "." + string(message.ParentFile().Package()) + "." + name
}

func (g *JSONSchemaGenerator) schemaOrReferenceForType(typeName string) *jsonschema.Schema {
	switch typeName {

	case ".google.protobuf.Timestamp":
		// Timestamps are serialized as strings
		typ := "string"
		format := "date-time"
		return &jsonschema.Schema{Type: &jsonschema.StringOrStringArray{String: &typ}, Format: &format}

	case ".google.type.Date":
		// Dates are serialized as strings
		typ := "string"
		format := "date"
		return &jsonschema.Schema{Type: &jsonschema.StringOrStringArray{String: &typ}, Format: &format}

	case ".google.type.DateTime":
		// DateTimes are serialized as strings
		typ := "string"
		format := "date-time"
		return &jsonschema.Schema{Type: &jsonschema.StringOrStringArray{String: &typ}, Format: &format}

	case ".google.protobuf.Struct":
		// Struct is equivalent to a JSON object
		typ := "object"
		return &jsonschema.Schema{Type: &jsonschema.StringOrStringArray{String: &typ}}

	case ".google.protobuf.Empty":
		// Empty is close to JSON undefined than null, so ignore this field
		return nil
	}

	ref := g.schemaReferenceForTypeName(typeName)
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
		typeName := fullMessageTypeName(field.Message())

		kindSchema = g.schemaOrReferenceForType(typeName)
		if kindSchema == nil {
			return nil
		}

		if kindSchema.Ref != nil && !refInDefinitions(*kindSchema.Ref, definitions) {
			ref := strings.Replace(*kindSchema.Ref, "#/definitions/", *g.conf.BaseURL, 1)
			ref += ".json"
			kindSchema.Ref = &ref
		}

	case protoreflect.StringKind:
		typ := "string"
		kindSchema = &jsonschema.Schema{Type: &jsonschema.StringOrStringArray{String: &typ}}

	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Uint32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Uint64Kind,
		protoreflect.Sfixed32Kind, protoreflect.Fixed32Kind, protoreflect.Sfixed64Kind,
		protoreflect.Fixed64Kind:
		typ := "integer"
		format := kind.String()
		kindSchema = &jsonschema.Schema{Type: &jsonschema.StringOrStringArray{String: &typ}, Format: &format}

	case protoreflect.EnumKind:
		typ := "integer"
		format := "enum"
		kindSchema = &jsonschema.Schema{Type: &jsonschema.StringOrStringArray{String: &typ}, Format: &format}

	case protoreflect.BoolKind:
		typ := "boolean"
		kindSchema = &jsonschema.Schema{Type: &jsonschema.StringOrStringArray{String: &typ}}

	case protoreflect.FloatKind, protoreflect.DoubleKind:
		typ := "number"
		format := kind.String()
		kindSchema = &jsonschema.Schema{Type: &jsonschema.StringOrStringArray{String: &typ}, Format: &format}

	case protoreflect.BytesKind:
		typ := "string"
		format := "bytes"
		kindSchema = &jsonschema.Schema{Type: &jsonschema.StringOrStringArray{String: &typ}, Format: &format}

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
			subSchemas := g.buildSchemasFromMessages(message.Messages)
			for i, subSchema := range subSchemas {
				idURL, _ := url.Parse(*subSchema.Value.ID)
				path := strings.TrimSuffix(idURL.Path, ".json")
				subSchemas[i].Value.ID = &path
			}
			schema.Value.Definitions = &subSchemas
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
			if strings.TrimSuffix(*schema.Value.Schema, "#") == "http://json-schema.org/draft-07/schema" {
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
