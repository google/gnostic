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
	"regexp"
	"sort"
	"strings"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	v3 "github.com/google/gnostic/openapiv3"
)

type Configuration struct {
	Version     *string
	Title       *string
	Description *string
	Naming      *string
}

const infoURL = "https://github.com/google/gnostic/tree/master/apps/protoc-gen-openapi"

// OpenAPIv3Generator holds internal state needed to generate an OpenAPIv3 document for a transcoded Protocol Buffer service.
type OpenAPIv3Generator struct {
	conf   Configuration
	plugin *protogen.Plugin

	singleService     bool     // 1 file with 1 service
	requiredSchemas   []string // Names of schemas that need to be generated.
	generatedSchemas  []string // Names of schemas that have already been generated.
	linterRulePattern *regexp.Regexp
	pathPattern       *regexp.Regexp
	namedPathPattern  *regexp.Regexp
}

// NewOpenAPIv3Generator creates a new generator for a protoc plugin invocation.
func NewOpenAPIv3Generator(plugin *protogen.Plugin, conf Configuration) *OpenAPIv3Generator {
	return &OpenAPIv3Generator{
		conf:   conf,
		plugin: plugin,

		requiredSchemas:   make([]string, 0),
		generatedSchemas:  make([]string, 0),
		linterRulePattern: regexp.MustCompile(`\(-- .* --\)`),
		pathPattern:       regexp.MustCompile("{([^=}]+)}"),
		namedPathPattern:  regexp.MustCompile("{(.+)=(.+)}"),
	}
}

// Run runs the generator.
func (g *OpenAPIv3Generator) Run() error {
	d := g.buildDocumentV3()
	bytes, err := d.YAMLValue("Generated with protoc-gen-openapi\n" + infoURL)
	if err != nil {
		return fmt.Errorf("failed to marshal yaml: %s", err.Error())
	}
	outputFile := g.plugin.NewGeneratedFile("openapi.yaml", "")
	outputFile.Write(bytes)
	return nil
}

// buildDocumentV3 builds an OpenAPIv3 document for a plugin request.
func (g *OpenAPIv3Generator) buildDocumentV3() *v3.Document {
	d := &v3.Document{}

	d.Openapi = "3.0.3"
	d.Info = &v3.Info{
		Version:     *g.conf.Version,
		Title:       *g.conf.Title,
		Description: *g.conf.Description,
	}

	d.Paths = &v3.Paths{}
	d.Components = &v3.Components{
		Schemas: &v3.SchemasOrReferences{
			AdditionalProperties: []*v3.NamedSchemaOrReference{},
		},
	}

	for _, file := range g.plugin.Files {
		if file.Generate {
			g.addPathsToDocumentV3(d, file)
		}
	}

	// If there is only 1 service, then use it's title for the document,
	//  if the document is missing it.
	if len(d.Tags) == 1 {
		if d.Info.Title == "" && d.Tags[0].Name != "" {
			d.Info.Title = d.Tags[0].Name + " API"
		}
		if d.Info.Description == "" {
			d.Info.Description = d.Tags[0].Description
		}
		d.Tags[0].Description = ""
	}

	for len(g.requiredSchemas) > 0 {
		count := len(g.requiredSchemas)
		for _, file := range g.plugin.Files {
			g.addSchemasToDocumentV3(d, file)
		}
		g.requiredSchemas = g.requiredSchemas[count:len(g.requiredSchemas)]
	}
	// Sort the tags.
	{
		pairs := d.Tags
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].Name < pairs[j].Name
		})
		d.Tags = pairs
	}
	// Sort the paths.
	{
		pairs := d.Paths.Path
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].Name < pairs[j].Name
		})
		d.Paths.Path = pairs
	}
	// Sort the schemas.
	{
		pairs := d.Components.Schemas.AdditionalProperties
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].Name < pairs[j].Name
		})
		d.Components.Schemas.AdditionalProperties = pairs
	}
	return d
}

// filterCommentString removes line breaks and linter rules from comments.
func (g *OpenAPIv3Generator) filterCommentString(c protogen.Comments, removeNewLines bool) string {
	comment := string(c)
	if removeNewLines {
		comment = strings.Replace(comment, "\n", "", -1)
	}
	comment = g.linterRulePattern.ReplaceAllString(comment, "")
	return strings.TrimSpace(comment)
}

// addPathsToDocumentV3 adds paths from a specified file descriptor.
func (g *OpenAPIv3Generator) addPathsToDocumentV3(d *v3.Document, file *protogen.File) {
	for _, service := range file.Services {
		comment := g.filterCommentString(service.Comments.Leading, false)
		d.Tags = append(d.Tags, &v3.Tag{Name: service.GoName, Description: comment})

		for _, method := range service.Methods {
			comment := g.filterCommentString(method.Comments.Leading, false)
			inputMessage := method.Input
			outputMessage := method.Output
			operationID := service.GoName + "_" + method.GoName
			xt := annotations.E_Http
			extension := proto.GetExtension(method.Desc.Options(), xt)
			var path string
			var methodName string
			var body string
			if extension != nil && extension != xt.InterfaceOf(xt.Zero()) {
				rule := extension.(*annotations.HttpRule)
				body = rule.Body
				switch pattern := rule.Pattern.(type) {
				case *annotations.HttpRule_Get:
					path = pattern.Get
					methodName = "GET"
				case *annotations.HttpRule_Post:
					path = pattern.Post
					methodName = "POST"
				case *annotations.HttpRule_Put:
					path = pattern.Put
					methodName = "PUT"
				case *annotations.HttpRule_Delete:
					path = pattern.Delete
					methodName = "DELETE"
				case *annotations.HttpRule_Patch:
					path = pattern.Patch
					methodName = "PATCH"
				case *annotations.HttpRule_Custom:
					path = "custom-unsupported"
				default:
					path = "unknown-unsupported"
				}
			}
			if methodName != "" {
				op, path2 := g.buildOperationV3(
					file, operationID, service.GoName, comment, path, body, inputMessage, outputMessage)
				g.addOperationV3(d, op, path2, methodName)
			}
		}
	}
}

func (g *OpenAPIv3Generator) formatMessageRef(name string) string {
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

func (g *OpenAPIv3Generator) formatMessageName(message *protogen.Message) string {
	if *g.conf.Naming == "proto" {
		return string(message.Desc.Name())
	}

	name := string(message.Desc.Name())
	if len(name) > 0 {
		return strings.ToUpper(name[0:1]) + name[1:]
	}

	return name
}

func (g *OpenAPIv3Generator) formatFieldName(field *protogen.Field) string {
	if *g.conf.Naming == "proto" {
		return string(field.Desc.Name())
	}

	return field.Desc.JSONName()
}

func (g *OpenAPIv3Generator) findAndFormatFieldName(name string, inMessage *protogen.Message) string {
	for _, field := range inMessage.Fields {
		if string(field.Desc.Name()) == name {
			return g.formatFieldName(field)
		}
	}

	return name
}

// buildOperationV3 constructs an operation for a set of values.
func (g *OpenAPIv3Generator) buildOperationV3(
	file *protogen.File,
	operationID string,
	tagName string,
	description string,
	path string,
	bodyField string,
	inputMessage *protogen.Message,
	outputMessage *protogen.Message,
) (*v3.Operation, string) {
	// coveredParameters tracks the parameters that have been used in the body or path.
	coveredParameters := make([]string, 0)
	if bodyField != "" {
		coveredParameters = append(coveredParameters, bodyField)
	}
	// Initialize the list of operation parameters.
	parameters := []*v3.ParameterOrReference{}

	// Build a list of path parameters.
	pathParameters := make([]string, 0)
	// Find simple path parameters like {id}
	if allMatches := g.pathPattern.FindAllStringSubmatch(path, -1); allMatches != nil {
		for _, matches := range allMatches {
			// Add the value to the list of covered parameters.
			coveredParameters = append(coveredParameters, matches[1])
			pathParameter := g.findAndFormatFieldName(matches[1], inputMessage)
			path = strings.Replace(path, matches[1], pathParameter, 1)
			pathParameters = append(pathParameters, pathParameter)
		}
	}

	// Add the path parameters to the operation parameters.
	for _, pathParameter := range pathParameters {
		parameters = append(parameters,
			&v3.ParameterOrReference{
				Oneof: &v3.ParameterOrReference_Parameter{
					Parameter: &v3.Parameter{
						Name:     pathParameter,
						In:       "path",
						Required: true,
						Schema: &v3.SchemaOrReference{
							Oneof: &v3.SchemaOrReference_Schema{
								Schema: &v3.Schema{
									Type: "string",
								},
							},
						},
					},
				},
			})
	}

	// Build a list of named path parameters.
	namedPathParameters := make([]string, 0)
	// Find named path parameters like {name=shelves/*}
	if matches := g.namedPathPattern.FindStringSubmatch(path); matches != nil {
		// Add the "name=" "name" value to the list of covered parameters.
		coveredParameters = append(coveredParameters, matches[1])
		// Convert the path from the starred form to use named path parameters.
		starredPath := matches[2]
		parts := strings.Split(starredPath, "/")
		// The starred path is assumed to be in the form "things/*/otherthings/*".
		// We want to convert it to "things/{thingsId}/otherthings/{otherthingsId}".
		for i := 0; i < len(parts)-1; i += 2 {
			section := parts[i]
			namedPathParameter := g.findAndFormatFieldName(section, inputMessage)
			namedPathParameter = singular(namedPathParameter)
			parts[i+1] = "{" + namedPathParameter + "}"
			namedPathParameters = append(namedPathParameters, namedPathParameter)
		}
		// Rewrite the path to use the path parameters.
		newPath := strings.Join(parts, "/")
		path = strings.Replace(path, matches[0], newPath, 1)
	}

	// Add the named path parameters to the operation parameters.
	for _, namedPathParameter := range namedPathParameters {
		parameters = append(parameters,
			&v3.ParameterOrReference{
				Oneof: &v3.ParameterOrReference_Parameter{
					Parameter: &v3.Parameter{
						Name:        namedPathParameter,
						In:          "path",
						Required:    true,
						Description: "The " + namedPathParameter + " id.",
						Schema: &v3.SchemaOrReference{
							Oneof: &v3.SchemaOrReference_Schema{
								Schema: &v3.Schema{
									Type: "string",
								},
							},
						},
					},
				},
			})
	}
	// Add any unhandled fields in the request message as query parameters.
	if bodyField != "*" {
		for _, field := range inputMessage.Fields {
			fieldName := string(field.Desc.Name())
			if !contains(coveredParameters, fieldName) {
				bodyFieldName := g.formatFieldName(field)
				// Get the field description from the comments.
				fieldDescription := g.filterCommentString(field.Comments.Leading, true)
				parameters = append(parameters,
					&v3.ParameterOrReference{
						Oneof: &v3.ParameterOrReference_Parameter{
							Parameter: &v3.Parameter{
								Name:        bodyFieldName,
								In:          "query",
								Description: fieldDescription,
								Required:    false,
								Schema: &v3.SchemaOrReference{
									Oneof: &v3.SchemaOrReference_Schema{
										Schema: &v3.Schema{
											Type: "string",
										},
									},
								},
							},
						},
					})
			}
		}
	}
	// Create the response.
	responses := &v3.Responses{
		ResponseOrReference: []*v3.NamedResponseOrReference{
			{
				Name: "200",
				Value: &v3.ResponseOrReference{
					Oneof: &v3.ResponseOrReference_Response{
						Response: &v3.Response{
							Description: "OK",
							Content:     g.responseContentForMessage(outputMessage),
						},
					},
				},
			},
		},
	}
	// Create the operation.
	op := &v3.Operation{
		Tags:        []string{tagName},
		Description: description,
		OperationId: operationID,
		Parameters:  parameters,
		Responses:   responses,
	}
	// If a body field is specified, we need to pass a message as the request body.
	if bodyField != "" {
		var bodyFieldScalarTypeName string
		var bodyFieldMessageTypeName string
		if bodyField == "*" {
			// Pass the entire request message as the request body.
			bodyFieldMessageTypeName = fullMessageTypeName(inputMessage)
		} else {
			// If body refers to a message field, use that type.
			for _, field := range inputMessage.Fields {
				if string(field.Desc.Name()) == bodyField {
					switch field.Desc.Kind() {
					case protoreflect.StringKind:
						bodyFieldScalarTypeName = "string"
					case protoreflect.MessageKind:
						bodyFieldMessageTypeName = fullMessageTypeName(field.Message)
					default:
						log.Printf("unsupported field type %+v", field.Desc)
					}
					break
				}
			}
		}
		var requestSchema *v3.SchemaOrReference
		if bodyFieldScalarTypeName != "" {
			requestSchema = &v3.SchemaOrReference{
				Oneof: &v3.SchemaOrReference_Schema{
					Schema: &v3.Schema{
						Type: bodyFieldScalarTypeName,
					},
				},
			}
		} else if bodyFieldMessageTypeName != "" {
			switch bodyFieldMessageTypeName {
			case ".google.protobuf.Empty":
				fallthrough
			case ".google.protobuf.Struct":
				requestSchema = &v3.SchemaOrReference{
					Oneof: &v3.SchemaOrReference_Schema{
						Schema: &v3.Schema{
							Type: "object",
						},
					},
				}
			default:
				requestSchema = &v3.SchemaOrReference{
					Oneof: &v3.SchemaOrReference_Reference{
						Reference: &v3.Reference{
							XRef: g.schemaReferenceForTypeName(bodyFieldMessageTypeName),
						}},
				}
			}
		}
		op.RequestBody = &v3.RequestBodyOrReference{
			Oneof: &v3.RequestBodyOrReference_RequestBody{
				RequestBody: &v3.RequestBody{
					Required: true,
					Content: &v3.MediaTypes{
						AdditionalProperties: []*v3.NamedMediaType{
							{
								Name: "application/json",
								Value: &v3.MediaType{
									Schema: requestSchema,
								},
							},
						},
					},
				},
			},
		}
	}
	return op, path
}

// addOperationV3 adds an operation to the specified path/method.
func (g *OpenAPIv3Generator) addOperationV3(d *v3.Document, op *v3.Operation, path string, methodName string) {
	var selectedPathItem *v3.NamedPathItem
	for _, namedPathItem := range d.Paths.Path {
		if namedPathItem.Name == path {
			selectedPathItem = namedPathItem
			break
		}
	}
	// If we get here, we need to create a path item.
	if selectedPathItem == nil {
		selectedPathItem = &v3.NamedPathItem{Name: path, Value: &v3.PathItem{}}
		d.Paths.Path = append(d.Paths.Path, selectedPathItem)
	}
	// Set the operation on the specified method.
	switch methodName {
	case "GET":
		selectedPathItem.Value.Get = op
	case "POST":
		selectedPathItem.Value.Post = op
	case "PUT":
		selectedPathItem.Value.Put = op
	case "DELETE":
		selectedPathItem.Value.Delete = op
	case "PATCH":
		selectedPathItem.Value.Patch = op
	}
}

// schemaReferenceForTypeName returns an OpenAPI JSON Reference to the schema that represents a type.
func (g *OpenAPIv3Generator) schemaReferenceForTypeName(typeName string) string {
	if !contains(g.requiredSchemas, typeName) {
		g.requiredSchemas = append(g.requiredSchemas, typeName)
	}
	parts := strings.Split(typeName, ".")
	lastPart := parts[len(parts)-1]
	return "#/components/schemas/" + g.formatMessageRef(lastPart)
}

// itemsItemForTypeName is a helper constructor.
func itemsItemForTypeName(typeName string) *v3.ItemsItem {
	return &v3.ItemsItem{SchemaOrReference: []*v3.SchemaOrReference{{
		Oneof: &v3.SchemaOrReference_Schema{
			Schema: &v3.Schema{
				Type: typeName}}}}}
}

// itemsItemForReference is a helper constructor.
func itemsItemForReference(xref string) *v3.ItemsItem {
	return &v3.ItemsItem{SchemaOrReference: []*v3.SchemaOrReference{{
		Oneof: &v3.SchemaOrReference_Reference{
			Reference: &v3.Reference{
				XRef: xref}}}}}
}

// fullMessageTypeName builds the full type name of a message.
func fullMessageTypeName(message *protogen.Message) string {
	return "." + string(message.Desc.ParentFile().Package()) + "." + string(message.Desc.Name())
}

func (g *OpenAPIv3Generator) responseContentForMessage(outputMessage *protogen.Message) *v3.MediaTypes {
	typeName := fullMessageTypeName(outputMessage)

	if typeName == ".google.protobuf.Empty" {
		return &v3.MediaTypes{}
	}
	if typeName == ".google.protobuf.Struct" {
		return &v3.MediaTypes{}
	}

	if typeName == ".google.api.HttpBody" {
		return &v3.MediaTypes{
			AdditionalProperties: []*v3.NamedMediaType{
				{
					Name:  "application/octet-stream",
					Value: &v3.MediaType{},
				},
			},
		}
	}

	return &v3.MediaTypes{
		AdditionalProperties: []*v3.NamedMediaType{
			{
				Name: "application/json",
				Value: &v3.MediaType{
					Schema: &v3.SchemaOrReference{
						Oneof: &v3.SchemaOrReference_Reference{
							Reference: &v3.Reference{
								XRef: g.schemaReferenceForTypeName(fullMessageTypeName(outputMessage)),
							},
						},
					},
				},
			},
		},
	}
}

// addSchemasToDocumentV3 adds info from one file descriptor.
func (g *OpenAPIv3Generator) addSchemasToDocumentV3(d *v3.Document, file *protogen.File) {
	// For each message, generate a definition.
	for _, message := range file.Messages {
		typeName := fullMessageTypeName(message)
		// Only generate this if we need it and haven't already generated it.
		if !contains(g.requiredSchemas, typeName) ||
			contains(g.generatedSchemas, typeName) {
			continue
		}
		g.generatedSchemas = append(g.generatedSchemas, typeName)
		// Get the message description from the comments.
		messageDescription := g.filterCommentString(message.Comments.Leading, true)
		// Build an array holding the fields of the message.
		definitionProperties := &v3.Properties{
			AdditionalProperties: make([]*v3.NamedSchemaOrReference, 0),
		}
		for _, field := range message.Fields {
			// Check the field annotations to see if this is a readonly field.
			outputOnly := false
			extension := proto.GetExtension(field.Desc.Options(), annotations.E_FieldBehavior)
			if extension != nil {
				switch v := extension.(type) {
				case []annotations.FieldBehavior:
					for _, vv := range v {
						if vv == annotations.FieldBehavior_OUTPUT_ONLY {
							outputOnly = true
						}
					}
				default:
					log.Printf("unsupported extension type %T", extension)
				}
			}
			// Get the field description from the comments.
			fieldDescription := g.filterCommentString(field.Comments.Leading, true)
			// The field is either described by a reference or a schema.
			XRef := ""
			fieldSchema := &v3.Schema{
				Description: fieldDescription,
			}
			if outputOnly {
				fieldSchema.ReadOnly = true
			}
			if field.Desc.IsList() {
				fieldSchema.Type = "array"
				switch field.Desc.Kind() {
				case protoreflect.MessageKind:
					typeName := fullMessageTypeName(field.Message)
					switch typeName {
					case ".google.protobuf.Timestamp":
						// Timestamps are serialized as strings
						fieldSchema.Items = itemsItemForTypeName("string")
					case ".google.type.Date":
						// Dates are serialized as strings
						fieldSchema.Items = itemsItemForTypeName("string")
					case ".google.type.DateTime":
						// DateTimes are serialized as strings
						fieldSchema.Items = itemsItemForTypeName("string")
					case ".google.protobuf.Struct":
						// Struct is equivalent to a JSON object
						fieldSchema.Items = itemsItemForTypeName("object")
					case ".google.protobuf.Empty":
						// Struct is close to JSON null, so ignore this field
						continue
					default:
						// The field is described by a reference.
						fieldSchema.Items = itemsItemForReference(
							g.schemaReferenceForTypeName(typeName))
					}
				case protoreflect.StringKind:
					fieldSchema.Items = itemsItemForTypeName("string")
				case protoreflect.Int32Kind,
					protoreflect.Sint32Kind,
					protoreflect.Uint32Kind,
					protoreflect.Int64Kind,
					protoreflect.Sint64Kind,
					protoreflect.Uint64Kind,
					protoreflect.Sfixed32Kind,
					protoreflect.Fixed32Kind,
					protoreflect.Sfixed64Kind,
					protoreflect.Fixed64Kind:
					fieldSchema.Items = itemsItemForTypeName("integer")
				case protoreflect.EnumKind:
					fieldSchema.Items = itemsItemForTypeName("integer")
				case protoreflect.BoolKind:
					fieldSchema.Items = itemsItemForTypeName("boolean")
				case protoreflect.FloatKind, protoreflect.DoubleKind:
					fieldSchema.Items = itemsItemForTypeName("number")
				case protoreflect.BytesKind:
					fieldSchema.Items = itemsItemForTypeName("string")
				default:
					log.Printf("(TODO) Unsupported array type: %+v", fullMessageTypeName(field.Message))
				}
			} else if field.Desc.IsMap() &&
				field.Desc.MapKey().Kind() == protoreflect.StringKind &&
				field.Desc.MapValue().Kind() == protoreflect.StringKind {
				fieldSchema.Type = "object"
			} else {
				k := field.Desc.Kind()
				switch k {
				case protoreflect.MessageKind:
					typeName := fullMessageTypeName(field.Message)
					switch typeName {
					case ".google.protobuf.Timestamp":
						// Timestamps are serialized as strings
						fieldSchema.Type = "string"
						fieldSchema.Format = "RFC3339"
					case ".google.type.Date":
						// Dates are serialized as strings
						fieldSchema.Type = "string"
						fieldSchema.Format = "date"
					case ".google.type.DateTime":
						// DateTimes are serialized as strings
						fieldSchema.Type = "string"
						fieldSchema.Format = "date-time"
					case ".google.protobuf.Struct":
						// Struct is equivalent to a JSON object
						fieldSchema.Type = "object"
					case ".google.protobuf.Empty":
						// Struct is close to JSON null, so ignore this field
						continue
					default:
						// The field is described by a reference.
						XRef = g.schemaReferenceForTypeName(typeName)
					}
				case protoreflect.StringKind:
					fieldSchema.Type = "string"
				case protoreflect.Int32Kind,
					protoreflect.Sint32Kind,
					protoreflect.Uint32Kind,
					protoreflect.Int64Kind,
					protoreflect.Sint64Kind,
					protoreflect.Uint64Kind,
					protoreflect.Sfixed32Kind,
					protoreflect.Fixed32Kind,
					protoreflect.Sfixed64Kind,
					protoreflect.Fixed64Kind:
					fieldSchema.Type = "integer"
					fieldSchema.Format = k.String()
				case protoreflect.EnumKind:
					fieldSchema.Type = "integer"
					fieldSchema.Format = "enum"
				case protoreflect.BoolKind:
					fieldSchema.Type = "boolean"
				case protoreflect.FloatKind, protoreflect.DoubleKind:
					fieldSchema.Type = "number"
					fieldSchema.Format = k.String()
				case protoreflect.BytesKind:
					fieldSchema.Type = "string"
					fieldSchema.Format = "bytes"
				default:
					log.Printf("(TODO) Unsupported field type: %+v", fullMessageTypeName(field.Message))
				}
			}
			var value *v3.SchemaOrReference
			if XRef != "" {
				value = &v3.SchemaOrReference{
					Oneof: &v3.SchemaOrReference_Reference{
						Reference: &v3.Reference{
							XRef: XRef,
						},
					},
				}
			} else {
				value = &v3.SchemaOrReference{
					Oneof: &v3.SchemaOrReference_Schema{
						Schema: fieldSchema,
					},
				}
			}
			definitionProperties.AdditionalProperties = append(
				definitionProperties.AdditionalProperties,
				&v3.NamedSchemaOrReference{
					Name:  g.formatFieldName(field),
					Value: value,
				},
			)
		}
		// Add the schema to the components.schema list.
		d.Components.Schemas.AdditionalProperties = append(d.Components.Schemas.AdditionalProperties,
			&v3.NamedSchemaOrReference{
				Name: g.formatMessageName(message),
				Value: &v3.SchemaOrReference{
					Oneof: &v3.SchemaOrReference_Schema{
						Schema: &v3.Schema{
							Description: messageDescription,
							Properties:  definitionProperties,
						},
					},
				},
			},
		)
	}
}

// contains returns true if an array contains a specified string.
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// singular produces the singular form of a collection name.
func singular(plural string) string {
	if strings.HasSuffix(plural, "ves") {
		return strings.TrimSuffix(plural, "ves") + "f"
	}
	if strings.HasSuffix(plural, "ies") {
		return strings.TrimSuffix(plural, "ies") + "y"
	}
	if strings.HasSuffix(plural, "s") {
		return strings.TrimSuffix(plural, "s")
	}
	return plural
}
