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

package generator

import (
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	v3 "github.com/googleapis/gnostic/openapiv3"
	"google.golang.org/genproto/googleapis/api/annotations"
)

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func singular(plural string) string {
	if strings.HasSuffix(plural, "ies") {
		return strings.TrimSuffix(plural, "ies") + "y"
	}
	if strings.HasSuffix(plural, "s") {
		return strings.TrimSuffix(plural, "s")
	}
	return plural
}

func (g *Generator) filterCommentString(comment string) string {
	comment = strings.Replace(comment, "\n", "", -1)
	comment = g.linterRulePattern.ReplaceAllString(comment, "")
	return strings.TrimSpace(comment)
}

func (g *Generator) commentForPath2(file *FileDescriptor, servicePath, s int) string {
	path := []string{
		strconv.Itoa(servicePath),
		strconv.Itoa(s),
	}
	return g.filterCommentString(file.comments[strings.Join(path, ",")].GetLeadingComments())
}

func (g *Generator) commentForPath4(file *FileDescriptor, servicePath, s, messageFieldPath, m int) string {
	path := []string{
		strconv.Itoa(servicePath),
		strconv.Itoa(s),
		strconv.Itoa(messageFieldPath),
		strconv.Itoa(m),
	}
	return g.filterCommentString(file.comments[strings.Join(path, ",")].GetLeadingComments())
}

// schemaReferenceForTypeName returns an OpenAPI JSON Reference to the schema that represents a type.
func (g *Generator) schemaReferenceForTypeName(typeName string) string {
	if !contains(g.requiredSchemas, typeName) {
		g.requiredSchemas = append(g.requiredSchemas, typeName)
	}
	parts := strings.Split(typeName, ".")
	lastPart := parts[len(parts)-1]
	return "#/components/schemas/" + lastPart
}

// GenerateOpenAPIv3 creates a new OpenAPIv3 document
func (g *Generator) GenerateOpenAPIv3() *v3.Document {
	g.requiredSchemas = make([]string, 0)
	g.generatedSchemas = make([]string, 0)

	g.linterRulePattern = regexp.MustCompile(`\(-- .* --\)`)
	g.namePattern = regexp.MustCompile("{(.*)=(.*)}")

	d := &v3.Document{}
	d.Openapi = "3.0"
	d.Info = &v3.Info{
		Title:       "",
		Version:     "0.0.1",
		Description: "",
	}
	d.Paths = &v3.Paths{}
	d.Components = &v3.Components{
		Schemas: &v3.SchemasOrReferences{
			AdditionalProperties: []*v3.NamedSchemaOrReference{},
		},
	}
	for _, file := range g.allFiles {
		g.addPathsToDocumentV3(d, file)
	}
	for len(g.requiredSchemas) > 0 {
		count := len(g.requiredSchemas)
		for _, file := range g.allFiles {
			g.addSchemasToDocumentV3(d, file)
		}
		g.requiredSchemas = g.requiredSchemas[count:len(g.requiredSchemas)]
	}
	// sort the paths
	{
		pairs := d.Paths.Path
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].Name < pairs[j].Name
		})
		d.Paths.Path = pairs
	}
	// sort the schemas
	{
		pairs := d.Components.Schemas.AdditionalProperties
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].Name < pairs[j].Name
		})
		d.Components.Schemas.AdditionalProperties = pairs
	}
	return d
}

// addPathsToDocumentV3 adds paths from a specified file descriptor
func (g *Generator) addPathsToDocumentV3(d *v3.Document, file *FileDescriptor) {
	for s, service := range file.FileDescriptorProto.Service {
		comment := g.commentForPath2(file, servicePath, s)
		d.Info.Title = *service.Name
		d.Info.Description = comment

		for m, method := range service.Method {
			comment := g.commentForPath4(file, servicePath, s, messageFieldPath, m)
			inputType := *method.InputType
			inputIndex, inputMessage := g.FindMessage(inputType)
			outputType := *method.OutputType
			outputIndex, outputMessage := g.FindMessage(outputType)
			operationID := *service.Name + "_" + *method.Name
			extension, err := proto.GetExtension(method.Options, annotations.E_Http)
			if err != nil {
				log.Printf("%s", err.Error())
			}
			var path string
			var methodName string
			var body string
			if extension != nil {
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
					file, operationID, comment, path, body,
					inputType, inputIndex, inputMessage,
					outputType, outputIndex, outputMessage)
				g.addOperationV3(d, op, path2, methodName)
			}
		}
	}
}

// buildOperationV3 constructs an operation for a set of values
func (g *Generator) buildOperationV3(
	file *FileDescriptor,
	operationID string,
	description string,
	path string,
	bodyField string,
	inputType string,
	inputIndex int,
	inputMessage *Descriptor,
	outputType string,
	outputIndex int,
	outputMessage *Descriptor,
) (*v3.Operation, string) {

	// coveredParameters tracks the parameters that have been used in the body or path
	coveredParameters := make([]string, 0)
	if bodyField != "" {
		coveredParameters = append(coveredParameters, bodyField)
	}

	// initialize the list of operation parameters
	parameters := []*v3.ParameterOrReference{}

	// build a list of path parameters
	pathParameters := make([]string, 0)
	if matches := g.namePattern.FindStringSubmatch(path); matches != nil {
		// add the "name=" "name" value to the list of covered parameters
		coveredParameters = append(coveredParameters, matches[1])
		// convert the path from the starred form to use named path parameters
		starredPath := matches[2]
		parts := strings.Split(starredPath, "/")
		// the starred path is assumed to be in the form "things/*/otherthings/*"
		// we want to convert it to "things/{thing}/otherthings/{otherthing}"
		for i := 0; i < len(parts); i += 2 {
			section := parts[i]
			parameter := singular(section)
			parts[i+1] = "{" + parameter + "}"
			pathParameters = append(pathParameters, parameter)
		}
		// rewrite the path to use the path parameters
		newPath := strings.Join(parts, "/")
		path = strings.Replace(path, matches[0], newPath, 1)
	}
	// add the path parameters to the operation parameters
	for _, pathParameter := range pathParameters {
		parameters = append(parameters,
			&v3.ParameterOrReference{
				Oneof: &v3.ParameterOrReference_Parameter{
					Parameter: &v3.Parameter{
						Name:        pathParameter,
						In:          "path",
						Required:    true,
						Description: "The " + pathParameter + " id.",
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

	// add any unhandled fields in the request message as query parameters
	if bodyField != "*" {
		for fieldIndex, field := range inputMessage.Field {
			fieldName := *field.Name
			if !contains(coveredParameters, fieldName) {
				// get the field description from the comments
				fieldDescription := g.commentForPath4(file, messagePath, inputIndex, messageFieldPath, fieldIndex)
				parameters = append(parameters,
					&v3.ParameterOrReference{
						Oneof: &v3.ParameterOrReference_Parameter{
							Parameter: &v3.Parameter{
								Name:        fieldName,
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

	// create the response
	responses := &v3.Responses{
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
														XRef: g.schemaReferenceForTypeName(outputType),
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

	// create the operation
	op := &v3.Operation{
		Summary:     description,
		OperationId: operationID,
		Parameters:  parameters,
		Responses:   responses,
	}

	// if a body field is specified, we need to pass a message as the request body
	if bodyField != "" {
		var bodyFieldTypeName string
		if bodyField == "*" {
			// pass the entire request message as the request body
			bodyFieldTypeName = inputType
		} else {
			// if body refers to a message field, use that type
			for _, field := range inputMessage.Field {
				if *field.Name == bodyField {
					bodyFieldTypeName = *field.TypeName
					break
				}
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
												XRef: g.schemaReferenceForTypeName(bodyFieldTypeName),
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

// addOperationV3 adds an operation to the specified path/method
func (g *Generator) addOperationV3(d *v3.Document, op *v3.Operation, path string, methodName string) {
	var selectedPathItem *v3.NamedPathItem
	for _, namedPathItem := range d.Paths.Path {
		if namedPathItem.Name == path {
			selectedPathItem = namedPathItem
			break
		}
	}
	// if we get here, we need to create a path item
	if selectedPathItem == nil {
		selectedPathItem = &v3.NamedPathItem{Name: path, Value: &v3.PathItem{}}
		d.Paths.Path = append(d.Paths.Path, selectedPathItem)
	}
	// set the operation on the specified method
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

func itemsItemForTypeName(typeName string) *v3.ItemsItem {
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

// addSchemasToDocumentV3 adds info from one file descriptor
func (g *Generator) addSchemasToDocumentV3(d *v3.Document, file *FileDescriptor) {
	// for each message, generate a definition
	for i, desc := range file.desc {
		typeName := "." + *file.Package + "." + *desc.Name
		// only generate this if we need it and haven't already generated it
		if !contains(g.requiredSchemas, typeName) ||
			contains(g.generatedSchemas, typeName) {
			continue
		}
		g.generatedSchemas = append(g.generatedSchemas, typeName)
		// get the message description from the comments
		messageDescription := g.commentForPath2(file, messagePath, i)
		// build an array holding the fields of the message
		definitionProperties := &v3.Properties{
			AdditionalProperties: make([]*v3.NamedSchemaOrReference, 0),
		}
		for j, field := range desc.Field {
			// check the field annotations to see if this is a readonly field
			outputOnly := false
			extension, err := proto.GetExtension(field.Options, annotations.E_FieldBehavior)
			if err == nil {
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
			// get the field description from the comments
			fieldDescription := g.commentForPath4(file, messagePath, i, messageFieldPath, j)
			// the field is either described by a reference or a schema
			XRef := ""
			fieldSchema := &v3.Schema{
				Description: fieldDescription,
			}
			if outputOnly {
				fieldSchema.ReadOnly = true
			}
			if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
				fieldSchema.Type = "array"
				switch *field.Type {
				case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
					fieldSchema.Items = itemsItemForReference(g.schemaReferenceForTypeName(*field.TypeName))
				case descriptor.FieldDescriptorProto_TYPE_STRING:
					fieldSchema.Items = itemsItemForTypeName("string")
				case descriptor.FieldDescriptorProto_TYPE_INT32,
					descriptor.FieldDescriptorProto_TYPE_UINT32,
					descriptor.FieldDescriptorProto_TYPE_INT64,
					descriptor.FieldDescriptorProto_TYPE_UINT64:
					fieldSchema.Items = itemsItemForTypeName("integer")
				case descriptor.FieldDescriptorProto_TYPE_ENUM:
					fieldSchema.Items = itemsItemForTypeName("integer")
				case descriptor.FieldDescriptorProto_TYPE_BOOL:
					fieldSchema.Items = itemsItemForTypeName("boolean")
				case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
					fieldSchema.Items = itemsItemForTypeName("number")
				case descriptor.FieldDescriptorProto_TYPE_BYTES:
					fieldSchema.Items = itemsItemForTypeName("string")
				default:
					log.Printf("(TODO) Unsupported array type: %+v", field.Type)
				}
			} else {
				switch *field.Type {
				case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
					// the field is described by a reference
					XRef = g.schemaReferenceForTypeName(*field.TypeName)
				case descriptor.FieldDescriptorProto_TYPE_STRING:
					fieldSchema.Type = "string"
				case descriptor.FieldDescriptorProto_TYPE_INT32:
					fieldSchema.Type = "integer"
					fieldSchema.Format = "int32"
				case descriptor.FieldDescriptorProto_TYPE_UINT32:
					fieldSchema.Type = "integer"
					fieldSchema.Format = "uint32"
				case descriptor.FieldDescriptorProto_TYPE_INT64:
					fieldSchema.Type = "integer"
					fieldSchema.Format = "int64"
				case descriptor.FieldDescriptorProto_TYPE_UINT64:
					fieldSchema.Type = "integer"
					fieldSchema.Format = "uint64"
				case descriptor.FieldDescriptorProto_TYPE_ENUM:
					fieldSchema.Type = "integer"
					fieldSchema.Format = "enum"
				case descriptor.FieldDescriptorProto_TYPE_BOOL:
					fieldSchema.Type = "boolean"
				case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
					fieldSchema.Type = "number"
					fieldSchema.Format = "double"
				case descriptor.FieldDescriptorProto_TYPE_BYTES:
					fieldSchema.Type = "string"
					fieldSchema.Format = "bytes"
				default:
					log.Printf("(TODO) Unsupported field type: %+v", field.Type)
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
					Name:  *field.Name,
					Value: value,
				},
			)
		}
		// add the schema to the components.schema list
		d.Components.Schemas.AdditionalProperties = append(d.Components.Schemas.AdditionalProperties,
			&v3.NamedSchemaOrReference{
				Name: *desc.Name,
				Value: &v3.SchemaOrReference{
					Oneof: &v3.SchemaOrReference_Schema{
						Schema: &v3.Schema{
							Description: messageDescription,
							Properties:  definitionProperties,
						},
					},
				},
			})
	}
}
