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
	"net/url"
	"regexp"
	"sort"
	"strings"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	v3 "github.com/google/gnostic/openapiv3"
)

type Configuration struct {
	Version       *string
	Title         *string
	Description   *string
	Naming        *string
	CircularDepth *int
}

const (
	infoURL           = "https://github.com/google/gnostic/tree/master/apps/protoc-gen-openapi"
	protobufValueName = "AnyJSONValue"
)

// OpenAPIv3Generator holds internal state needed to generate an OpenAPIv3 document for a transcoded Protocol Buffer service.
type OpenAPIv3Generator struct {
	conf   Configuration
	plugin *protogen.Plugin

	requiredSchemas   []string // Names of schemas that need to be generated.
	generatedSchemas  []string // Names of schemas that have already been generated.
	linterRulePattern *regexp.Regexp
	pathPattern       *regexp.Regexp
	namedPathPattern  *regexp.Regexp
}

// NewOpenAPIv3Generator creates a new generator for a protoc plugin invocation.
func NewOpenAPIv3Generator(plugin *protogen.Plugin, conf Configuration) *OpenAPIv3Generator {
	return &OpenAPIv3Generator{
		conf:   conf,
		plugin: plugin,

		requiredSchemas:   make([]string, 0),
		generatedSchemas:  make([]string, 0),
		linterRulePattern: regexp.MustCompile(`\(-- .* --\)`),
		pathPattern:       regexp.MustCompile("{([^=}]+)}"),
		namedPathPattern:  regexp.MustCompile("{(.+)=(.+)}"),
	}
}

// Run runs the generator.
func (g *OpenAPIv3Generator) Run() error {
	d := g.buildDocumentV3()
	bytes, err := d.YAMLValue("Generated with protoc-gen-openapi\n" + infoURL)
	if err != nil {
		return fmt.Errorf("failed to marshal yaml: %s", err.Error())
	}
	outputFile := g.plugin.NewGeneratedFile("openapi.yaml", "")
	outputFile.Write(bytes)
	return nil
}

// buildDocumentV3 builds an OpenAPIv3 document for a plugin request.
func (g *OpenAPIv3Generator) buildDocumentV3() *v3.Document {
	d := &v3.Document{}

	d.Openapi = "3.0.3"
	d.Info = &v3.Info{
		Version:     *g.conf.Version,
		Title:       *g.conf.Title,
		Description: *g.conf.Description,
	}

	d.Paths = &v3.Paths{}
	d.Components = &v3.Components{
		Schemas: &v3.SchemasOrReferences{
			AdditionalProperties: []*v3.NamedSchemaOrReference{},
		},
	}

	for _, file := range g.plugin.Files {
		if file.Generate {
			g.addPathsToDocumentV3(d, file)
		}
	}

	// If there is only 1 service, then use it's title for the document,
	//  if the document is missing it.
	if len(d.Tags) == 1 {
		if d.Info.Title == "" && d.Tags[0].Name != "" {
			d.Info.Title = d.Tags[0].Name + " API"
		}
		if d.Info.Description == "" {
			d.Info.Description = d.Tags[0].Description
		}
		d.Tags[0].Description = ""
	}

	for len(g.requiredSchemas) > 0 {
		count := len(g.requiredSchemas)
		for _, file := range g.plugin.Files {
			g.addSchemasToDocumentV3(d, file.Messages)
		}
		g.requiredSchemas = g.requiredSchemas[count:len(g.requiredSchemas)]
	}

	allServers := []string{}

	// If paths methods has servers, but they're all the same, then move servers to path level
	for _, path := range d.Paths.Path {
		servers := []string{}
		// Only 1 server will ever be set, per method, by the generator

		if path.Value.Get != nil && len(path.Value.Get.Servers) == 1 {
			servers = appendUniuqe(servers, path.Value.Get.Servers[0].Url)
			allServers = appendUniuqe(servers, path.Value.Get.Servers[0].Url)
		}
		if path.Value.Post != nil && len(path.Value.Post.Servers) == 1 {
			servers = appendUniuqe(servers, path.Value.Post.Servers[0].Url)
			allServers = appendUniuqe(servers, path.Value.Post.Servers[0].Url)
		}
		if path.Value.Put != nil && len(path.Value.Put.Servers) == 1 {
			servers = appendUniuqe(servers, path.Value.Put.Servers[0].Url)
			allServers = appendUniuqe(servers, path.Value.Put.Servers[0].Url)
		}
		if path.Value.Delete != nil && len(path.Value.Delete.Servers) == 1 {
			servers = appendUniuqe(servers, path.Value.Delete.Servers[0].Url)
			allServers = appendUniuqe(servers, path.Value.Delete.Servers[0].Url)
		}
		if path.Value.Patch != nil && len(path.Value.Patch.Servers) == 1 {
			servers = appendUniuqe(servers, path.Value.Patch.Servers[0].Url)
			allServers = appendUniuqe(servers, path.Value.Patch.Servers[0].Url)
		}

		if len(servers) == 1 {
			path.Value.Servers = []*v3.Server{{Url: servers[0]}}

			if path.Value.Get != nil {
				path.Value.Get.Servers = nil
			}
			if path.Value.Post != nil {
				path.Value.Post.Servers = nil
			}
			if path.Value.Put != nil {
				path.Value.Put.Servers = nil
			}
			if path.Value.Delete != nil {
				path.Value.Delete.Servers = nil
			}
			if path.Value.Patch != nil {
				path.Value.Patch.Servers = nil
			}
		}
	}

	// Set all servers on API level
	if len(allServers) > 0 {
		d.Servers = []*v3.Server{}
		for _, server := range allServers {
			d.Servers = append(d.Servers, &v3.Server{Url: server})
		}
	}

	// If there is only 1 server, we can safely remove all path level servers
	if len(allServers) == 1 {
		for _, path := range d.Paths.Path {
			path.Value.Servers = nil
		}
	}

	// Sort the tags.
	{
		pairs := d.Tags
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].Name < pairs[j].Name
		})
		d.Tags = pairs
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
	return d
}

// filterCommentString removes line breaks and linter rules from comments.
func (g *OpenAPIv3Generator) filterCommentString(c protogen.Comments, removeNewLines bool) string {
	comment := string(c)
	if removeNewLines {
		comment = strings.Replace(comment, "\n", "", -1)
	}
	comment = g.linterRulePattern.ReplaceAllString(comment, "")
	return strings.TrimSpace(comment)
}

// addPathsToDocumentV3 adds paths from a specified file descriptor.
func (g *OpenAPIv3Generator) addPathsToDocumentV3(d *v3.Document, file *protogen.File) {
	for _, service := range file.Services {
		annotationsCount := 0

		for _, method := range service.Methods {
			comment := g.filterCommentString(method.Comments.Leading, false)
			inputMessage := method.Input
			outputMessage := method.Output
			operationID := service.GoName + "_" + method.GoName

			var path string
			var methodName string
			var body string

			extHTTP := proto.GetExtension(method.Desc.Options(), annotations.E_Http)
			if extHTTP != nil && extHTTP != annotations.E_Http.InterfaceOf(annotations.E_Http.Zero()) {
				annotationsCount++

				rule := extHTTP.(*annotations.HttpRule)
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
				defaultHost := proto.GetExtension(service.Desc.Options(), annotations.E_DefaultHost).(string)

				op, path2 := g.buildOperationV3(
					file, operationID, service.GoName, comment, defaultHost, path, body, inputMessage, outputMessage)
				g.addOperationV3(d, op, path2, methodName)
			}
		}

		if annotationsCount > 0 {
			comment := g.filterCommentString(service.Comments.Leading, false)
			d.Tags = append(d.Tags, &v3.Tag{Name: service.GoName, Description: comment})
		}
	}
}

func (g *OpenAPIv3Generator) formatMessageRef(name string) string {
	if *g.conf.Naming == "proto" {
		return name
	}

	if len(name) > 1 {
		return strings.ToUpper(name[0:1]) + name[1:]
	}

	if len(name) == 1 {
		return strings.ToLower(name)
	}

	return name
}

func getMessageName(message protoreflect.MessageDescriptor) string {
	prefix := ""
	parent := message.Parent()
	if message != nil {
		if _, ok := parent.(protoreflect.MessageDescriptor); ok {
			prefix = string(parent.Name()) + "_" + prefix
		}
	}

	return prefix + string(message.Name())
}

func (g *OpenAPIv3Generator) formatMessageName(message *protogen.Message) string {
	name := getMessageName(message.Desc)

	if *g.conf.Naming == "proto" {
		return name
	}

	if len(name) > 0 {
		return strings.ToUpper(name[0:1]) + name[1:]
	}

	return name
}

func (g *OpenAPIv3Generator) formatFieldName(field *protogen.Field) string {
	if *g.conf.Naming == "proto" {
		return string(field.Desc.Name())
	}

	return field.Desc.JSONName()
}

func (g *OpenAPIv3Generator) findField(name string, inMessage *protogen.Message) *protogen.Field {
	for _, field := range inMessage.Fields {
		if string(field.Desc.Name()) == name || string(field.Desc.JSONName()) == name {
			return field
		}
	}

	return nil
}

func (g *OpenAPIv3Generator) findAndFormatFieldName(name string, inMessage *protogen.Message) string {
	field := g.findField(name, inMessage)
	if field != nil {
		return g.formatFieldName(field)
	}

	return name
}

// Note that fields which are mapped to URL query parameters must have a primitive type
// or a repeated primitive type or a non-repeated message type.
// In the case of a repeated type, the parameter can be repeated in the URL as ...?param=A&param=B.
// In the case of a message type, each field of the message is mapped to a separate parameter,
// such as ...?foo.a=A&foo.b=B&foo.c=C.
//
// maps, Struct and Empty can NOT be used
// messages can have any number of sub messages - including circular (e.g. sub.subsub.sub.subsub.id)

// buildQueryParamsV3 extracts any valid query params, including sub and recursive messages
func (g *OpenAPIv3Generator) buildQueryParamsV3(field *protogen.Field) []*v3.ParameterOrReference {
	depths := map[string]int{}
	return g._buildQueryParamsV3(field, depths)
}

// depths are used to keep track of how many times a message's fields has been seen
func (g *OpenAPIv3Generator) _buildQueryParamsV3(field *protogen.Field, depths map[string]int) []*v3.ParameterOrReference {
	parameters := []*v3.ParameterOrReference{}

	queryFieldName := g.formatFieldName(field)
	fieldDescription := g.filterCommentString(field.Comments.Leading, true)

	if field.Desc.IsMap() {
		// Map types are not allowed in query parameteres
		return parameters

	} else if field.Desc.Kind() == protoreflect.MessageKind {
		// Represent google.protobuf.Value as reference to the value of const protobufValueName.
		if fullMessageTypeName(field.Desc.Message()) == ".google.protobuf.Value" {
			fieldSchema := g.schemaOrReferenceForField(field.Desc)
			parameters = append(parameters,
				&v3.ParameterOrReference{
					Oneof: &v3.ParameterOrReference_Parameter{
						Parameter: &v3.Parameter{
							Name:        queryFieldName,
							In:          "query",
							Description: fieldDescription,
							Required:    false,
							Schema:      fieldSchema,
						},
					},
				})
			return parameters
		} else if field.Desc.IsList() {
			// Only non-repeated message types are valid
			return parameters
		}

		// Represent field masks directly as strings (don't expand them).
		if fullMessageTypeName(field.Desc.Message()) == ".google.protobuf.FieldMask" {
			fieldSchema := g.schemaOrReferenceForField(field.Desc)
			parameters = append(parameters,
				&v3.ParameterOrReference{
					Oneof: &v3.ParameterOrReference_Parameter{
						Parameter: &v3.Parameter{
							Name:        queryFieldName,
							In:          "query",
							Description: fieldDescription,
							Required:    false,
							Schema:      fieldSchema,
						},
					},
				})
			return parameters
		}

		// Sub messages are allowed, even circular, as long as the final type is a primitive.
		// Go through each of the sub message fields
		for _, subField := range field.Message.Fields {
			subFieldFullName := string(subField.Desc.FullName())
			seen, ok := depths[subFieldFullName]
			if !ok {
				depths[subFieldFullName] = 0
			}

			if seen < *g.conf.CircularDepth {
				depths[subFieldFullName]++
				subParams := g._buildQueryParamsV3(subField, depths)
				for _, subParam := range subParams {
					if param, ok := subParam.Oneof.(*v3.ParameterOrReference_Parameter); ok {
						param.Parameter.Name = queryFieldName + "." + param.Parameter.Name
						parameters = append(parameters, subParam)
					}
				}
			}
		}

	} else if field.Desc.Kind() != protoreflect.GroupKind {
		// schemaOrReferenceForField also handles array types
		fieldSchema := g.schemaOrReferenceForField(field.Desc)

		parameters = append(parameters,
			&v3.ParameterOrReference{
				Oneof: &v3.ParameterOrReference_Parameter{
					Parameter: &v3.Parameter{
						Name:        queryFieldName,
						In:          "query",
						Description: fieldDescription,
						Required:    false,
						Schema:      fieldSchema,
					},
				},
			})
	}

	return parameters
}

// buildOperationV3 constructs an operation for a set of values.
func (g *OpenAPIv3Generator) buildOperationV3(
	file *protogen.File,
	operationID string,
	tagName string,
	description string,
	defaultHost string,
	path string,
	bodyField string,
	inputMessage *protogen.Message,
	outputMessage *protogen.Message,
) (*v3.Operation, string) {
	// coveredParameters tracks the parameters that have been used in the body or path.
	coveredParameters := make([]string, 0)
	if bodyField != "" {
		coveredParameters = append(coveredParameters, bodyField)
	}
	// Initialize the list of operation parameters.
	parameters := []*v3.ParameterOrReference{}

	// Find simple path parameters like {id}
	if allMatches := g.pathPattern.FindAllStringSubmatch(path, -1); allMatches != nil {
		for _, matches := range allMatches {
			// Add the value to the list of covered parameters.
			coveredParameters = append(coveredParameters, matches[1])
			pathParameter := g.findAndFormatFieldName(matches[1], inputMessage)
			path = strings.Replace(path, matches[1], pathParameter, 1)

			// Add the path parameters to the operation parameters.
			var fieldSchema *v3.SchemaOrReference

			var fieldDescription string
			field := g.findField(pathParameter, inputMessage)
			if field != nil {
				fieldSchema = g.schemaOrReferenceForField(field.Desc)
				fieldDescription = g.filterCommentString(field.Comments.Leading, true)
			} else {
				// If field dooes not exist, it is safe to set it to string, as it is ignored downstream
				fieldSchema = &v3.SchemaOrReference{
					Oneof: &v3.SchemaOrReference_Schema{
						Schema: &v3.Schema{
							Type: "string",
						},
					},
				}
			}

			parameters = append(parameters,
				&v3.ParameterOrReference{
					Oneof: &v3.ParameterOrReference_Parameter{
						Parameter: &v3.Parameter{
							Name:        pathParameter,
							In:          "path",
							Description: fieldDescription,
							Required:    true,
							Schema:      fieldSchema,
						},
					},
				})
		}
	}

	// Find named path parameters like {name=shelves/*}
	if matches := g.namedPathPattern.FindStringSubmatch(path); matches != nil {
		// Build a list of named path parameters.
		namedPathParameters := make([]string, 0)

		// Add the "name=" "name" value to the list of covered parameters.
		coveredParameters = append(coveredParameters, matches[1])
		// Convert the path from the starred form to use named path parameters.
		starredPath := matches[2]
		parts := strings.Split(starredPath, "/")
		// The starred path is assumed to be in the form "things/*/otherthings/*".
		// We want to convert it to "things/{thingsId}/otherthings/{otherthingsId}".
		for i := 0; i < len(parts)-1; i += 2 {
			section := parts[i]
			namedPathParameter := g.findAndFormatFieldName(section, inputMessage)
			namedPathParameter = singular(namedPathParameter)
			parts[i+1] = "{" + namedPathParameter + "}"
			namedPathParameters = append(namedPathParameters, namedPathParameter)
		}
		// Rewrite the path to use the path parameters.
		newPath := strings.Join(parts, "/")
		path = strings.Replace(path, matches[0], newPath, 1)

		// Add the named path parameters to the operation parameters.
		for _, namedPathParameter := range namedPathParameters {
			parameters = append(parameters,
				&v3.ParameterOrReference{
					Oneof: &v3.ParameterOrReference_Parameter{
						Parameter: &v3.Parameter{
							Name:        namedPathParameter,
							In:          "path",
							Required:    true,
							Description: "The " + namedPathParameter + " id.",
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
	}

	// Add any unhandled fields in the request message as query parameters.
	if bodyField != "*" {
		for _, field := range inputMessage.Fields {
			fieldName := string(field.Desc.Name())
			if !contains(coveredParameters, fieldName) && fieldName != bodyField {
				fieldParams := g.buildQueryParamsV3(field)
				parameters = append(parameters, fieldParams...)
			}
		}
	}

	// Create the response.
	responses := &v3.Responses{
		ResponseOrReference: []*v3.NamedResponseOrReference{
			{
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
		Tags:        []string{tagName},
		Description: description,
		OperationId: operationID,
		Parameters:  parameters,
		Responses:   responses,
	}

	if defaultHost != "" {
		hostURL, err := url.Parse(defaultHost)
		if err == nil {
			hostURL.Scheme = "https"
			op.Servers = append(op.Servers, &v3.Server{Url: hostURL.String()})
		}
	}

	// If a body field is specified, we need to pass a message as the request body.
	if bodyField != "" {
		var requestSchema *v3.SchemaOrReference

		if bodyField == "*" {
			// Pass the entire request message as the request body.
			typeName := fullMessageTypeName(inputMessage.Desc)
			requestSchema = g.schemaOrReferenceForType(typeName)

		} else {
			// If body refers to a message field, use that type.
			for _, field := range inputMessage.Fields {
				if string(field.Desc.Name()) == bodyField {
					switch field.Desc.Kind() {
					case protoreflect.StringKind:
						requestSchema = &v3.SchemaOrReference{
							Oneof: &v3.SchemaOrReference_Schema{
								Schema: &v3.Schema{
									Type: "string",
								},
							},
						}

					case protoreflect.MessageKind:
						typeName := fullMessageTypeName(field.Message.Desc)
						requestSchema = g.schemaOrReferenceForType(typeName)

					default:
						log.Printf("unsupported field type %+v", field.Desc)
					}
					break
				}
			}
		}

		op.RequestBody = &v3.RequestBodyOrReference{
			Oneof: &v3.RequestBodyOrReference_RequestBody{
				RequestBody: &v3.RequestBody{
					Required: true,
					Content: &v3.MediaTypes{
						AdditionalProperties: []*v3.NamedMediaType{
							{
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
	return op, path
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

	if typeName == ".google.protobuf.Value" {
		return "#/components/schemas/" + protobufValueName
	}

	parts := strings.Split(typeName, ".")
	lastPart := parts[len(parts)-1]
	return "#/components/schemas/" + g.formatMessageRef(lastPart)
}

// fullMessageTypeName builds the full type name of a message.
func fullMessageTypeName(message protoreflect.MessageDescriptor) string {
	name := getMessageName(message)
	return "." + string(message.ParentFile().Package()) + "." + name
}

func (g *OpenAPIv3Generator) responseContentForMessage(outputMessage *protogen.Message) *v3.MediaTypes {
	typeName := fullMessageTypeName(outputMessage.Desc)

	if typeName == ".google.protobuf.Empty" {
		return &v3.MediaTypes{}
	}

	if typeName == ".google.api.HttpBody" {
		return &v3.MediaTypes{
			AdditionalProperties: []*v3.NamedMediaType{
				{
					Name:  "application/octet-stream",
					Value: &v3.MediaType{},
				},
			},
		}
	}

	return &v3.MediaTypes{
		AdditionalProperties: []*v3.NamedMediaType{
			{
				Name: "application/json",
				Value: &v3.MediaType{
					Schema: g.schemaOrReferenceForType(typeName),
				},
			},
		},
	}
}

func (g *OpenAPIv3Generator) schemaOrReferenceForType(typeName string) *v3.SchemaOrReference {
	switch typeName {

	case ".google.protobuf.Timestamp":
		// Timestamps are serialized as strings
		return &v3.SchemaOrReference{
			Oneof: &v3.SchemaOrReference_Schema{
				Schema: &v3.Schema{Type: "string", Format: "RFC3339"}}}

	case ".google.type.Date":
		// Dates are serialized as strings
		return &v3.SchemaOrReference{
			Oneof: &v3.SchemaOrReference_Schema{
				Schema: &v3.Schema{Type: "string", Format: "date"}}}

	case ".google.type.DateTime":
		// DateTimes are serialized as strings
		return &v3.SchemaOrReference{
			Oneof: &v3.SchemaOrReference_Schema{
				Schema: &v3.Schema{Type: "string", Format: "date-time"}}}

	case ".google.protobuf.FieldMask":
		// Field masks are serialized as strings
		return &v3.SchemaOrReference{
			Oneof: &v3.SchemaOrReference_Schema{
				Schema: &v3.Schema{Type: "string", Format: "field-mask"}}}

	case ".google.protobuf.Struct":
		// Struct is equivalent to a JSON object
		return &v3.SchemaOrReference{
			Oneof: &v3.SchemaOrReference_Schema{
				Schema: &v3.Schema{Type: "object"}}}

	case ".google.protobuf.Empty":
		// Empty is close to JSON undefined than null, so ignore this field
		return nil //&v3.SchemaOrReference{Oneof: &v3.SchemaOrReference_Schema{Schema: &v3.Schema{Type: "null"}}}

	default:
		ref := g.schemaReferenceForTypeName(typeName)
		return &v3.SchemaOrReference{
			Oneof: &v3.SchemaOrReference_Reference{
				Reference: &v3.Reference{XRef: ref}}}
	}
}

func (g *OpenAPIv3Generator) schemaOrReferenceForField(field protoreflect.FieldDescriptor) *v3.SchemaOrReference {
	if field.IsMap() {
		return &v3.SchemaOrReference{
			Oneof: &v3.SchemaOrReference_Schema{
				Schema: &v3.Schema{Type: "object",
					AdditionalProperties: &v3.AdditionalPropertiesItem{
						Oneof: &v3.AdditionalPropertiesItem_SchemaOrReference{
							SchemaOrReference: g.schemaOrReferenceForField(field.MapValue())}}}}}
	}

	var kindSchema *v3.SchemaOrReference

	kind := field.Kind()

	switch kind {

	case protoreflect.MessageKind:
		typeName := fullMessageTypeName(field.Message())
		kindSchema = g.schemaOrReferenceForType(typeName)
		if kindSchema == nil {
			return nil
		}

	case protoreflect.StringKind:
		kindSchema = &v3.SchemaOrReference{
			Oneof: &v3.SchemaOrReference_Schema{
				Schema: &v3.Schema{Type: "string"}}}

	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Uint32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Uint64Kind,
		protoreflect.Sfixed32Kind, protoreflect.Fixed32Kind, protoreflect.Sfixed64Kind,
		protoreflect.Fixed64Kind:
		kindSchema = &v3.SchemaOrReference{
			Oneof: &v3.SchemaOrReference_Schema{
				Schema: &v3.Schema{Type: "integer", Format: kind.String()}}}

	case protoreflect.EnumKind:
		kindSchema = &v3.SchemaOrReference{
			Oneof: &v3.SchemaOrReference_Schema{
				Schema: &v3.Schema{Type: "integer", Format: "enum"}}}

	case protoreflect.BoolKind:
		kindSchema = &v3.SchemaOrReference{
			Oneof: &v3.SchemaOrReference_Schema{
				Schema: &v3.Schema{Type: "boolean"}}}

	case protoreflect.FloatKind, protoreflect.DoubleKind:
		kindSchema = &v3.SchemaOrReference{
			Oneof: &v3.SchemaOrReference_Schema{
				Schema: &v3.Schema{Type: "number", Format: kind.String()}}}

	case protoreflect.BytesKind:
		kindSchema = &v3.SchemaOrReference{
			Oneof: &v3.SchemaOrReference_Schema{
				Schema: &v3.Schema{Type: "string", Format: "bytes"}}}

	default:
		log.Printf("(TODO) Unsupported field type: %+v", fullMessageTypeName(field.Message()))
	}

	if field.IsList() {
		return &v3.SchemaOrReference{
			Oneof: &v3.SchemaOrReference_Schema{
				Schema: &v3.Schema{
					Type:  "array",
					Items: &v3.ItemsItem{SchemaOrReference: []*v3.SchemaOrReference{kindSchema}},
				},
			},
		}
	}

	return kindSchema
}

// addSchemasToDocumentV3 adds info from one file descriptor.
func (g *OpenAPIv3Generator) addSchemasToDocumentV3(d *v3.Document, messages []*protogen.Message) {
	// For each message, generate a definition.
	for _, message := range messages {
		if message.Messages != nil {
			g.addSchemasToDocumentV3(d, message.Messages)
		}

		typeName := fullMessageTypeName(message.Desc)

		// Only generate this if we need it and haven't already generated it.
		if !contains(g.requiredSchemas, typeName) ||
			contains(g.generatedSchemas, typeName) {
			continue
		}

		g.generatedSchemas = append(g.generatedSchemas, typeName)

		// google.protobuf.Value is handled like a special value when doing transcoding.
		// It's interpreted as a "catch all" JSON value, that can be anything.
		if message.Desc != nil && message.Desc.FullName() == "google.protobuf.Value" {
			// Add the schema to the components.schema list.
			description := protobufValueName + ` is a "catch all" type that can hold any JSON value, except null as this is not allowed in OpenAPI`

			d.Components.Schemas.AdditionalProperties = append(d.Components.Schemas.AdditionalProperties,
				&v3.NamedSchemaOrReference{
					Name: protobufValueName,
					Value: &v3.SchemaOrReference{
						Oneof: &v3.SchemaOrReference_Schema{
							Schema: &v3.Schema{
								Description: description,
								OneOf: []*v3.SchemaOrReference{
									// type is not allow to be null in OpenAPI
									{
										Oneof: &v3.SchemaOrReference_Schema{
											Schema: &v3.Schema{Type: "string"},
										},
									}, {
										Oneof: &v3.SchemaOrReference_Schema{
											Schema: &v3.Schema{Type: "number"},
										},
									}, {
										Oneof: &v3.SchemaOrReference_Schema{
											Schema: &v3.Schema{Type: "integer"},
										},
									}, {
										Oneof: &v3.SchemaOrReference_Schema{
											Schema: &v3.Schema{Type: "boolean"},
										},
									}, {
										Oneof: &v3.SchemaOrReference_Schema{
											Schema: &v3.Schema{Type: "object"},
										},
									}, {
										Oneof: &v3.SchemaOrReference_Schema{
											Schema: &v3.Schema{
												Type: "array",
												Items: &v3.ItemsItem{
													SchemaOrReference: []*v3.SchemaOrReference{{
														Oneof: &v3.SchemaOrReference_Reference{
															Reference: &v3.Reference{XRef: "#/components/schemas/" + protobufValueName},
														},
													}},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			)
			continue
		}

		// Get the message description from the comments.
		messageDescription := g.filterCommentString(message.Comments.Leading, true)

		// Build an array holding the fields of the message.
		definitionProperties := &v3.Properties{
			AdditionalProperties: make([]*v3.NamedSchemaOrReference, 0),
		}

		var required []string
		for _, field := range message.Fields {
			// Check the field annotations to see if this is a readonly or writeonly field.
			inputOnly := false
			outputOnly := false
			extension := proto.GetExtension(field.Desc.Options(), annotations.E_FieldBehavior)
			if extension != nil {
				switch v := extension.(type) {
				case []annotations.FieldBehavior:
					for _, vv := range v {
						switch vv {
						case annotations.FieldBehavior_OUTPUT_ONLY:
							outputOnly = true
						case annotations.FieldBehavior_INPUT_ONLY:
							inputOnly = true
						case annotations.FieldBehavior_REQUIRED:
							required = append(required, g.formatFieldName(field))
						}
					}
				default:
					log.Printf("unsupported extension type %T", extension)
				}
			}

			// The field is either described by a reference or a schema.
			fieldSchema := g.schemaOrReferenceForField(field.Desc)
			if fieldSchema == nil {
				continue
			}

			if schema, ok := fieldSchema.Oneof.(*v3.SchemaOrReference_Schema); ok {
				// Get the field description from the comments.
				schema.Schema.Description = g.filterCommentString(field.Comments.Leading, true)
				if outputOnly {
					schema.Schema.ReadOnly = true
				}
				if inputOnly {
					schema.Schema.WriteOnly = true
				}
			}

			definitionProperties.AdditionalProperties = append(
				definitionProperties.AdditionalProperties,
				&v3.NamedSchemaOrReference{
					Name:  g.formatFieldName(field),
					Value: fieldSchema,
				},
			)
		}
		// Add the schema to the components.schema list.
		d.Components.Schemas.AdditionalProperties = append(d.Components.Schemas.AdditionalProperties,
			&v3.NamedSchemaOrReference{
				Name: g.formatMessageName(message),
				Value: &v3.SchemaOrReference{
					Oneof: &v3.SchemaOrReference_Schema{
						Schema: &v3.Schema{
							Type:        "object",
							Description: messageDescription,
							Properties:  definitionProperties,
							Required:    required,
						},
					},
				},
			},
		)
	}
}

// contains returns true if an array contains a specified string.
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// appendUniuqe appends a string, to a string slice, if the string is not already in the slice
func appendUniuqe(s []string, e string) []string {
	if !contains(s, e) {
		return append(s, e)
	}
	return s
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
