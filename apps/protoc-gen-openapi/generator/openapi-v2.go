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
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	v2 "github.com/googleapis/gnostic/OpenAPIv2"
	"github.com/golang/protobuf/proto"
	options "google.golang.org/genproto/googleapis/api/annotations"
)

func (g *Generator) BuildDocumentV2(file *FileDescriptor) *v2.Document {
	g.file = file

	d := &v2.Document{}
	d.Swagger = "2.0"
	d.Info = &v2.Info{
		Title:   "Swagger Petstore",
		Version: "1.0.0",
		License: &v2.License{Name: "MIT"},
	}
	d.Host = "petstore.swagger.io"
	d.BasePath = "/v1"
	d.Schemes = []string{"http"}
	d.Consumes = []string{"application/json"}
	d.Produces = []string{"application/json"}
	d.Paths = &v2.Paths{}

	for _, service := range file.FileDescriptorProto.Service {
		log.Printf("SERVICE %s", *service.Name)
		for _, method := range service.Method {
			log.Printf("METHOD %s", *method.Name)
			log.Printf(" INPUT %s", *method.InputType)
			log.Printf(" OUTPUT %s", *method.OutputType)
			log.Printf(" OPTIONS %+v", *method.Options)
			log.Printf(" EXTENSIONS %+v", method.Options.XXX_InternalExtensions)

			extension, err := proto.GetExtension(method.Options, options.E_Http)
			log.Printf(" extensions: %T %+v (%+v)", extension, extension, err)
			if extension != nil {
				rule := extension.(*options.HttpRule)
				log.Printf("  PATTERN %T %v", rule.Pattern, rule.Pattern)
				log.Printf("  SELECTOR %s", rule.Selector)
				log.Printf("  BODY %s", rule.Body)
				log.Printf("  BINDINGS %s", rule.AdditionalBindings)
			}

			var methodName string
			name := *method.Name
			switch {
			case strings.HasPrefix(name, "Get"):
				methodName = "GET"
			case strings.HasPrefix(name, "List"):
				methodName = "GET"
			case strings.HasPrefix(name, "Create"):
				methodName = "POST"
			case strings.HasPrefix(name, "Delete"):
				methodName = "DELETE"
			default:
				methodName = "UNKNOWN"
			}
			log.Printf("%s", methodName)
		}
	}
	g.BuildSamplePaths(d)

	// for each message, generate a definition
	d.Definitions = &v2.Definitions{}
	for _, desc := range g.file.desc {
		definitionProperties := &v2.Properties{}
		for _, field := range desc.Field {
			fieldSchema := &v2.Schema{}
			if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
				fieldSchema.Type = &v2.TypeItem{[]string{"array"}}
				switch *field.Type {
				case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
					fieldSchema.Items =
						&v2.ItemsItem{[]*v2.Schema{&v2.Schema{
							XRef: g.definitionReferenceForTypeName(*field.TypeName)}}}
				default:
					log.Printf("(TODO) Unsupported array type: %+v", field.Type)
				}
			} else {
				switch *field.Type {
				case descriptor.FieldDescriptorProto_TYPE_STRING:
					fieldSchema.Type = &v2.TypeItem{[]string{"string"}}
				case descriptor.FieldDescriptorProto_TYPE_INT64:
					fieldSchema.Type = &v2.TypeItem{[]string{"integer"}}
					fieldSchema.Format = "int64"
				case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
					fieldSchema.XRef = g.definitionReferenceForTypeName(*field.TypeName)
				default:
					log.Printf("(TODO) Unsupported field type: %+v", field.Type)
				}
			}
			definitionProperties.AdditionalProperties = append(
				definitionProperties.AdditionalProperties,
				&v2.NamedSchema{
					Name:  *field.Name,
					Value: fieldSchema,
				},
			)
		}
		d.Definitions.AdditionalProperties = append(d.Definitions.AdditionalProperties,
			&v2.NamedSchema{
				Name:  *desc.Name,
				Value: &v2.Schema{Properties: definitionProperties},
			})
	}
	return d
}

func (g *Generator) definitionReferenceForTypeName(typeName string) string {
	parts := strings.Split(typeName, ".")
	lastPart := parts[len(parts)-1]
	return "#/definitions/" + lastPart
}

func (g *Generator) BuildSamplePaths(d *v2.Document) {
	d.Paths.Path = append(d.Paths.Path,
		&v2.NamedPathItem{
			Name: "/pets",
			Value: &v2.PathItem{
				Get: &v2.Operation{
					Summary:     "List all pets",
					OperationId: "listPets",
					Tags:        []string{"pets"},
					Parameters: []*v2.ParametersItem{
						&v2.ParametersItem{
							Oneof: &v2.ParametersItem_Parameter{
								Parameter: &v2.Parameter{
									Oneof: &v2.Parameter_NonBodyParameter{
										NonBodyParameter: &v2.NonBodyParameter{
											Oneof: &v2.NonBodyParameter_QueryParameterSubSchema{
												QueryParameterSubSchema: &v2.QueryParameterSubSchema{
													Name:        "limit",
													In:          "query",
													Description: "How many items to return at one time (max 100)",
													Required:    false,
													Type:        "integer",
													Format:      "int32",
												},
											},
										},
									},
								},
							},
						},
					},
					Responses: &v2.Responses{
						ResponseCode: []*v2.NamedResponseValue{
							&v2.NamedResponseValue{
								Name: "200",
								Value: &v2.ResponseValue{
									Oneof: &v2.ResponseValue_Response{
										Response: &v2.Response{
											Description: "An paged array of pets", // [sic] match other examples
											Schema: &v2.SchemaItem{
												Oneof: &v2.SchemaItem_Schema{
													Schema: &v2.Schema{
														XRef: "#/definitions/Pets",
													},
												},
											},
											Headers: &v2.Headers{
												AdditionalProperties: []*v2.NamedHeader{
													&v2.NamedHeader{
														Name: "x-next",
														Value: &v2.Header{
															Type:        "string",
															Description: "A link to the next page of responses",
														},
													},
												},
											},
										},
									},
								},
							},
							&v2.NamedResponseValue{
								Name: "default",
								Value: &v2.ResponseValue{
									Oneof: &v2.ResponseValue_Response{
										Response: &v2.Response{
											Description: "unexpected error",
											Schema: &v2.SchemaItem{
												Oneof: &v2.SchemaItem_Schema{
													Schema: &v2.Schema{
														XRef: "#/definitions/Error",
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
				Post: &v2.Operation{
					Summary:     "Create a pet",
					OperationId: "createPets",
					Tags:        []string{"pets"},
					Parameters:  []*v2.ParametersItem{},
					Responses: &v2.Responses{
						ResponseCode: []*v2.NamedResponseValue{
							&v2.NamedResponseValue{
								Name: "201",
								Value: &v2.ResponseValue{
									Oneof: &v2.ResponseValue_Response{
										Response: &v2.Response{
											Description: "Null response",
										},
									},
								},
							},
							&v2.NamedResponseValue{
								Name: "default",
								Value: &v2.ResponseValue{
									Oneof: &v2.ResponseValue_Response{
										Response: &v2.Response{
											Description: "unexpected error",
											Schema: &v2.SchemaItem{
												Oneof: &v2.SchemaItem_Schema{
													Schema: &v2.Schema{
														XRef: "#/definitions/Error",
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
			}})
	d.Paths.Path = append(d.Paths.Path,
		&v2.NamedPathItem{
			Name: "/pets/{petId}",
			Value: &v2.PathItem{
				Get: &v2.Operation{
					Summary:     "Info for a specific pet",
					OperationId: "showPetById",
					Tags:        []string{"pets"},
					Parameters: []*v2.ParametersItem{
						&v2.ParametersItem{
							Oneof: &v2.ParametersItem_Parameter{
								Parameter: &v2.Parameter{
									Oneof: &v2.Parameter_NonBodyParameter{
										NonBodyParameter: &v2.NonBodyParameter{
											Oneof: &v2.NonBodyParameter_PathParameterSubSchema{
												PathParameterSubSchema: &v2.PathParameterSubSchema{
													Name:        "petId",
													In:          "path",
													Description: "The id of the pet to retrieve",
													Required:    true,
													Type:        "string",
												},
											},
										},
									},
								},
							},
						},
					},
					Responses: &v2.Responses{
						ResponseCode: []*v2.NamedResponseValue{
							&v2.NamedResponseValue{
								Name: "200",
								Value: &v2.ResponseValue{
									Oneof: &v2.ResponseValue_Response{
										Response: &v2.Response{
											Description: "Expected response to a valid request",
											Schema: &v2.SchemaItem{
												Oneof: &v2.SchemaItem_Schema{
													Schema: &v2.Schema{
														XRef: "#/definitions/Pets",
													},
												},
											},
										},
									},
								},
							},
							&v2.NamedResponseValue{
								Name: "default",
								Value: &v2.ResponseValue{
									Oneof: &v2.ResponseValue_Response{
										Response: &v2.Response{
											Description: "unexpected error",
											Schema: &v2.SchemaItem{
												Oneof: &v2.SchemaItem_Schema{
													Schema: &v2.Schema{
														XRef: "#/definitions/Error",
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
			}})
}
