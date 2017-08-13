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
	"log"
	"net/url"

	pb "github.com/googleapis/gnostic/OpenAPIv3"
)

func addOpenAPI3SchemaForSchema(d *pb.Document, name string, schema *Schema) {
	d.Components.Schemas.AdditionalProperties = append(d.Components.Schemas.AdditionalProperties,
		&pb.NamedSchemaOrReference{
			Name:  schema.Name,
			Value: buildOpenAPI3SchemaOrReferenceForSchema(schema),
		})
}

func buildOpenAPI3SchemaOrReferenceForSchema(schema *Schema) *pb.SchemaOrReference {
	if ref := schema.Ref; ref != "" {
		return &pb.SchemaOrReference{
			Oneof: &pb.SchemaOrReference_Reference{
				Reference: &pb.Reference{
					XRef: "#/definitions/" + ref,
				},
			},
		}
	}

	s := &pb.Schema{}

	if description := schema.Description; description != "" {
		s.Description = description
	}
	if typeName := schema.Type; typeName != "" {
		s.Type = typeName
	}
	if len(schema.Enums) > 0 {
		for _, e := range schema.Enums {
			s.Enum = append(s.Enum, &pb.Any{Yaml: e})
		}
	}
	if schema.ItemSchema != nil {
		s2 := buildOpenAPI3SchemaOrReferenceForSchema(schema.ItemSchema)
		s.Items = &pb.ItemsItem{}
		s.Items.SchemaOrReference = append(s.Items.SchemaOrReference, s2)
	}
	if len(schema.Properties) > 0 {
		s.Properties = &pb.Properties{}
		for _, property := range schema.Properties {
			s.Properties.AdditionalProperties = append(s.Properties.AdditionalProperties,
				&pb.NamedSchemaOrReference{
					Name:  property.Name,
					Value: buildOpenAPI3SchemaOrReferenceForSchema(property.Schema),
				},
			)
		}
	}
	return &pb.SchemaOrReference{
		Oneof: &pb.SchemaOrReference_Schema{
			Schema: s,
		},
	}
}

func buildOpenAPI3ParameterForParameter(p *Parameter) *pb.Parameter {
	typeName := p.Schema.Type
	format := p.Schema.Format
	location := p.Location
	switch location {
	case "query", "path":
		return &pb.Parameter{
			Name:        p.Name,
			In:          location,
			Description: p.Description,
			Required:    p.Required,
			Schema: &pb.SchemaOrReference{
				Oneof: &pb.SchemaOrReference_Schema{
					Schema: &pb.Schema{
						Type:   typeName,
						Format: format,
					},
				},
			},
		}
	default:
		return nil
	}
}

func buildOpenAPI3RequestBodyForRequest(schema *Schema) *pb.RequestBody {
	ref := schema.Ref
	if ref == "" {
		log.Printf("WARNING: Unhandled request schema %+v", schema)
	}
	return &pb.RequestBody{
		Content: &pb.MediaTypes{
			AdditionalProperties: []*pb.NamedMediaType{
				&pb.NamedMediaType{
					Name: "application/json",
					Value: &pb.MediaType{
						Schema: &pb.SchemaOrReference{
							Oneof: &pb.SchemaOrReference_Reference{
								Reference: &pb.Reference{
									XRef: "#/definitions/" + ref,
								},
							},
						},
					},
				},
			},
		},
	}
}

func buildOpenAPI3ResponseForSchema(schema *Schema, hasDataWrapper bool) *pb.Response {
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
			Content: &pb.MediaTypes{
				AdditionalProperties: []*pb.NamedMediaType{
					&pb.NamedMediaType{
						Name: "application/json",
						Value: &pb.MediaType{
							Schema: &pb.SchemaOrReference{
								Oneof: &pb.SchemaOrReference_Reference{
									Reference: &pb.Reference{
										XRef: "#/definitions/" + ref,
									},
								},
							},
						},
					},
				},
			},
		}
	}
}

func buildOpenAPI3OperationForMethod(method *Method, hasDataWrapper bool) *pb.Operation {
	parameters := make([]*pb.ParameterOrReference, 0)
	for _, p := range method.Parameters {
		parameters = append(parameters, &pb.ParameterOrReference{
			Oneof: &pb.ParameterOrReference_Parameter{
				Parameter: buildOpenAPI3ParameterForParameter(p),
			},
		})
	}
	responses := &pb.Responses{
		ResponseOrReference: []*pb.NamedResponseOrReference{
			&pb.NamedResponseOrReference{
				Name: "default",
				Value: &pb.ResponseOrReference{
					Oneof: &pb.ResponseOrReference_Response{
						Response: buildOpenAPI3ResponseForSchema(method.Response, hasDataWrapper),
					},
				},
			},
		},
	}
	var requestBodyOrReference *pb.RequestBodyOrReference
	if method.Request != nil {
		requestBody := buildOpenAPI3RequestBodyForRequest(method.Request)
		requestBodyOrReference = &pb.RequestBodyOrReference{
			Oneof: &pb.RequestBodyOrReference_RequestBody{
				RequestBody: requestBody,
			},
		}
	}
	return &pb.Operation{
		Description: method.Description,
		OperationId: method.ID,
		Parameters:  parameters,
		Responses:   responses,
		RequestBody: requestBodyOrReference,
	}
}

func getOpenAPI3PathItemForPath(d *pb.Document, path string) *pb.PathItem {
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

func addOpenAPI3PathsForMethod(d *pb.Document, method *Method, hasDataWrapper bool) {
	operation := buildOpenAPI3OperationForMethod(method, hasDataWrapper)
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
}

func addOpenAPI3PathsForResource(d *pb.Document, resource *Resource, hasDataWrapper bool) {
	for _, method := range resource.Methods {
		addOpenAPI3PathsForMethod(d, method, hasDataWrapper)
	}
	for _, resource2 := range resource.Resources {
		addOpenAPI3PathsForResource(d, resource2, hasDataWrapper)
	}
}

// OpenAPIv3 returns an OpenAPI v3 representation of a Discovery document
func (api *Document) OpenAPIv3() (*pb.Document, error) {
	d := &pb.Document{}
	d.Openapi = "3.0"
	d.Info = &pb.Info{
		Title:       api.Title,
		Version:     api.Version,
		Description: api.Description,
	}

	d.Servers = make([]*pb.Server, 0)

	hasDataWrapper := false
	for _, feature := range api.Features {
		if feature == "dataWrapper" {
			hasDataWrapper = true
		}
	}

	url, _ := url.Parse(api.RootURL)
	host := url.Host
	basePath := api.BasePath
	if basePath == "" {
		basePath = "/"
	}
	d.Servers = append(d.Servers, &pb.Server{Url: "https://" + host + basePath})

	d.Components = &pb.Components{}
	d.Components.Schemas = &pb.SchemasOrReferences{}
	for name, schema := range api.Schemas {
		addOpenAPI3SchemaForSchema(d, name, schema)
	}

	d.Paths = &pb.Paths{}
	for _, method := range api.Methods {
		addOpenAPI3PathsForMethod(d, method, hasDataWrapper)
	}
	for _, resource := range api.Resources {
		addOpenAPI3PathsForResource(d, resource, hasDataWrapper)
	}
	return d, nil
}
