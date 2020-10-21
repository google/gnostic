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
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	v2 "github.com/googleapis/gnostic/openapiv2"
	"google.golang.org/genproto/googleapis/api/annotations"
)

// NewDocumentV2 creates a new OpenAPIv2 document
func (g *Generator) NewDocumentV2() *v2.Document {
	d := &v2.Document{}
	d.Swagger = "2.0"
	d.Info = &v2.Info{
		Title:   "",
		Version: "1.0.0",
		License: &v2.License{Name: "Unknown"},
	}
	d.Host = "example.com"
	d.BasePath = ""
	d.Schemes = []string{"http"}
	d.Consumes = []string{"application/json"}
	d.Produces = []string{"application/json"}
	d.Paths = &v2.Paths{}
	d.Definitions = &v2.Definitions{}
	return d
}

// AddToDocumentV2 constructs an OpenAPIv2 document for a file descriptor
func (g *Generator) AddToDocumentV2(d *v2.Document, file *FileDescriptor) {
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
				op := g.buildOperation(operationID, comment, body, inputMessage, outputMessage)
				g.addOperation(d, op, path, methodName)
			}
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

				case descriptor.FieldDescriptorProto_TYPE_STRING:
					fieldSchema.Items =
						&v2.ItemsItem{Schema: []*v2.Schema{&v2.Schema{
							Type: &v2.TypeItem{Value: []string{"string"}}}}}

				case descriptor.FieldDescriptorProto_TYPE_INT32:
					fieldSchema.Items =
						&v2.ItemsItem{Schema: []*v2.Schema{&v2.Schema{
							Type: &v2.TypeItem{Value: []string{"integer"}}}}}

				case descriptor.FieldDescriptorProto_TYPE_UINT64:
					fieldSchema.Items =
						&v2.ItemsItem{Schema: []*v2.Schema{&v2.Schema{
							Type: &v2.TypeItem{Value: []string{"integer"}}}}}

				case descriptor.FieldDescriptorProto_TYPE_ENUM:
					fieldSchema.Items =
						&v2.ItemsItem{Schema: []*v2.Schema{&v2.Schema{
							Type: &v2.TypeItem{Value: []string{"integer"}}}}}

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
				case descriptor.FieldDescriptorProto_TYPE_UINT64:
					fieldSchema.Type = &v2.TypeItem{Value: []string{"integer"}}
					fieldSchema.Format = "uint64"
				case descriptor.FieldDescriptorProto_TYPE_INT32:
					fieldSchema.Type = &v2.TypeItem{Value: []string{"integer"}}
					fieldSchema.Format = "int32"
				case descriptor.FieldDescriptorProto_TYPE_ENUM:
					fieldSchema.Type = &v2.TypeItem{Value: []string{"integer"}}
					fieldSchema.Format = "enum"
				case descriptor.FieldDescriptorProto_TYPE_BOOL:
					fieldSchema.Type = &v2.TypeItem{Value: []string{"boolean"}}
				case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
					fieldSchema.Type = &v2.TypeItem{Value: []string{"double"}}
				case descriptor.FieldDescriptorProto_TYPE_BYTES:
					fieldSchema.Type = &v2.TypeItem{Value: []string{"byte"}}
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

func (g *Generator) buildOperation(
	operationID string,
	description string,
	bodyField string,
	inputMessage *Descriptor,
	outputMessage *Descriptor,
) *v2.Operation {

	parameters := []*v2.ParametersItem{}

	if bodyField != "" {
		parameters = append(parameters,
			&v2.ParametersItem{
				Oneof: &v2.ParametersItem_Parameter{
					Parameter: &v2.Parameter{
						Oneof: &v2.Parameter_BodyParameter{
							BodyParameter: &v2.BodyParameter{
								Name:        bodyField,
								In:          "body",
								Description: "something here",
								Required:    true,
								Schema: &v2.Schema{
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
			})
	}

	return &v2.Operation{
		Summary:     description,
		OperationId: operationID,
		Tags:        []string{"tag"},
		Parameters:  parameters,
		Responses: &v2.Responses{
			ResponseCode: []*v2.NamedResponseValue{
				&v2.NamedResponseValue{
					Name: "200",
					Value: &v2.ResponseValue{
						Oneof: &v2.ResponseValue_Response{
							Response: &v2.Response{
								Schema: &v2.SchemaItem{
									Oneof: &v2.SchemaItem_Schema{
										Schema: &v2.Schema{
											XRef: "#/definitions/" + *outputMessage.Name,
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

func (g *Generator) addOperation(d *v2.Document, op *v2.Operation, path string, methodName string) {
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
	case "PATCH":
		namedPathItem.Value.Patch = op
	}
	d.Paths.Path = append(d.Paths.Path, namedPathItem)
}
