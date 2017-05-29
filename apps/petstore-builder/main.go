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

package main

import (
	"io/ioutil"

	"github.com/golang/protobuf/proto"
	pb "github.com/googleapis/gnostic/OpenAPIv2"
)

func buildDocument() *pb.Document {
	d := &pb.Document{}
	d.Swagger = "2.0"
	d.Info = &pb.Info{
		Title:   "Swagger Petstore",
		Version: "1.0.0",
		License: &pb.License{Name: "MIT"},
	}
	d.Host = "petstore.swagger.io"
	d.BasePath = "/v1"
	d.Schemes = []string{"http"}
	d.Consumes = []string{"application/json"}
	d.Produces = []string{"application/json"}
	d.Paths = &pb.Paths{}
	d.Paths.Path = append(d.Paths.Path,
		&pb.NamedPathItem{
			Name: "/pets",
			Value: &pb.PathItem{
				Get: &pb.Operation{
					Summary:     "List all pets",
					OperationId: "listPets",
					Tags:        []string{"pets"},
					Parameters: []*pb.ParametersItem{
						&pb.ParametersItem{
							Oneof: &pb.ParametersItem_Parameter{
								Parameter: &pb.Parameter{
									Oneof: &pb.Parameter_NonBodyParameter{
										NonBodyParameter: &pb.NonBodyParameter{
											Oneof: &pb.NonBodyParameter_QueryParameterSubSchema{
												QueryParameterSubSchema: &pb.QueryParameterSubSchema{
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
					Responses: &pb.Responses{
						ResponseCode: []*pb.NamedResponseValue{
							&pb.NamedResponseValue{
								Name: "200",
								Value: &pb.ResponseValue{
									Oneof: &pb.ResponseValue_Response{
										Response: &pb.Response{
											Description: "An paged array of pets", // [sic] match other examples
											Schema: &pb.SchemaItem{
												Oneof: &pb.SchemaItem_Schema{
													Schema: &pb.Schema{
														XRef: "#/definitions/Pets",
													},
												},
											},
											Headers: &pb.Headers{
												AdditionalProperties: []*pb.NamedHeader{
													&pb.NamedHeader{
														Name: "x-next",
														Value: &pb.Header{
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
							&pb.NamedResponseValue{
								Name: "default",
								Value: &pb.ResponseValue{
									Oneof: &pb.ResponseValue_Response{
										Response: &pb.Response{
											Description: "unexpected error",
											Schema: &pb.SchemaItem{
												Oneof: &pb.SchemaItem_Schema{
													Schema: &pb.Schema{
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
				Post: &pb.Operation{
					Summary:     "Create a pet",
					OperationId: "createPets",
					Tags:        []string{"pets"},
					Parameters:  []*pb.ParametersItem{},
					Responses: &pb.Responses{
						ResponseCode: []*pb.NamedResponseValue{
							&pb.NamedResponseValue{
								Name: "201",
								Value: &pb.ResponseValue{
									Oneof: &pb.ResponseValue_Response{
										Response: &pb.Response{
											Description: "Null response",
										},
									},
								},
							},
							&pb.NamedResponseValue{
								Name: "default",
								Value: &pb.ResponseValue{
									Oneof: &pb.ResponseValue_Response{
										Response: &pb.Response{
											Description: "unexpected error",
											Schema: &pb.SchemaItem{
												Oneof: &pb.SchemaItem_Schema{
													Schema: &pb.Schema{
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
		&pb.NamedPathItem{
			Name: "/pets/{petId}",
			Value: &pb.PathItem{
				Get: &pb.Operation{
					Summary:     "Info for a specific pet",
					OperationId: "showPetById",
					Tags:        []string{"pets"},
					Parameters: []*pb.ParametersItem{
						&pb.ParametersItem{
							Oneof: &pb.ParametersItem_Parameter{
								Parameter: &pb.Parameter{
									Oneof: &pb.Parameter_NonBodyParameter{
										NonBodyParameter: &pb.NonBodyParameter{
											Oneof: &pb.NonBodyParameter_PathParameterSubSchema{
												PathParameterSubSchema: &pb.PathParameterSubSchema{
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
					Responses: &pb.Responses{
						ResponseCode: []*pb.NamedResponseValue{
							&pb.NamedResponseValue{
								Name: "200",
								Value: &pb.ResponseValue{
									Oneof: &pb.ResponseValue_Response{
										Response: &pb.Response{
											Description: "Expected response to a valid request",
											Schema: &pb.SchemaItem{
												Oneof: &pb.SchemaItem_Schema{
													Schema: &pb.Schema{
														XRef: "#/definitions/Pets",
													},
												},
											},
										},
									},
								},
							},
							&pb.NamedResponseValue{
								Name: "default",
								Value: &pb.ResponseValue{
									Oneof: &pb.ResponseValue_Response{
										Response: &pb.Response{
											Description: "unexpected error",
											Schema: &pb.SchemaItem{
												Oneof: &pb.SchemaItem_Schema{
													Schema: &pb.Schema{
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
	d.Definitions = &pb.Definitions{}
	d.Definitions.AdditionalProperties = append(d.Definitions.AdditionalProperties,
		&pb.NamedSchema{
			Name: "Pet",
			Value: &pb.Schema{
				Required: []string{"id", "name"},
				Properties: &pb.Properties{
					AdditionalProperties: []*pb.NamedSchema{
						&pb.NamedSchema{Name: "id", Value: &pb.Schema{
							Type:   &pb.TypeItem{[]string{"integer"}},
							Format: "int64"}},
						&pb.NamedSchema{Name: "name", Value: &pb.Schema{Type: &pb.TypeItem{[]string{"string"}}}},
						&pb.NamedSchema{Name: "tag", Value: &pb.Schema{Type: &pb.TypeItem{[]string{"string"}}}},
					},
				},
			}})
	d.Definitions.AdditionalProperties = append(d.Definitions.AdditionalProperties,
		&pb.NamedSchema{
			Name: "Pets",
			Value: &pb.Schema{
				Type:  &pb.TypeItem{[]string{"array"}},
				Items: &pb.ItemsItem{[]*pb.Schema{&pb.Schema{XRef: "#/definitions/Pet"}}},
			}})
	d.Definitions.AdditionalProperties = append(d.Definitions.AdditionalProperties,
		&pb.NamedSchema{
			Name: "Error",
			Value: &pb.Schema{
				Required: []string{"code", "message"},
				Properties: &pb.Properties{
					AdditionalProperties: []*pb.NamedSchema{
						&pb.NamedSchema{Name: "code", Value: &pb.Schema{
							Type:   &pb.TypeItem{[]string{"integer"}},
							Format: "int32"}},
						&pb.NamedSchema{Name: "message", Value: &pb.Schema{Type: &pb.TypeItem{[]string{"string"}}}},
					},
				},
			}})
	return d
}

func main() {
	document := buildDocument()
	bytes, err := proto.Marshal(document)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile("petstore.pb", bytes, 0644)
	if err != nil {
		panic(err)
	}
}
