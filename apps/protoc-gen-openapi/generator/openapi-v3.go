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
	v3 "github.com/googleapis/gnostic/openapiv3"
	"google.golang.org/genproto/googleapis/api/annotations"
)

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
	d.Components = &v3.Components{}
	return d
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
			//log.Printf(" EXTENSIONS %+v", method.Options.XXX_InternalExtensions)

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
				op := g.buildOperationV3(operationID, comment, body, inputMessage, outputMessage)
				g.addOperationV3(d, op, path, methodName)
			}
		}
	}

}

// BuildDocumentV3 generates an OpenAPI description
func (g *Generator) BuildDocumentV3() *v3.Document {
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
	d.Paths.Path = append(d.Paths.Path,
		&v3.NamedPathItem{
			Name: "/pets",
			Value: &v3.PathItem{
				Get: &v3.Operation{
					Summary:     "List all pets",
					OperationId: "listPets",
					Tags:        []string{"pets"},
					Parameters: []*v3.ParameterOrReference{
						&v3.ParameterOrReference{
							Oneof: &v3.ParameterOrReference_Parameter{
								Parameter: &v3.Parameter{
									Name:        "limit",
									In:          "query",
									Description: "How many items to return at one time (max 100)",
									Required:    false,
									Schema: &v3.SchemaOrReference{
										Oneof: &v3.SchemaOrReference_Schema{
											Schema: &v3.Schema{
												Type:   "integer",
												Format: "int32",
											},
										},
									},
								},
							},
						},
					},
					Responses: &v3.Responses{
						Default: &v3.ResponseOrReference{
							Oneof: &v3.ResponseOrReference_Response{
								Response: &v3.Response{
									Description: "unexpected error",
									Content: &v3.MediaTypes{
										AdditionalProperties: []*v3.NamedMediaType{
											&v3.NamedMediaType{
												Name: "application/json",
												Value: &v3.MediaType{
													Schema: &v3.SchemaOrReference{
														Oneof: &v3.SchemaOrReference_Reference{
															Reference: &v3.Reference{
																XRef: "#/components/schemas/Error",
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
						ResponseOrReference: []*v3.NamedResponseOrReference{
							&v3.NamedResponseOrReference{
								Name: "200",
								Value: &v3.ResponseOrReference{
									Oneof: &v3.ResponseOrReference_Response{
										Response: &v3.Response{
											Description: "An paged array of pets", // [sic] match other examples
											Content: &v3.MediaTypes{
												AdditionalProperties: []*v3.NamedMediaType{
													&v3.NamedMediaType{
														Name: "application/json",
														Value: &v3.MediaType{
															Schema: &v3.SchemaOrReference{
																Oneof: &v3.SchemaOrReference_Reference{
																	Reference: &v3.Reference{
																		XRef: "#/components/schemas/Pets",
																	},
																},
															},
														},
													},
												},
											},
											Headers: &v3.HeadersOrReferences{
												AdditionalProperties: []*v3.NamedHeaderOrReference{
													&v3.NamedHeaderOrReference{
														Name: "x-next",
														Value: &v3.HeaderOrReference{
															Oneof: &v3.HeaderOrReference_Header{
																Header: &v3.Header{
																	Description: "A link to the next page of responses",
																	Schema: &v3.SchemaOrReference{
																		Oneof: &v3.SchemaOrReference_Schema{
																			Schema: &v3.Schema{
																				Type: "string",
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
						},
					},
				},
				Post: &v3.Operation{
					Summary:     "Create a pet",
					OperationId: "createPets",
					Tags:        []string{"pets"},
					Responses: &v3.Responses{
						Default: &v3.ResponseOrReference{
							Oneof: &v3.ResponseOrReference_Response{
								Response: &v3.Response{
									Description: "unexpected error",
									Content: &v3.MediaTypes{
										AdditionalProperties: []*v3.NamedMediaType{
											&v3.NamedMediaType{
												Name: "application/json",
												Value: &v3.MediaType{
													Schema: &v3.SchemaOrReference{
														Oneof: &v3.SchemaOrReference_Reference{
															Reference: &v3.Reference{
																XRef: "#/components/schemas/Error",
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
						ResponseOrReference: []*v3.NamedResponseOrReference{
							&v3.NamedResponseOrReference{
								Name: "201",
								Value: &v3.ResponseOrReference{
									Oneof: &v3.ResponseOrReference_Response{
										Response: &v3.Response{
											Description: "Null response",
										},
									},
								},
							},
						},
					},
				},
			}},
		&v3.NamedPathItem{
			Name: "/pets/{petId}",
			Value: &v3.PathItem{
				Get: &v3.Operation{
					Summary:     "Info for a specific pet",
					OperationId: "showPetById",
					Tags:        []string{"pets"},
					Parameters: []*v3.ParameterOrReference{
						&v3.ParameterOrReference{
							Oneof: &v3.ParameterOrReference_Parameter{
								Parameter: &v3.Parameter{
									Name:        "petId",
									In:          "path",
									Description: "The id of the pet to retrieve",
									Required:    true,
									Schema: &v3.SchemaOrReference{
										Oneof: &v3.SchemaOrReference_Schema{
											Schema: &v3.Schema{
												Type: "string",
											},
										},
									},
								},
							},
						},
					},
					Responses: &v3.Responses{
						Default: &v3.ResponseOrReference{
							Oneof: &v3.ResponseOrReference_Response{
								Response: &v3.Response{
									Description: "unexpected error",
									Content: &v3.MediaTypes{
										AdditionalProperties: []*v3.NamedMediaType{
											&v3.NamedMediaType{
												Name: "application/json",
												Value: &v3.MediaType{
													Schema: &v3.SchemaOrReference{
														Oneof: &v3.SchemaOrReference_Reference{
															Reference: &v3.Reference{
																XRef: "#/components/schemas/Error",
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
						ResponseOrReference: []*v3.NamedResponseOrReference{
							&v3.NamedResponseOrReference{
								Name: "200",
								Value: &v3.ResponseOrReference{
									Oneof: &v3.ResponseOrReference_Response{
										Response: &v3.Response{
											Description: "Expected response to a valid request",
											Content: &v3.MediaTypes{
												AdditionalProperties: []*v3.NamedMediaType{
													&v3.NamedMediaType{
														Name: "application/json",
														Value: &v3.MediaType{
															Schema: &v3.SchemaOrReference{
																Oneof: &v3.SchemaOrReference_Reference{
																	Reference: &v3.Reference{
																		XRef: "#/components/schemas/Pets",
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
				},
			}})
	d.Components = &v3.Components{
		Schemas: &v3.SchemasOrReferences{
			AdditionalProperties: []*v3.NamedSchemaOrReference{
				&v3.NamedSchemaOrReference{
					Name: "Pet",
					Value: &v3.SchemaOrReference{
						Oneof: &v3.SchemaOrReference_Schema{
							Schema: &v3.Schema{
								Required: []string{"id", "name"},
								Properties: &v3.Properties{
									AdditionalProperties: []*v3.NamedSchemaOrReference{
										&v3.NamedSchemaOrReference{
											Name: "id",
											Value: &v3.SchemaOrReference{
												Oneof: &v3.SchemaOrReference_Schema{
													Schema: &v3.Schema{
														Type:   "integer",
														Format: "int64",
													},
												},
											},
										},
										&v3.NamedSchemaOrReference{
											Name: "name",
											Value: &v3.SchemaOrReference{
												Oneof: &v3.SchemaOrReference_Schema{
													Schema: &v3.Schema{
														Type: "string",
													},
												},
											},
										},
										&v3.NamedSchemaOrReference{
											Name: "tag",
											Value: &v3.SchemaOrReference{
												Oneof: &v3.SchemaOrReference_Schema{
													Schema: &v3.Schema{
														Type: "string",
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
				&v3.NamedSchemaOrReference{
					Name: "Pets",
					Value: &v3.SchemaOrReference{
						Oneof: &v3.SchemaOrReference_Schema{
							Schema: &v3.Schema{
								Type: "array",
								Items: &v3.ItemsItem{
									SchemaOrReference: []*v3.SchemaOrReference{
										&v3.SchemaOrReference{
											Oneof: &v3.SchemaOrReference_Reference{
												Reference: &v3.Reference{
													XRef: "#/components/schemas/Pet",
												},
											},
										},
									},
								},
							},
						},
					},
				},
				&v3.NamedSchemaOrReference{
					Name: "Error",
					Value: &v3.SchemaOrReference{
						Oneof: &v3.SchemaOrReference_Schema{
							Schema: &v3.Schema{
								Required: []string{"code", "message"},
								Properties: &v3.Properties{
									AdditionalProperties: []*v3.NamedSchemaOrReference{
										&v3.NamedSchemaOrReference{
											Name: "code",
											Value: &v3.SchemaOrReference{
												Oneof: &v3.SchemaOrReference_Schema{
													Schema: &v3.Schema{
														Type:   "integer",
														Format: "int32",
													},
												},
											},
										},
										&v3.NamedSchemaOrReference{
											Name: "message",
											Value: &v3.SchemaOrReference{
												Oneof: &v3.SchemaOrReference_Schema{
													Schema: &v3.Schema{
														Type: "string",
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
	return d
}

func (g *Generator) buildOperationV3(
	operationID string,
	description string,
	bodyField string,
	inputMessage *Descriptor,
	outputMessage *Descriptor,
) *v3.Operation {

	parameters := []*v3.ParameterOrReference{}

	if bodyField != "" {
		parameters = append(parameters,
			&v3.ParameterOrReference{
				Oneof: &v3.ParameterOrReference_Parameter{
					Parameter: &v3.Parameter{
						Name:        "limit",
						In:          "body",
						Description: "something else",
						Required:    true,
						Schema: &v3.SchemaOrReference{
							Oneof: &v3.SchemaOrReference_Reference{
								Reference: &v3.Reference{
									XRef: "#/definitions/" + bodyField,
								},
							},
						},
					},
				},
			})
	}

	if false {
		parameters = append(parameters,
			&v3.ParameterOrReference{
				Oneof: &v3.ParameterOrReference_Parameter{
					Parameter: &v3.Parameter{
						Name:        "limit",
						In:          "query",
						Description: "How many items to return at one time (max 100)",
						Required:    false,
						Schema: &v3.SchemaOrReference{
							Oneof: &v3.SchemaOrReference_Schema{
								Schema: &v3.Schema{
									Type:   "integer",
									Format: "int32",
								},
							},
						},
					},
				},
			})
	}

	op := &v3.Operation{
		Summary:     description,
		OperationId: operationID,
		Tags:        []string{"tag"},
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
												XRef: "#/components/schemas/" + bodyField,
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

	return op
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
