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
		log.Fatalf("%+v", err)
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
		log.Fatalf("Error: API not found")
	}
	//fmt.Printf("API: %+v\n", api)
	// Fetch the discovery description of the API.
	bytes, err = compiler.FetchFile(api.DiscoveryRestURL)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	// Unpack the discovery response.
	discoveryDocument, err := disco.NewDocument(bytes)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	//fmt.Printf("DISCOVERY: %+v\n", discoveryDocument)
	// Generate the OpenAPI equivalent
	openAPIDocument := buildOpenAPI2DocumentForDocument(discoveryDocument)
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
	//log.Printf("SCHEMA %s\n", name)
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

func buildOpenAPI2ParameterForParameter(p *disco.Parameter) *pb.Parameter {
	//log.Printf("- PARAMETER %+v\n", p.Name)
	typeName := p.Schema.Type
	format := p.Schema.Format
	location := p.Location
	switch location {
	case "query":
		return &pb.Parameter{
			Oneof: &pb.Parameter_NonBodyParameter{
				NonBodyParameter: &pb.NonBodyParameter{
					Oneof: &pb.NonBodyParameter_QueryParameterSubSchema{
						QueryParameterSubSchema: &pb.QueryParameterSubSchema{
							Name:        p.Name,
							In:          "query",
							Description: p.Description,
							Required:    false,
							Type:        typeName,
							Format:      format,
						},
					},
				},
			},
		}
	case "path":
		return &pb.Parameter{
			Oneof: &pb.Parameter_NonBodyParameter{
				NonBodyParameter: &pb.NonBodyParameter{
					Oneof: &pb.NonBodyParameter_PathParameterSubSchema{
						PathParameterSubSchema: &pb.PathParameterSubSchema{
							Name:        p.Name,
							In:          "path",
							Description: p.Description,
							Required:    false,
							Type:        typeName,
							Format:      format,
						},
					},
				},
			},
		}
	default:
		return nil
	}
}

func buildOpenAPI2ResponseForSchema(schema *disco.Schema) *pb.Response {
	//log.Printf("- RESPONSE %+v\n", schema)
	ref := schema.Ref
	if ref == "" {
		log.Printf("ERROR: UNHANDLED RESPONSE SCHEMA %+v", schema)
	}
	return &pb.Response{
		Description: "Successful operation",
		Schema: &pb.SchemaItem{
			Oneof: &pb.SchemaItem_Schema{
				Schema: &pb.Schema{
					XRef: "#/definitions/" + ref,
				},
			},
		},
	}
}

func buildOpenAPI2OperationForMethod(method *disco.Method) *pb.Operation {
	//log.Printf("METHOD %s %s %s %s\n", method.Name, method.FlatPath, method.HTTPMethod, method.ID)
	//log.Printf("MAP %+v\n", method.JSONMap)
	parameters := make([]*pb.ParametersItem, 0)
	for _, p := range method.Parameters {
		parameters = append(parameters, &pb.ParametersItem{
			Oneof: &pb.ParametersItem_Parameter{
				Parameter: buildOpenAPI2ParameterForParameter(p),
			},
		})
	}
	responses := &pb.Responses{
		ResponseCode: []*pb.NamedResponseValue{
			&pb.NamedResponseValue{
				Name: "default",
				Value: &pb.ResponseValue{
					Oneof: &pb.ResponseValue_Response{
						Response: buildOpenAPI2ResponseForSchema(method.Response),
					},
				},
			},
		},
	}
	return &pb.Operation{
		Summary:     method.Description,
		OperationId: method.ID,
		Parameters:  parameters,
		Responses:   responses,
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
		log.Printf("ERROR: UNKNOWN HTTP METHOD %s", method.HTTPMethod)
	}
}

func addOpenAPI2PathsForResource(d *pb.Document, resource *disco.Resource) {
	//log.Printf("RESOURCE %s (%s)\n", resource.Name, resource.FullName)
	for _, method := range resource.Methods {
		addOpenAPI2PathsForMethod(d, method)
	}
	for _, resource2 := range resource.Resources {
		addOpenAPI2PathsForResource(d, resource2)
	}
}

func buildOpenAPI2DocumentForDocument(api *disco.Document) *pb.Document {
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
	return d
}
