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

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	v2 "github.com/googleapis/gnostic/openapiv2"
	"google.golang.org/genproto/googleapis/api/annotations"
)

func (g *Generator) NewDocumentV2() *v2.Document {
	d := &v2.Document{}
	d.Swagger = "2.0"
	d.Paths = &v2.Paths{}
	d.Definitions = &v2.Definitions{}
	return d
}

// AddToDocumentV2 constructs an OpenAPIv2 document for a file descriptor
func (g *Generator) AddToDocumentV2(d *v2.Document, file *FileDescriptor) {
	g.file = file

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

	for _, service := range file.FileDescriptorProto.Service {
		log.Printf("SERVICE %s", *service.Name)
		for _, method := range service.Method {
			log.Printf("METHOD %s", *method.Name)
			log.Printf(" INPUT %s", *method.InputType)
			log.Printf(" OUTPUT %s", *method.OutputType)
			log.Printf(" OPTIONS %+v", *method.Options)
			//log.Printf(" EXTENSIONS %+v", method.Options.XXX_InternalExtensions)

			operationID := *service.Name + "_" + *method.Name

			extension, err := proto.GetExtension(method.Options, annotations.E_Http)
			log.Printf(" extensions: %T %+v (%+v)", extension, extension, err)
			var path string
			if extension != nil {
				rule := extension.(*annotations.HttpRule)
				log.Printf("  PATTERN %T %v", rule.Pattern, rule.Pattern)
				log.Printf("  SELECTOR %s", rule.Selector)
				log.Printf("  BODY %s", rule.Body)
				log.Printf("  BINDINGS %s", rule.AdditionalBindings)

				switch pattern := rule.Pattern.(type) {
				case *annotations.HttpRule_Get:
					path = pattern.Get
				case *annotations.HttpRule_Post:
					path = pattern.Post
				case *annotations.HttpRule_Put:
					path = pattern.Put
				case *annotations.HttpRule_Delete:
					path = pattern.Delete
				case *annotations.HttpRule_Patch:
					path = pattern.Patch
				case *annotations.HttpRule_Custom:
					path = "custom-unsupported"
				default:
					path = "unknown-unsupported"
				}
				log.Printf("  PATH %s", path)
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

			op := g.BuildOperation(operationID)

			g.AddOperation(d, op, path, methodName)
		}
	}

	// for each message, generate a definition
	for _, desc := range g.file.desc {
		definitionProperties := &v2.Properties{}
		for _, field := range desc.Field {
			fieldSchema := &v2.Schema{}
			if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
				fieldSchema.Type = &v2.TypeItem{Value: []string{"array"}}
				switch *field.Type {
				case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
					fieldSchema.Items =
						&v2.ItemsItem{Schema: []*v2.Schema{&v2.Schema{
							XRef: g.definitionReferenceForTypeName(*field.TypeName)}}}
				default:
					log.Printf("(TODO) Unsupported array type: %+v", field.Type)
				}
			} else {
				switch *field.Type {
				case descriptor.FieldDescriptorProto_TYPE_STRING:
					fieldSchema.Type = &v2.TypeItem{Value: []string{"string"}}
				case descriptor.FieldDescriptorProto_TYPE_INT64:
					fieldSchema.Type = &v2.TypeItem{Value: []string{"integer"}}
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
}

func (g *Generator) definitionReferenceForTypeName(typeName string) string {
	parts := strings.Split(typeName, ".")
	lastPart := parts[len(parts)-1]
	return "#/definitions/" + lastPart
}

func (g *Generator) BuildOperation(operationID string) *v2.Operation {
	return &v2.Operation{
		Summary:     "Summary",
		OperationId: operationID,
		Tags:        []string{"tag"},
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
	}
}

func (g *Generator) AddOperation(d *v2.Document, op *v2.Operation, path string, methodName string) {
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
			}
			return
		}
	}
	// if we get here, we need to create a path item
	namedPathItem := &v2.NamedPathItem{Name: path, Value: &v2.PathItem{}}
	switch methodName {
	case "GET":
		namedPathItem.Value.Get = op
	case "POST":
		namedPathItem.Value.Post = op
	case "PUT":
		namedPathItem.Value.Put = op
	case "DELETE":
		namedPathItem.Value.Delete = op
	}
	d.Paths.Path = append(d.Paths.Path, namedPathItem)
}
