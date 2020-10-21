// Copyright 2017 Google Inc. All Rights Reserved.
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

package generator

import (
	"log"
	"regexp"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	v3 "github.com/googleapis/gnostic/openapiv3"
	"google.golang.org/genproto/googleapis/api/annotations"
)

func (g *Generator) definitionReferenceForTypeName(typeName string) string {
	parts := strings.Split(typeName, ".")
	lastPart := parts[len(parts)-1]
	return "#/definitions/" + lastPart
}

// NewDocumentV3 creates a new OpenAPIv3 document
func (g *Generator) NewDocumentV3() *v3.Document {
	d := &v3.Document{}
	d.Openapi = "3.0"
	d.Info = &v3.Info{
		Title:   "OpenAPI Petstore",
		Version: "1.0.0",
		License: &v3.License{Name: "MIT"},
	}
	d.Servers = append(d.Servers, &v3.Server{
		Url:         "https://petstore.openapis.org/v1",
		Description: "Development server",
	})
	d.Paths = &v3.Paths{}
	d.Components = &v3.Components{
		Schemas: &v3.SchemasOrReferences{
			AdditionalProperties: []*v3.NamedSchemaOrReference{},
		},
	}
	return d
}

func itemsItemForType(typeName string) *v3.ItemsItem {
	return &v3.ItemsItem{SchemaOrReference: []*v3.SchemaOrReference{&v3.SchemaOrReference{
		Oneof: &v3.SchemaOrReference_Schema{
			Schema: &v3.Schema{
				Type: typeName}}}}}
}

func itemsItemForReference(xref string) *v3.ItemsItem {
	return &v3.ItemsItem{SchemaOrReference: []*v3.SchemaOrReference{&v3.SchemaOrReference{
		Oneof: &v3.SchemaOrReference_Reference{
			Reference: &v3.Reference{
				XRef: xref}}}}}
}

// AddToDocumentV3 adds info from one file descriptor
func (g *Generator) AddToDocumentV3(d *v3.Document, file *FileDescriptor) {
	g.file = file
	sourceCodeInfo := file.SourceCodeInfo
	if false {
		for _, location := range sourceCodeInfo.Location {
			log.Printf("%+v", location)
		}
	}
	linterRulePattern := regexp.MustCompile(`\(-- .* --\)`)

	for _, service := range file.FileDescriptorProto.Service {
		log.Printf("SERVICE %s", *service.Name)
		for i, method := range service.Method {
			log.Printf("METHOD %d: %s", i, *method.Name)

			// search source info for leading comments
			var comment string
			for _, location := range sourceCodeInfo.Location {
				paths := location.GetPath()
				if len(paths) == 4 &&
					paths[0] == 6 &&
					paths[1] == 0 &&
					paths[2] == 2 &&
					paths[3] == int32(i) {
					comment = location.GetLeadingComments()
					comment = strings.Replace(comment, "\n", "", -1)
					comment = linterRulePattern.ReplaceAllString(comment, "")
					comment = strings.TrimSpace(comment)
				}
			}

			inputType := *method.InputType
			log.Printf(" INPUT %s", inputType)

			inputMessage := g.FindMessage(inputType)
			log.Printf(" INPUT MESSAGE %+v", inputMessage)

			outputType := *method.OutputType
			log.Printf(" OUTPUT %s", outputType)

			outputMessage := g.FindMessage(outputType)
			log.Printf(" OUTPUT MESSAGE %+v", outputMessage)

			log.Printf(" OPTIONS %+v", *method.Options)

			operationID := *service.Name + "_" + *method.Name

			extension, err := proto.GetExtension(method.Options, annotations.E_Http)
			log.Printf(" extensions: %T %+v (%+v)", extension, extension, err)
			var path string
			var methodName string
			var body string
			if extension != nil {
				rule := extension.(*annotations.HttpRule)
				log.Printf("  PATTERN %T %v", rule.Pattern, rule.Pattern)
				log.Printf("  SELECTOR %s", rule.Selector)
				log.Printf("  BODY %s", rule.Body)
				log.Printf("  BINDINGS %s", rule.AdditionalBindings)
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
				log.Printf("  PATH %s", path)
			}
			log.Printf("%s", methodName)
			if methodName != "" {
				op, path2 := g.buildOperationV3(operationID, comment, path, body, inputMessage, outputMessage)
				g.addOperationV3(d, op, path2, methodName)
			}
		}
	}

	// for each message, generate a definition
	for _, desc := range g.file.desc {
		definitionProperties := &v3.Properties{
			AdditionalProperties: make([]*v3.NamedSchemaOrReference, 0),
		}
		for _, field := range desc.Field {
			XRef := ""
			fieldSchema := &v3.Schema{}
			if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
				fieldSchema.Type = "array"
				switch *field.Type {
				case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
					fieldSchema.Items = itemsItemForReference(g.definitionReferenceForTypeName(*field.TypeName))
				case descriptor.FieldDescriptorProto_TYPE_STRING:
					fieldSchema.Items = itemsItemForType("string")
				case descriptor.FieldDescriptorProto_TYPE_INT32:
					fieldSchema.Items = itemsItemForType("integer")
				case descriptor.FieldDescriptorProto_TYPE_UINT64:
					fieldSchema.Items = itemsItemForType("integer")
				case descriptor.FieldDescriptorProto_TYPE_ENUM:
					fieldSchema.Items = itemsItemForType("integer")
				default:
					log.Printf("(TODO) Unsupported array type: %+v", field.Type)
				}
			} else {
				switch *field.Type {
				case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
					XRef = g.definitionReferenceForTypeName(*field.TypeName)
				case descriptor.FieldDescriptorProto_TYPE_STRING:
					fieldSchema.Type = "string"
				case descriptor.FieldDescriptorProto_TYPE_INT64:
					fieldSchema.Type = "integer"
					fieldSchema.Format = "int64"
				case descriptor.FieldDescriptorProto_TYPE_UINT64:
					fieldSchema.Type = "integer"
					fieldSchema.Format = "uint64"
				case descriptor.FieldDescriptorProto_TYPE_INT32:
					fieldSchema.Type = "integer"
					fieldSchema.Format = "int32"
				case descriptor.FieldDescriptorProto_TYPE_ENUM:
					fieldSchema.Type = "integer"
					fieldSchema.Format = "enum"
				case descriptor.FieldDescriptorProto_TYPE_BOOL:
					fieldSchema.Type = "boolean"
				case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
					fieldSchema.Type = "double"
				case descriptor.FieldDescriptorProto_TYPE_BYTES:
					fieldSchema.Type = "byte"
				default:
					log.Printf("(TODO) Unsupported field type: %+v", field.Type)
				}
			}
			if XRef != "" {
				definitionProperties.AdditionalProperties = append(
					definitionProperties.AdditionalProperties,
					&v3.NamedSchemaOrReference{
						Name: *field.Name,
						Value: &v3.SchemaOrReference{
							Oneof: &v3.SchemaOrReference_Reference{
								Reference: &v3.Reference{XRef: XRef},
							},
						},
					},
				)
			} else {
				definitionProperties.AdditionalProperties = append(
					definitionProperties.AdditionalProperties,
					&v3.NamedSchemaOrReference{
						Name: *field.Name,
						Value: &v3.SchemaOrReference{
							Oneof: &v3.SchemaOrReference_Schema{
								Schema: fieldSchema,
							},
						},
					},
				)
			}
		}
		d.Components.Schemas.AdditionalProperties = append(d.Components.Schemas.AdditionalProperties,
			&v3.NamedSchemaOrReference{
				Name: *desc.Name,
				Value: &v3.SchemaOrReference{
					Oneof: &v3.SchemaOrReference_Schema{
						Schema: &v3.Schema{Properties: definitionProperties},
					},
				},
			})
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func singular(plural string) string {
	if strings.HasSuffix(plural, "s") {
		return strings.TrimSuffix(plural, "s")
	}
	return plural
}

func (g *Generator) buildOperationV3(
	operationID string,
	description string,
	path string,
	bodyField string,
	inputMessage *Descriptor,
	outputMessage *Descriptor,
) (*v3.Operation, string) {

	namePattern := regexp.MustCompile("{(.*)=(.*)}")

	coveredParameters := make([]string, 0)
	if bodyField != "" {
		coveredParameters = append(coveredParameters, bodyField)
	}

	pathParameters := make([]string, 0)

	if matches := namePattern.FindStringSubmatch(path); matches != nil {
		log.Printf("MATCHES %+v", matches)
		coveredParameters = append(coveredParameters, matches[1])
		starredPath := matches[2]

		parts := strings.Split(starredPath, "/")
		log.Printf("parts: %+v", parts)

		for i := 0; i < len(parts); i += 2 {
			section := parts[i]
			parameter := singular(section)
			parts[i+1] = "{" + parameter + "}"
			pathParameters = append(pathParameters, parameter)
		}

		newPath := strings.Join(parts, "/")
		log.Printf("new: %s", newPath)

		path = strings.Replace(path, matches[0], newPath, 1)
		log.Printf("path %s", path)
	}

	parameters := []*v3.ParameterOrReference{}

	for _, pathParameter := range pathParameters {
		parameters = append(parameters,
			&v3.ParameterOrReference{
				Oneof: &v3.ParameterOrReference_Parameter{
					Parameter: &v3.Parameter{
						Name:        pathParameter,
						In:          "path",
						Required:    true,
						Description: pathParameter + " identifier",
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

	for _, field := range inputMessage.Field {
		fieldName := *field.Name

		if !contains(coveredParameters, fieldName) {
			parameters = append(parameters,
				&v3.ParameterOrReference{
					Oneof: &v3.ParameterOrReference_Parameter{
						Parameter: &v3.Parameter{
							Name:        fieldName,
							In:          "query",
							Description: "",
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

	op := &v3.Operation{
		Summary:     description,
		OperationId: operationID,
		Parameters:  parameters,
		Responses: &v3.Responses{
			ResponseOrReference: []*v3.NamedResponseOrReference{
				&v3.NamedResponseOrReference{
					Name: "200",
					Value: &v3.ResponseOrReference{
						Oneof: &v3.ResponseOrReference_Response{
							Response: &v3.Response{
								Description: "OK",
								Content: &v3.MediaTypes{
									AdditionalProperties: []*v3.NamedMediaType{
										&v3.NamedMediaType{
											Name: "application/json",
											Value: &v3.MediaType{
												Schema: &v3.SchemaOrReference{
													Oneof: &v3.SchemaOrReference_Reference{
														Reference: &v3.Reference{
															XRef: "#/components/schemas/" + *outputMessage.Name,
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if bodyField != "" {
		bodyFieldTypeName := ""
		for _, field := range inputMessage.Field {
			if *field.Name == bodyField {
				bodyFieldTypeName = *field.TypeName
				break
			}
		}
		op.RequestBody = &v3.RequestBodyOrReference{
			Oneof: &v3.RequestBodyOrReference_RequestBody{
				RequestBody: &v3.RequestBody{
					Required: true,
					Content: &v3.MediaTypes{
						AdditionalProperties: []*v3.NamedMediaType{
							&v3.NamedMediaType{
								Name: "application/json",
								Value: &v3.MediaType{
									Schema: &v3.SchemaOrReference{
										Oneof: &v3.SchemaOrReference_Reference{
											Reference: &v3.Reference{
												XRef: g.definitionReferenceForTypeName(bodyFieldTypeName),
											},
										},
									},
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

func (g *Generator) addOperationV3(d *v3.Document, op *v3.Operation, path string, methodName string) {
	for _, namedPathItem := range d.Paths.Path {
		if namedPathItem.Name == path {
			switch methodName {
			case "GET":
				namedPathItem.Value.Get = op
			case "POST":
				namedPathItem.Value.Post = op
			case "PUT":
				namedPathItem.Value.Put = op
			case "DELETE":
				namedPathItem.Value.Delete = op
			case "PATCH":
				namedPathItem.Value.Patch = op
			}
			return
		}
	}
	// if we get here, we need to create a path item
	namedPathItem := &v3.NamedPathItem{Name: path, Value: &v3.PathItem{}}
	switch methodName {
	case "GET":
		namedPathItem.Value.Get = op
	case "POST":
		namedPathItem.Value.Post = op
	case "PUT":
		namedPathItem.Value.Put = op
	case "DELETE":
		namedPathItem.Value.Delete = op
	case "PATCH":
		namedPathItem.Value.Patch = op
	}
	d.Paths.Path = append(d.Paths.Path, namedPathItem)
}
