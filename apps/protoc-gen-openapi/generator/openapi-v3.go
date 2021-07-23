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
//

package generator

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"sort"
	"strings"

	v3 "github.com/googleapis/gnostic/openapiv3"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const infoURL = "https://github.com/googleapis/gnostic/tree/master/apps/protoc-gen-openapi"

type parameterKeys struct {
	Version            string
	Serverurls         string // separated by &
	Serverdescriptions string // separated by &
}

type specialType struct {
	Name   string
	Type   string
	Format string
}

var specialTypes = []specialType{
	{
		Name:   ".google.protobuf.Timestamp",
		Type:   "string",
		Format: "RFC3339",
	},
	{
		Name:   ".google.type.Date",
		Type:   "string",
		Format: "date",
	},
	{
		Name:   ".google.type.DateTime",
		Type:   "string",
		Format: "date-time",
	},
}

type protoField struct {
	fld     *protogen.Field
	fldPath []string
}

func (f *protoField) getWholePath() string {
	return strings.Join(f.fldPath, ".")
}

// OpenAPIv3Generator holds internal state needed to generate an OpenAPIv3 document for a
//  transcoded Protocol Buffer service.
type OpenAPIv3Generator struct {
	plugin *protogen.Plugin

	requiredSchemas      []string // Names of schemas that need to be generated.
	generatedSchemas     []string // Names of schemas that have already been generated.
	linterRulePattern    *regexp.Regexp
	pathParameterPattern *regexp.Regexp
}

// getParameters returns a mapping of parameter key on parameter value
func (g *OpenAPIv3Generator) getParameters() (*parameterKeys, error) {
	parameter := g.plugin.Request.Parameter
	if parameter == nil {
		return &parameterKeys{}, nil
	}
	result := parameterKeys{}
	params := strings.Split(*parameter, ",")
	for _, param := range params {
		keyValuePair := strings.Split(param, "=")
		if len(keyValuePair) != 2 {
			return nil, fmt.Errorf("the keys and values of your parameters have to be separated by a '='")
		}
		v := reflect.ValueOf(&result).Elem()
		field := v.FieldByName(strings.ToUpper(keyValuePair[0][:1]) + keyValuePair[0][1:])
		if !field.IsValid() {
			return nil, fmt.Errorf("the parameter key %s is not known", keyValuePair[0])
		}
		field.SetString(keyValuePair[1])
	}
	return &result, nil
}

// NewOpenAPIv3Generator creates a new generator for a protoc plugin invocation.
func NewOpenAPIv3Generator(plugin *protogen.Plugin) *OpenAPIv3Generator {
	return &OpenAPIv3Generator{
		plugin:               plugin,
		requiredSchemas:      make([]string, 0),
		generatedSchemas:     make([]string, 0),
		linterRulePattern:    regexp.MustCompile(`\(-- .* --\)`),
		pathParameterPattern: regexp.MustCompile(`{([^}]+)}`),
	}
}

// Run runs the generator.
func (g *OpenAPIv3Generator) Run() error {
	d, err := g.buildDocumentV3()
	if err != nil {
		return err
	}
	bytes, err := d.YAMLValue("Generated with protoc-gen-openapi\n" + infoURL)
	if err != nil {
		return fmt.Errorf("failed to marshal yaml: %s", err.Error())
	}
	outputFile := g.plugin.NewGeneratedFile("openapi.yaml", "")
	outputFile.Write(bytes)
	return nil
}

// buildDocumentV3 builds an OpenAPIv3 document for a plugin request.
func (g *OpenAPIv3Generator) buildDocumentV3() (*v3.Document, error) {
	d := &v3.Document{}
	d.Openapi = "3.0.3"
	parameters, err := g.getParameters()
	if err != nil {
		return nil, err
	}
	var version string
	if parameters.Version == "" {
		version = "0.0.1"
	} else {
		version = parameters.Version
	}
	if parameters.Serverurls != "" {
		serverUrls := strings.Split(parameters.Serverurls, "&")
		var serverDescriptions []string
		if parameters.Serverdescriptions == "" {
			serverDescriptions = make([]string, len(serverUrls))
		} else {
			serverDescriptions = strings.Split(parameters.Serverdescriptions, "&")
			if len(serverUrls) != len(serverDescriptions) {
				return nil, fmt.Errorf("the server description count must match the server" +
					" url count or be empty. If you only want some endpoints to have descriptions you" +
					" can place an '&' without description at the index you do not want a description to appear")
			}
		}
		d.Servers = []*v3.Server{}
		for i, serverUrl := range serverUrls {
			d.Servers = append(d.Servers, &v3.Server{
				Url:         serverUrl,
				Description: serverDescriptions[i],
			})
		}
	} else if parameters.Serverdescriptions != "" {
		return nil, fmt.Errorf("you need to also pass server urls if you pass server descriptions")
	}
	d.Info = &v3.Info{
		Title:       "Definition for following service(s):",
		Version:     version,
		Description: "",
	}
	d.Paths = &v3.Paths{}
	d.Components = &v3.Components{
		Schemas: &v3.SchemasOrReferences{
			AdditionalProperties: []*v3.NamedSchemaOrReference{},
		},
	}
	var serviceNames []string
	var serviceComments []string
	for _, file := range g.plugin.Files {
		if err := g.addPathsToDocumentV3(d, file, &serviceNames, &serviceComments); err != nil {
			return nil, err
		}
	}
	d.Info.Title += " " + strings.Join(serviceNames, ", ")
	d.Info.Description += strings.Join(serviceComments, "\n")
	for len(g.requiredSchemas) > 0 {
		count := len(g.requiredSchemas)
		for _, file := range g.plugin.Files {
			// For each message, generate a definition.
			for _, message := range file.Messages {
				g.addSchemasToDocumentV3(d, message)
			}
		}
		g.requiredSchemas = g.requiredSchemas[count:len(g.requiredSchemas)]
	}
	// Sort the paths.
	{
		pairs := d.Paths.Path
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].Name < pairs[j].Name
		})
		d.Paths.Path = pairs
	}
	// Sort the schemas.
	{
		pairs := d.Components.Schemas.AdditionalProperties
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].Name < pairs[j].Name
		})
		d.Components.Schemas.AdditionalProperties = pairs
	}
	return d, nil
}

// filterCommentString removes line breaks and linter rules from comments.
func (g *OpenAPIv3Generator) filterCommentString(c protogen.Comments) string {
	comment := string(c)
	comment = strings.Replace(comment, "\n", "", -1)
	comment = g.linterRulePattern.ReplaceAllString(comment, "")
	return strings.TrimSpace(comment)
}

// addPathsToDocumentV3 adds paths from a specified file descriptor.
func (g *OpenAPIv3Generator) addPathsToDocumentV3(
	d *v3.Document,
	file *protogen.File,
	serviceNames *[]string,
	serviceComments *[]string,
) error {
	for _, service := range file.Services {
		{
			serviceName := string(service.Desc.Name())
			*serviceNames = append(*serviceNames, serviceName)
			serviceComment := g.filterCommentString(service.Comments.Leading)
			if serviceComment != "" {
				*serviceComments = append(*serviceComments, fmt.Sprintf("%s - %s", serviceName, serviceComment))
			}
		}
		for _, method := range service.Methods {
			comment := g.filterCommentString(method.Comments.Leading)
			inputMessage := method.Input
			outputMessage := method.Output
			operationID := service.GoName + "_" + method.GoName
			xt := annotations.E_Http
			extension := proto.GetExtension(method.Desc.Options(), xt)
			var path string
			var methodName string
			var body string
			if extension != nil && extension != xt.InterfaceOf(xt.Zero()) {
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
				op, path2, err := g.buildOperationV3(
					operationID, comment, path, body, inputMessage, outputMessage)
				if err != nil {
					return err
				}
				g.addOperationV3(d, op, path2, methodName)
			}
		}
	}
	return nil
}

// buildOperationV3 constructs an operation for a set of values.
func (g *OpenAPIv3Generator) buildOperationV3(
	operationID string,
	description string,
	path string,
	bodyField string,
	inputMessage *protogen.Message,
	outputMessage *protogen.Message,
) (*v3.Operation, string, error) {
	// coveredFields tracks the fields that have been used in the body or path.
	var coveredFields []protoField
	// Initialize the list of operation parameters.
	var parameters []*v3.ParameterOrReference
	// Add the path parameters to the operation parameters.
	for _, match := range g.pathParameterPattern.FindAllStringSubmatch(path, -1) {
		if matches := regexp.MustCompile(`^([^=]+)=(.+)$`).FindStringSubmatch(match[1]); matches == nil {
			pathParameter := match[1]
			field, err := getFieldForParameter(inputMessage, pathParameter)
			if err != nil {
				return nil, "", err
			}
			// create a parameter based on the field
			parameters = append(parameters,
				&v3.ParameterOrReference{
					Oneof: &v3.ParameterOrReference_Parameter{
						Parameter: &v3.Parameter{
							Name:        pathParameter,
							In:          "path",
							Required:    true,
							Description: g.filterCommentString(field.fld.Comments.Leading),
							Schema: &v3.SchemaOrReference{
								Oneof: &v3.SchemaOrReference_Schema{
									Schema: &v3.Schema{
										Type:   "string",
										Format: field.fld.Desc.Kind().String(),
									},
								},
							},
						},
					},
				})
			// add the path parameter to the covered parameters
			coveredFields = append(coveredFields, *field)
		} else {
			if coveredField, err := getFieldForParameter(inputMessage, matches[1]); err == nil {
				// Add the "name=" "name" value to the list of covered fields.
				coveredFields = append(coveredFields, *coveredField)
			} else {
				return nil, "", err
			}
			// Convert the path from the starred form to use named path parameters.
			starredPath := matches[2]
			parts := strings.Split(starredPath, "/")
			// The starred path is assumed to be in the form "things/*/otherthings/*".
			// We want to convert it to "things/{thing}/otherthings/{otherthing}".
			for i := 0; i < len(parts)-1; i += 2 {
				section := parts[i]
				pathParameter := singular(section)
				parts[i+1] = "{" + pathParameter + "}"
				parameters = append(parameters,
					&v3.ParameterOrReference{
						Oneof: &v3.ParameterOrReference_Parameter{
							Parameter: &v3.Parameter{
								Name:        pathParameter,
								In:          "path",
								Required:    true,
								Description: "The " + pathParameter + " id.",
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
			// Rewrite the path to use the path parameters.
			newPath := strings.Join(parts, "/")
			path = strings.Replace(path, match[0], newPath, 1)
		}
	}
	// Add any unhandled fields in the request message as query parameters.
	if bodyField != "*" {
		if bodyField != "" {
			if field, err := getFieldForParameter(inputMessage, bodyField); err != nil {
				return nil, "", err
			} else {
				coveredFields = append(coveredFields, *field)
			}
		}
		var fields []protoField
		for _, fld := range inputMessage.Fields {
			fields = append(fields, protoField{
				fld:     fld,
				fldPath: []string{string(fld.Desc.Name())},
			})
		}
		for i := 0; i < len(fields); i++ {
			fld := fields[i]
			if !containsField(coveredFields, &fld) {
				var schemaType string
				var schemaFormat string
				var toBeAddedAsParameter bool
				if fld.fld.Desc.Kind() == protoreflect.MessageKind {
					toBeAddedAsParameter = false
					// prevent recursive self reference of messages being parsed by adding a reference parameter
					if referencesAnyParentMessage(inputMessage, &fld) {
						descriptionComponents := []string{"You can extend the parameter's name by any parameter" +
							" in the referenced schema using a '.' for separation"}
						fieldDescription := g.filterCommentString(fld.fld.Comments.Leading)
						if fieldDescription != "" {
							descriptionComponents = append(descriptionComponents, fieldDescription)
						}
						parameters = append(parameters,
							&v3.ParameterOrReference{
								Oneof: &v3.ParameterOrReference_Parameter{
									Parameter: &v3.Parameter{
										Name:        fields[i].getWholePath(),
										In:          "query",
										Description: strings.Join(descriptionComponents, ". "),
										Required:    false,
										Schema: &v3.SchemaOrReference{
											Oneof: &v3.SchemaOrReference_Reference{
												Reference: &v3.Reference{
													XRef: g.schemaReferenceForTypeName(
														fullMessageTypeName(fld.fld.Message)),
												},
											},
										},
									},
								},
							})
					} else {
						typeName := fullMessageTypeName(fld.fld.Message)
						for _, specType := range specialTypes {
							if specType.Name == typeName {
								schemaType = specType.Type
								schemaFormat = specType.Format
								toBeAddedAsParameter = true
								break
							}
						}
						if !toBeAddedAsParameter {
							for _, f := range fld.fld.Message.Fields {
								fields = append(fields, protoField{
									fld:     f,
									fldPath: append(fields[i].fldPath, string(f.Desc.Name())),
								})
							}
						}
					}
				} else {
					toBeAddedAsParameter = true
					schemaType = "string"
					schemaFormat = fld.fld.Desc.Kind().String()
				}
				if toBeAddedAsParameter {
					// Get the field description from the comments.
					fieldDescription := g.filterCommentString(fld.fld.Comments.Leading)
					parameters = append(parameters,
						&v3.ParameterOrReference{
							Oneof: &v3.ParameterOrReference_Parameter{
								Parameter: &v3.Parameter{
									Name:        fields[i].getWholePath(),
									In:          "query",
									Description: fieldDescription,
									Required:    false,
									Schema: &v3.SchemaOrReference{
										Oneof: &v3.SchemaOrReference_Schema{
											Schema: &v3.Schema{
												Type:   schemaType,
												Format: schemaFormat,
											},
										},
									},
								},
							},
						})
				}
				coveredFields = append(coveredFields, fld)
			}
		}
	}
	// Create the response.
	responses := &v3.Responses{
		ResponseOrReference: []*v3.NamedResponseOrReference{
			&v3.NamedResponseOrReference{
				Name: "200",
				Value: &v3.ResponseOrReference{
					Oneof: &v3.ResponseOrReference_Response{
						Response: &v3.Response{
							Description: "OK",
							Content:     g.responseContentForMessage(outputMessage),
						},
					},
				},
			},
		},
	}
	// Create the operation.
	op := &v3.Operation{
		Summary:     description,
		OperationId: operationID,
		Parameters:  parameters,
		Responses:   responses,
	}
	// If a body field is specified, we need to pass a message as the request body.
	if bodyField != "" {
		var bodyFieldScalarTypeName string
		var bodyFieldMessageTypeName string
		if bodyField == "*" {
			// Pass the entire request message as the request body.
			bodyFieldMessageTypeName = fullMessageTypeName(inputMessage)
		} else {
			// If body refers to a message field, use that type.
			for _, field := range inputMessage.Fields {
				if string(field.Desc.Name()) == bodyField {
					switch field.Desc.Kind() {
					case protoreflect.StringKind:
						bodyFieldScalarTypeName = "string"
					case protoreflect.MessageKind:
						bodyFieldMessageTypeName = fullMessageTypeName(field.Message)
					default:
						log.Printf("unsupported field type %+v", field.Desc)
					}
					break
				}
			}
		}
		var requestSchema *v3.SchemaOrReference
		if bodyFieldScalarTypeName != "" {
			requestSchema = &v3.SchemaOrReference{
				Oneof: &v3.SchemaOrReference_Schema{
					Schema: &v3.Schema{
						Type: bodyFieldScalarTypeName,
					},
				},
			}
		} else if bodyFieldMessageTypeName != "" {
			requestSchema = &v3.SchemaOrReference{
				Oneof: &v3.SchemaOrReference_Reference{
					Reference: &v3.Reference{
						XRef: g.schemaReferenceForTypeName(bodyFieldMessageTypeName),
					}},
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
									Schema: requestSchema,
								},
							},
						},
					},
				},
			},
		}
	}
	return op, path, nil
}

// referencesAnyParentMessage returns if a field references one of its parent Messages. Requires the passed
//  field to be of type message
func referencesAnyParentMessage(sourceMessage *protogen.Message, field *protoField) bool {
	tmpMsg := sourceMessage.Desc
	for _, fldPath := range field.fldPath {
		if tmpMsg.FullName() == field.fld.Message.Desc.FullName() {
			return true
		}
		tmpMsg = tmpMsg.Fields().ByTextName(fldPath).Message()
	}
	return false
}

// getFieldForParameter gets a field related to the passed parameter string using sourceMessage as root
func getFieldForParameter(sourceMessage *protogen.Message, parameter string) (*protoField, error) {
	subParameters := strings.Split(parameter, ".")
	// get the field the path parameter points to
	var pField protoField
	for _, subParameter := range subParameters {
		var message *protogen.Message
		if pField.fld == nil {
			message = sourceMessage
		} else if pField.fld.Desc.Kind() == protoreflect.MessageKind {
			message = pField.fld.Message
		} else {
			return nil, fmt.Errorf(
				"only the last subparameter of a parameter is allowed to point"+
					" to a non message type (%s does not fulfil this criterium)",
				subParameter)
		}
		fieldDesc := message.Desc.Fields().ByTextName(subParameter)
		if fieldDesc == nil {
			return nil, fmt.Errorf(
				"the subparameter %s of parameter %s does not exist",
				subParameter,
				parameter)
		}
		pField.fld = message.Fields[fieldDesc.Index()]
		if pField.fld == nil {
			return nil, fmt.Errorf(
				"the parameter %s has a subparameter %s that does not exist",
				parameter,
				subParameter)
		}
		pField.fldPath = append(pField.fldPath, string(pField.fld.Desc.Name()))
	}
	if pField.fld == nil {
		return nil, fmt.Errorf("could not resolve field for %s", parameter)
	}
	return &pField, nil
}

// addOperationV3 adds an operation to the specified path/method.
func (g *OpenAPIv3Generator) addOperationV3(d *v3.Document, op *v3.Operation, path string, methodName string) {
	var selectedPathItem *v3.NamedPathItem
	for _, namedPathItem := range d.Paths.Path {
		if namedPathItem.Name == path {
			selectedPathItem = namedPathItem
			break
		}
	}
	// If we get here, we need to create a path item.
	if selectedPathItem == nil {
		selectedPathItem = &v3.NamedPathItem{Name: path, Value: &v3.PathItem{}}
		d.Paths.Path = append(d.Paths.Path, selectedPathItem)
	}
	// Set the operation on the specified method.
	switch methodName {
	case "GET":
		selectedPathItem.Value.Get = op
	case "POST":
		selectedPathItem.Value.Post = op
	case "PUT":
		selectedPathItem.Value.Put = op
	case "DELETE":
		selectedPathItem.Value.Delete = op
	case "PATCH":
		selectedPathItem.Value.Patch = op
	}
}

// schemaReferenceForTypeName returns an OpenAPI JSON Reference to the schema that represents a type.
func (g *OpenAPIv3Generator) schemaReferenceForTypeName(typeName string) string {
	if !contains(g.requiredSchemas, typeName) {
		g.requiredSchemas = append(g.requiredSchemas, typeName)
	}
	parts := strings.Split(typeName, ".")
	lastPart := parts[len(parts)-1]
	return "#/components/schemas/" + lastPart
}

// itemsItemForTypeName is a helper constructor.
func itemsItemForTypeName(typeName string) *v3.ItemsItem {
	return &v3.ItemsItem{SchemaOrReference: []*v3.SchemaOrReference{&v3.SchemaOrReference{
		Oneof: &v3.SchemaOrReference_Schema{
			Schema: &v3.Schema{
				Type: typeName}}}}}
}

// itemsItemForReference is a helper constructor.
func itemsItemForReference(xref string) *v3.ItemsItem {
	return &v3.ItemsItem{SchemaOrReference: []*v3.SchemaOrReference{&v3.SchemaOrReference{
		Oneof: &v3.SchemaOrReference_Reference{
			Reference: &v3.Reference{
				XRef: xref}}}}}
}

// fullMessageTypeName builds the full type name of a message.
func fullMessageTypeName(message *protogen.Message) string {
	return "." + string(message.Desc.ParentFile().Package()) + "." + string(message.Desc.Name())
}

func (g *OpenAPIv3Generator) responseContentForMessage(outputMessage *protogen.Message) *v3.MediaTypes {
	typeName := fullMessageTypeName(outputMessage)

	if typeName == ".google.protobuf.Empty" {
		return &v3.MediaTypes{}
	}

	if typeName == ".google.api.HttpBody" {
		return &v3.MediaTypes{
			AdditionalProperties: []*v3.NamedMediaType{
				&v3.NamedMediaType{
					Name:  "application/octet-stream",
					Value: &v3.MediaType{},
				},
			},
		}
	}

	return &v3.MediaTypes{
		AdditionalProperties: []*v3.NamedMediaType{
			&v3.NamedMediaType{
				Name: "application/json",
				Value: &v3.MediaType{
					Schema: &v3.SchemaOrReference{
						Oneof: &v3.SchemaOrReference_Reference{
							Reference: &v3.Reference{
								XRef: g.schemaReferenceForTypeName(fullMessageTypeName(outputMessage)),
							},
						},
					},
				},
			},
		},
	}
}

// addSchemasToDocumentV3 adds info from one file descriptor.
func (g *OpenAPIv3Generator) addSchemasToDocumentV3(d *v3.Document, message *protogen.Message) {
	typeName := fullMessageTypeName(message)
	// Only generate this if we need it and haven't already generated it.
	if !contains(g.requiredSchemas, typeName) ||
		contains(g.generatedSchemas, typeName) {
		return
	}
	g.generatedSchemas = append(g.generatedSchemas, typeName)
	// Get the message description from the comments.
	messageDescription := g.filterCommentString(message.Comments.Leading)
	// Build an array holding the fields of the message.
	definitionProperties := &v3.Properties{
		AdditionalProperties: make([]*v3.NamedSchemaOrReference, 0),
	}
	for _, field := range message.Fields {
		// Check the field annotations to see if this is a readonly field.
		outputOnly := false
		extension := proto.GetExtension(field.Desc.Options(), annotations.E_FieldBehavior)
		if extension != nil {
			switch v := extension.(type) {
			case []annotations.FieldBehavior:
				for _, vv := range v {
					if vv == annotations.FieldBehavior_OUTPUT_ONLY {
						outputOnly = true
					}
				}
			default:
				log.Printf("unsupported extension type %T", extension)
			}
		}
		// Get the field description from the comments.
		fieldDescription := g.filterCommentString(field.Comments.Leading)
		// The field is either described by a reference or a schema.
		XRef := ""
		fieldSchema := &v3.Schema{
			Description: fieldDescription,
		}
		if outputOnly {
			fieldSchema.ReadOnly = true
		}
		if field.Desc.IsList() {
			fieldSchema.Type = "array"
			switch field.Desc.Kind() {
			case protoreflect.MessageKind:
				fieldSchema.Items = itemsItemForReference(
					g.schemaReferenceForTypeName(
						fullMessageTypeName(field.Message)))
				g.addSchemasToDocumentV3(d, field.Message)
			case protoreflect.StringKind:
				fieldSchema.Items = itemsItemForTypeName("string")
			case protoreflect.Int32Kind,
				protoreflect.Sint32Kind,
				protoreflect.Uint32Kind,
				protoreflect.Int64Kind,
				protoreflect.Sint64Kind,
				protoreflect.Uint64Kind,
				protoreflect.Sfixed32Kind,
				protoreflect.Fixed32Kind,
				protoreflect.Sfixed64Kind,
				protoreflect.Fixed64Kind:
				fieldSchema.Items = itemsItemForTypeName("integer")
			case protoreflect.EnumKind:
				fieldSchema.Items = itemsItemForTypeName("integer")
			case protoreflect.BoolKind:
				fieldSchema.Items = itemsItemForTypeName("boolean")
			case protoreflect.FloatKind, protoreflect.DoubleKind:
				fieldSchema.Items = itemsItemForTypeName("number")
			case protoreflect.BytesKind:
				fieldSchema.Items = itemsItemForTypeName("string")
			default:
				log.Printf("(TODO) Unsupported array type: %+v", fullMessageTypeName(field.Message))
			}
		} else if field.Desc.IsMap() &&
			field.Desc.MapKey().Kind() == protoreflect.StringKind &&
			field.Desc.MapValue().Kind() == protoreflect.StringKind {
			fieldSchema.Type = "object"
		} else {
			k := field.Desc.Kind()
			switch k {
			case protoreflect.MessageKind:
				typeName := fullMessageTypeName(field.Message)
				var isSpecial bool
				for _, specType := range specialTypes {
					if specType.Name == typeName {
						fieldSchema.Type = specType.Type
						fieldSchema.Format = specType.Format
						isSpecial = true
						break
					}
				}
				if !isSpecial {
					// The field is described by a reference.
					XRef = g.schemaReferenceForTypeName(typeName)
					g.addSchemasToDocumentV3(d, field.Message)
				}
			case protoreflect.StringKind:
				fieldSchema.Type = "string"
			case protoreflect.Int32Kind,
				protoreflect.Sint32Kind,
				protoreflect.Uint32Kind,
				protoreflect.Int64Kind,
				protoreflect.Sint64Kind,
				protoreflect.Uint64Kind,
				protoreflect.Sfixed32Kind,
				protoreflect.Fixed32Kind,
				protoreflect.Sfixed64Kind,
				protoreflect.Fixed64Kind:
				fieldSchema.Type = "integer"
				fieldSchema.Format = k.String()
			case protoreflect.EnumKind:
				fieldSchema.Type = "integer"
				fieldSchema.Format = "enum"
			case protoreflect.BoolKind:
				fieldSchema.Type = "boolean"
			case protoreflect.FloatKind, protoreflect.DoubleKind:
				fieldSchema.Type = "number"
				fieldSchema.Format = k.String()
			case protoreflect.BytesKind:
				fieldSchema.Type = "string"
				fieldSchema.Format = "bytes"
			default:
				log.Printf("(TODO) Unsupported field type: %+v", fullMessageTypeName(field.Message))
			}
		}
		var value *v3.SchemaOrReference
		if XRef != "" {
			value = &v3.SchemaOrReference{
				Oneof: &v3.SchemaOrReference_Reference{
					Reference: &v3.Reference{
						XRef: XRef,
					},
				},
			}
		} else {
			value = &v3.SchemaOrReference{
				Oneof: &v3.SchemaOrReference_Schema{
					Schema: fieldSchema,
				},
			}
		}
		definitionProperties.AdditionalProperties = append(
			definitionProperties.AdditionalProperties,
			&v3.NamedSchemaOrReference{
				Name:  string(field.Desc.Name()),
				Value: value,
			},
		)
	}
	// Add the schema to the components.schema list.
	d.Components.Schemas.AdditionalProperties = append(d.Components.Schemas.AdditionalProperties,
		&v3.NamedSchemaOrReference{
			Name: string(message.Desc.Name()),
			Value: &v3.SchemaOrReference{
				Oneof: &v3.SchemaOrReference_Schema{
					Schema: &v3.Schema{
						Description: messageDescription,
						Properties:  definitionProperties,
					},
				},
			},
		},
	)
}

// containsField checks if a slice of protofields contains a value with the same whole path as the
//  passed val parameter
func containsField(store []protoField, val *protoField) bool {
	for _, field := range store {
		if field.getWholePath() == val.getWholePath() {
			return true
		}
	}
	return false
}

// contains returns true if a store (slice or array) contains a specified value.
func contains(store interface{}, val interface{}) bool {
	switch reflect.TypeOf(store).Kind() {
	case reflect.Slice,
		reflect.Array:
		s := reflect.ValueOf(store)
		for i := 0; i < s.Len(); i++ {
			if s.Index(i).Interface() == val {
				return true
			}
		}
	}
	return false
}

// singular produces the singular form of a collection name.
func singular(plural string) string {
	if strings.HasSuffix(plural, "ves") {
		return strings.TrimSuffix(plural, "ves") + "f"
	}
	if strings.HasSuffix(plural, "ies") {
		return strings.TrimSuffix(plural, "ies") + "y"
	}
	if strings.HasSuffix(plural, "s") {
		return strings.TrimSuffix(plural, "s")
	}
	return plural
}
