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
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	v3 "github.com/googleapis/gnostic/openapiv3"
	"google.golang.org/genproto/googleapis/api/annotations"
)

var linterRulePattern *regexp.Regexp

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

	linterRulePattern = regexp.MustCompile(`\(-- .* --\)`)

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
	if true {
		pairs := d.Paths.Path
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].Name < pairs[j].Name
		})
		d.Paths.Path = pairs
	}
	// sort the schemas
	if true {
		pairs := d.Components.Schemas.AdditionalProperties
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].Name < pairs[j].Name
		})
		d.Components.Schemas.AdditionalProperties = pairs
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

// addPathsToDocumentV3 adds info from one file descriptor
func (g *Generator) addPathsToDocumentV3(d *v3.Document, file *FileDescriptor) {
	g.file = file
	sourceCodeInfo := file.SourceCodeInfo

	for s, service := range file.FileDescriptorProto.Service {
		var comment string
		for _, location := range sourceCodeInfo.Location {
			paths := location.GetPath()
			if len(paths) == 2 &&
				paths[0] == servicePath &&
				paths[1] == int32(s) {
				comment = location.GetLeadingComments()
				comment = strings.Replace(comment, "\n", "", -1)
				comment = linterRulePattern.ReplaceAllString(comment, "")
				comment = strings.TrimSpace(comment)
			}

			d.Info.Title = *service.Name
			d.Info.Description = comment
		}

		for m, method := range service.Method {
			// search source info for leading comments
			var comment string
			for _, location := range sourceCodeInfo.Location {
				paths := location.GetPath()
				if len(paths) == 4 &&
					paths[0] == servicePath &&
					paths[1] == int32(s) &&
					paths[2] == messageFieldPath &&
					paths[3] == int32(m) {
					comment = location.GetLeadingComments()
					comment = strings.Replace(comment, "\n", "", -1)
					comment = linterRulePattern.ReplaceAllString(comment, "")
					comment = strings.TrimSpace(comment)
				}
			}
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
					operationID, comment, path, body,
					inputType, inputIndex, inputMessage,
					outputType, outputIndex, outputMessage)
				g.addOperationV3(d, op, path2, methodName)
			}
		}
	}
}

// addSchemasToDocumentV3 adds info from one file descriptor
func (g *Generator) addSchemasToDocumentV3(d *v3.Document, file *FileDescriptor) {
	g.file = file
	// for each message, generate a definition
	for i, desc := range g.file.desc {

		var messageDescription string
		for _, location := range file.SourceCodeInfo.Location {
			paths := location.GetPath()
			if len(paths) == 2 &&
				paths[0] == messagePath &&
				paths[1] == int32(i) {
				comment := location.GetLeadingComments()
				comment = strings.Replace(comment, "\n", "", -1)
				comment = linterRulePattern.ReplaceAllString(comment, "")
				comment = strings.TrimSpace(comment)
				messageDescription = comment
			}
		}

		typeName := "." + *g.file.Package + "." + *desc.Name
		if !contains(g.requiredSchemas, typeName) ||
			contains(g.generatedSchemas, typeName) {
			continue
		}
		g.generatedSchemas = append(g.generatedSchemas, typeName)
		definitionProperties := &v3.Properties{
			AdditionalProperties: make([]*v3.NamedSchemaOrReference, 0),
		}
		for j, field := range desc.Field {

			var fieldDescription string
			for _, location := range file.SourceCodeInfo.Location {
				paths := location.GetPath()
				if len(paths) == 4 &&
					paths[0] == messagePath &&
					paths[1] == int32(i) &&
					paths[2] == messageFieldPath &&
					paths[3] == int32(j) {
					comment := location.GetLeadingComments()
					comment = strings.Replace(comment, "\n", "", -1)
					comment = linterRulePattern.ReplaceAllString(comment, "")
					comment = strings.TrimSpace(comment)
					fieldDescription = comment
				}
			}

			XRef := ""
			fieldSchema := &v3.Schema{}
			fieldSchema.Description = fieldDescription
			if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
				fieldSchema.Type = "array"
				switch *field.Type {
				case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
					fieldSchema.Items = itemsItemForReference(g.schemaReferenceForTypeName(*field.TypeName))
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
					XRef = g.schemaReferenceForTypeName(*field.TypeName)
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
					fieldSchema.Type = "number"
					fieldSchema.Format = "double"
				case descriptor.FieldDescriptorProto_TYPE_BYTES:
					fieldSchema.Type = "string"
					fieldSchema.Format = "bytes"
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
								Reference: &v3.Reference{
									XRef: XRef},
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
						Schema: &v3.Schema{
							Description: messageDescription,
							Properties:  definitionProperties},
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
	if strings.HasSuffix(plural, "ies") {
		return strings.TrimSuffix(plural, "ies") + "y"
	}
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
	inputType string,
	inputIndex int,
	inputMessage *Descriptor,
	outputType string,
	outputIndex int,
	outputMessage *Descriptor,
) (*v3.Operation, string) {

	namePattern := regexp.MustCompile("{(.*)=(.*)}")

	// track the parameters that are covered in the body or path
	coveredParameters := make([]string, 0)
	if bodyField != "" {
		coveredParameters = append(coveredParameters, bodyField)
	}

	parameters := []*v3.ParameterOrReference{}

	// build a list of path parameters
	pathParameters := make([]string, 0)
	if matches := namePattern.FindStringSubmatch(path); matches != nil {
		coveredParameters = append(coveredParameters, matches[1])
		starredPath := matches[2]
		parts := strings.Split(starredPath, "/")
		for i := 0; i < len(parts); i += 2 {
			section := parts[i]
			parameter := singular(section)
			parts[i+1] = "{" + parameter + "}"
			pathParameters = append(pathParameters, parameter)
		}
		newPath := strings.Join(parts, "/")
		path = strings.Replace(path, matches[0], newPath, 1)
	}
	// add the path parameters
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

	// add any unhandled parameters as query parameters
	if bodyField != "*" {
		for fieldIndex, field := range inputMessage.Field {
			fieldName := *field.Name
			if !contains(coveredParameters, fieldName) {

				var fieldDescription string
				for _, location := range g.file.SourceCodeInfo.Location {
					paths := location.GetPath()
					if len(paths) == 4 &&
						paths[0] == messagePath &&
						paths[1] == int32(inputIndex) &&
						paths[2] == messageFieldPath &&
						paths[3] == int32(fieldIndex) {
						comment := location.GetLeadingComments()
						comment = strings.Replace(comment, "\n", "", -1)
						comment = linterRulePattern.ReplaceAllString(comment, "")
						comment = strings.TrimSpace(comment)
						fieldDescription = comment
					}
				}

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
		},
	}

	if bodyField != "" {
		bodyFieldTypeName := ""
		if bodyField == "*" {
			bodyFieldTypeName = inputType
		} else {
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

func addOpToPathForMethod(
	op *v3.Operation,
	pathItem *v3.PathItem,
	methodName string) {
	switch methodName {
	case "GET":
		pathItem.Get = op
	case "POST":
		pathItem.Post = op
	case "PUT":
		pathItem.Put = op
	case "DELETE":
		pathItem.Delete = op
	case "PATCH":
		pathItem.Patch = op
	}
}

func (g *Generator) addOperationV3(d *v3.Document, op *v3.Operation, path string, methodName string) {
	for _, namedPathItem := range d.Paths.Path {
		if namedPathItem.Name == path {
			addOpToPathForMethod(op, namedPathItem.Value, methodName)
			return
		}
	}
	// if we get here, we need to create a path item
	namedPathItem := &v3.NamedPathItem{Name: path, Value: &v3.PathItem{}}
	addOpToPathForMethod(op, namedPathItem.Value, methodName)
	d.Paths.Path = append(d.Paths.Path, namedPathItem)
}
