package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"

	"github.com/golang/protobuf/proto"
	pb "github.com/googleapis/gnostic/OpenAPIv2"
	"github.com/googleapis/gnostic/apps/discovery/disco"
	"github.com/googleapis/gnostic/compiler"
)

// Select an API.
const apiName = "people"
const apiVersion = "v1"

func main() {
	// Read the list of APIs from the apis/list service.
	apiListServiceURL := "https://www.googleapis.com/discovery/v1/apis"
	bytes, err := compiler.FetchFile(apiListServiceURL)
	if err != nil {
		log.Fatal("%+v", err)
	}
	// Unpack the apis/list response.
	listResponse, err := disco.NewList(bytes)
	// List the APIs.
	for _, api := range listResponse.APIs {
		fmt.Printf("%s\n", api.ID)
	}
	// Get the description of an API
	api := listResponse.APIWithID(apiName + ":" + apiVersion)
	if api == nil {
		log.Fatal("Error: API not found")
	}
	fmt.Printf("API: %+v\n", api)
	// Fetch the discovery description of the API.
	bytes, err = compiler.FetchFile(api.DiscoveryRestURL)
	if err != nil {
		log.Fatal("%+v", err)
	}
	// Unpack the discovery response.
	discoveryDocument, err := disco.NewDocument(bytes)
	if err != nil {
		log.Fatal("%+v", err)
	}
	fmt.Printf("DISCOVERY: %+v\n", discoveryDocument)
	// Generate the OpenAPI equivalent
	openAPIDocument := buildOpenAPI2Document(discoveryDocument)
	bytes, err = proto.Marshal(openAPIDocument)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(apiName+"-"+apiVersion+".pb", bytes, 0644)
	if err != nil {
		panic(err)
	}
}

func addOpenAPI2SchemaForSchema(d *pb.Document, name string, schema *disco.Schema) {
	fmt.Printf("SCHEMA %s\n", name)
	d.Definitions.AdditionalProperties = append(d.Definitions.AdditionalProperties,
		&pb.NamedSchema{
			Name:  schema.Name,
			Value: buildOpenAPI2SchemaForSchema(schema),
		})
}

func buildOpenAPI2SchemaForSchema(schema *disco.Schema) *pb.Schema {
	s := &pb.Schema{}

	if description := schema.Description; description != "" {
		s.Description = description
	}
	if typeName := schema.Type; typeName != "" {
		s.Type = &pb.TypeItem{[]string{typeName}}
	}
	if ref := schema.Ref; ref != "" {
		s.XRef = "#/definitions/" + ref
	}
	if len(schema.Enums) > 0 {
		for _, e := range schema.Enums {
			s.Enum = append(s.Enum, &pb.Any{Yaml: e})
		}
	}
	if schema.ItemSchema != nil {
		s2 := buildOpenAPI2SchemaForSchema(schema.ItemSchema)
		s.Items = &pb.ItemsItem{}
		s.Items.Schema = append(s.Items.Schema, s2)
	}
	if len(schema.Properties) > 0 {
		s.Properties = &pb.Properties{}
		for _, property := range schema.Properties {
			s.Properties.AdditionalProperties = append(s.Properties.AdditionalProperties,
				&pb.NamedSchema{
					Name:  property.Name,
					Value: buildOpenAPI2SchemaForSchema(property.Schema),
				},
			)
		}
	}
	return s
}

func buildOpenAPI2OperationForMethod(method *disco.Method) *pb.Operation {
	fmt.Printf("METHOD %s %s %s %s\n", method.Name, method.FlatPath, method.HTTPMethod, method.ID)
	//fmt.Printf("MAP %+v\n", method.JSONMap)

	parameters := make([]*pb.ParametersItem, 0)
	for _, p := range method.Parameters {
		fmt.Printf("- PARAMETER %+v\n", p)
		parameters = append(parameters,
			&pb.ParametersItem{
				Oneof: &pb.ParametersItem_Parameter{
					Parameter: &pb.Parameter{
						Oneof: &pb.Parameter_NonBodyParameter{
							NonBodyParameter: &pb.NonBodyParameter{
								Oneof: &pb.NonBodyParameter_QueryParameterSubSchema{
									QueryParameterSubSchema: &pb.QueryParameterSubSchema{
										Name:        p.Name,
										In:          "XXX",
										Description: p.Description,
										Required:    false,
										Type:        "XXX",
										Format:      "XXX",
									},
								},
							},
						},
					},
				},
			})
	}

	fmt.Printf("- RESPONSE %+v\n", method.Response)

	responses := &pb.Responses{
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
	}

	return &pb.Operation{
		Summary:     method.Description,
		OperationId: method.ID,
		Responses:   responses,
		Parameters:  parameters,
	}
}

func getOpenAPI2PathItemForPath(d *pb.Document, path string) *pb.PathItem {
	// First, try to find a path item with the specified path. If it exists, return it.
	for _, item := range d.Paths.Path {
		if item.Name == path {
			return item.Value
		}
	}
	// Otherwise, create and return a new path item.
	pathItem := &pb.PathItem{}
	d.Paths.Path = append(d.Paths.Path,
		&pb.NamedPathItem{
			Name:  path,
			Value: pathItem,
		},
	)
	return pathItem
}

func addOpenAPI2PathsForMethod(d *pb.Document, method *disco.Method) {
	operation := buildOpenAPI2OperationForMethod(method)
	pathItem := getOpenAPI2PathItemForPath(d, "/"+method.FlatPath)
	switch method.HTTPMethod {
	case "GET":
		pathItem.Get = operation
	case "POST":
		pathItem.Post = operation
	case "PUT":
		pathItem.Put = operation
	case "DELETE":
		pathItem.Delete = operation
	default:
		log.Printf("UNKNOWN HTTP METHOD %s", method.HTTPMethod)
	}
}

func addOpenAPI2PathsForResource(d *pb.Document, resource *disco.Resource) {
	fmt.Printf("RESOURCE %s (%s)\n", resource.Name, resource.FullName)
	for _, method := range resource.Methods {
		addOpenAPI2PathsForMethod(d, method)
	}
	for _, resource2 := range resource.Resources {
		addOpenAPI2PathsForResource(d, resource2)
	}
}

func buildOpenAPI2Document(api *disco.Document) *pb.Document {
	d := &pb.Document{}
	d.Swagger = "2.0"
	d.Info = &pb.Info{
		Title:       api.Title,
		Version:     api.Version,
		Description: api.Description,
	}

	url, _ := url.Parse(api.RootURL)
	d.Host = url.Host
	d.BasePath = url.Path
	d.Schemes = []string{url.Scheme}
	d.Consumes = []string{"application/json"}
	d.Produces = []string{"application/json"}
	d.Paths = &pb.Paths{}
	d.Definitions = &pb.Definitions{}

	for name, schema := range api.Schemas {
		addOpenAPI2SchemaForSchema(d, name, schema)
	}

	for _, method := range api.Methods {
		addOpenAPI2PathsForMethod(d, method)
	}

	for _, resource := range api.Resources {
		addOpenAPI2PathsForResource(d, resource)
	}

	/*
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
	*/
	/*
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
	*/
	return d
}
