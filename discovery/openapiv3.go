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

package discovery

import (
	//"log"
	//"net/url"

	pb "github.com/googleapis/gnostic/OpenAPIv3"
)

func addOpenAPI3SchemaForSchema(d *pb.Document, name string, schema *Schema) {
	//log.Printf("SCHEMA %s\n", name)
	/*
		d.Definitions.AdditionalProperties = append(d.Definitions.AdditionalProperties,
			&pb.NamedSchema{
				Name:  schema.Name,
				Value: buildOpenAPI3SchemaForSchema(schema),
			})
	*/
}

func buildOpenAPI3SchemaForSchema(schema *Schema) *pb.Schema {
	s := &pb.Schema{}
	/*
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
			s2 := buildOpenAPI3SchemaForSchema(schema.ItemSchema)
			s.Items = &pb.ItemsItem{}
			s.Items.Schema = append(s.Items.Schema, s2)
		}
		if len(schema.Properties) > 0 {
			s.Properties = &pb.Properties{}
			for _, property := range schema.Properties {
				s.Properties.AdditionalProperties = append(s.Properties.AdditionalProperties,
					&pb.NamedSchema{
						Name:  property.Name,
						Value: buildOpenAPI3SchemaForSchema(property.Schema),
					},
				)
			}
		}
	*/
	return s
}

func buildOpenAPI3ParameterForParameter(p *Parameter) *pb.Parameter {
	//log.Printf("- PARAMETER %+v\n", p.Name)
	/*
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
								Required:    p.Required,
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
								Required:    p.Required,
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
	*/
	return nil
}

func buildOpenAPI3ResponseForSchema(schema *Schema) *pb.Response {
	/*
		//log.Printf("- RESPONSE %+v\n", schema)
		if schema == nil {
			return &pb.Response{
				Description: "Successful operation",
			}
		} else {
			ref := schema.Ref
			if ref == "" {
				log.Printf("WARNING: Unhandled response schema %+v", schema)
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
	*/
	return nil
}

func buildOpenAPI3OperationForMethod(method *Method) *pb.Operation {
	/*
		//log.Printf("METHOD %s %s %s %s\n", method.Name, method.path(), method.HTTPMethod, method.ID)
		//log.Printf("MAP %+v\n", method.JSONMap)
		parameters := make([]*pb.ParametersItem, 0)
		for _, p := range method.Parameters {
			parameters = append(parameters, &pb.ParametersItem{
				Oneof: &pb.ParametersItem_Parameter{
					Parameter: buildOpenAPI3ParameterForParameter(p),
				},
			})
		}
		responses := &pb.Responses{
			ResponseCode: []*pb.NamedResponseValue{
				&pb.NamedResponseValue{
					Name: "default",
					Value: &pb.ResponseValue{
						Oneof: &pb.ResponseValue_Response{
							Response: buildOpenAPI3ResponseForSchema(method.Response),
						},
					},
				},
			},
		}
		return &pb.Operation{
			Description: method.Description,
			OperationId: method.ID,
			Parameters:  parameters,
			Responses:   responses,
		}
	*/
	return nil
}

func getOpenAPI3PathItemForPath(d *pb.Document, path string) *pb.PathItem {
	/*
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
	*/
	return nil
}

func addOpenAPI3PathsForMethod(d *pb.Document, method *Method) {
	/*
		operation := buildOpenAPI3OperationForMethod(method)
		pathItem := getOpenAPI3PathItemForPath(d, method.path())
		switch method.HTTPMethod {
		case "GET":
			pathItem.Get = operation
		case "POST":
			pathItem.Post = operation
		case "PUT":
			pathItem.Put = operation
		case "DELETE":
			pathItem.Delete = operation
		case "PATCH":
			pathItem.Patch = operation
		default:
			log.Printf("WARNING: Unknown HTTP method %s", method.HTTPMethod)
		}
	*/
}

func addOpenAPI3PathsForResource(d *pb.Document, resource *Resource) {
	/*
		//log.Printf("RESOURCE %s (%s)\n", resource.Name, resource.FullName)
		for _, method := range resource.Methods {
			addOpenAPI3PathsForMethod(d, method)
		}
		for _, resource2 := range resource.Resources {
			addOpenAPI3PathsForResource(d, resource2)
		}
	*/
}

// OpenAPIv3 returns an OpenAPI v3 representation of this Discovery document
func (api *Document) OpenAPIv3() (*pb.Document, error) {
	d := &pb.Document{}
	d.Openapi = "3.0"
	/*
		d.Info = &pb.Info{
			Title:       api.Title,
			Version:     api.Version,
			Description: api.Description,
		}
		url, _ := url.Parse(api.RootURL)
		d.Host = url.Host
		d.BasePath = api.BasePath
		if d.BasePath == "" {
			d.BasePath = "/"
		}
		d.Schemes = []string{url.Scheme}
		d.Consumes = []string{"application/json"}
		d.Produces = []string{"application/json"}
		d.Paths = &pb.Paths{}
		d.Definitions = &pb.Definitions{}
		for name, schema := range api.Schemas {
			addOpenAPI3SchemaForSchema(d, name, schema)
		}
		for _, method := range api.Methods {
			addOpenAPI3PathsForMethod(d, method)
		}
		for _, resource := range api.Resources {
			addOpenAPI3PathsForResource(d, resource)
		}
	*/
	return d, nil
}