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

package gnostic_surface_v1

import (
	"errors"
	"fmt"
	"log"

	openapiv3 "github.com/googleapis/gnostic/OpenAPIv3"
)

// NewModelFromOpenAPIv3 builds a model of an API service for use in code generation.
func NewModelFromOpenAPI3(document *openapiv3.Document) (*Model, error) {
	return newOpenAPI3Builder().buildModel(document)
}

type OpenAPI3Builder struct {
	model *Model
}

func newOpenAPI3Builder() *OpenAPI3Builder {
	return &OpenAPI3Builder{model: &Model{}}
}

func (b *OpenAPI3Builder) buildModel(document *openapiv3.Document) (*Model, error) {
	// Set model properties from passed-in document.
	b.model.Name = document.Info.Title
	b.model.Types = make([]*Type, 0)
	b.model.Methods = make([]*Method, 0)
	err := b.build(document)
	if err != nil {
		return nil, err
	}
	return b.model, nil
}

// build builds an API service description, preprocessing its types and methods for code generation.
func (b *OpenAPI3Builder) build(document *openapiv3.Document) (err error) {
	// Collect service type descriptions from Components/Schemas.
	if document.Components != nil && document.Components.Schemas != nil {
		for _, pair := range document.Components.Schemas.AdditionalProperties {
			t, err := b.buildTypeFromSchemaOrReference(pair.Name, pair.Value)
			if err != nil {
				return err
			}
			if t != nil {
				b.model.addType(t)
			}
		}
	}
	// Collect service method descriptions from each PathItem.
	for _, pair := range document.Paths.Path {
		b.buildMethodFromPathItem(pair.Name, pair.Value)
	}
	return err
}

// buildTypeFromSchemaOrReference builds a service type description from a schema in the API description.
func (b *OpenAPI3Builder) buildTypeFromSchemaOrReference(
	name string,
	schemaOrReference *openapiv3.SchemaOrReference) (t *Type, err error) {
	if schema := schemaOrReference.GetSchema(); schema != nil {
		t = &Type{}
		t.Name = name
		t.Description = "implements the service definition of " + name
		t.Fields = make([]*Field, 0)
		if schema.Properties != nil {
			if len(schema.Properties.AdditionalProperties) > 0 {
				// If the schema has properties, generate a struct.
				t.Kind = Kind_STRUCT
			}
			for _, pair2 := range schema.Properties.AdditionalProperties {
				if schema := pair2.Value; schema != nil {
					var f Field
					f.Name = pair2.Name
					f.Type = b.typeForSchemaOrReference(schema)
					f.Serialize = true
					t.addField(&f)
				}
			}
		}
		if len(t.Fields) == 0 {
			if schema.AdditionalProperties != nil {
				// If the schema has no fixed properties and additional properties of a specified type,
				// generate a map pointing to objects of that type.
				mapType := typeForRef(schema.AdditionalProperties.GetSchemaOrReference().GetReference().GetXRef())
				t.Kind = Kind_MAP
				t.MapType = mapType
			}
		}
		return t, err
	} else {
		return nil, errors.New("unable to determine service type for referenced schema " + name)
	}
}

// buildMethodFromOperation builds a service method description
func (b *OpenAPI3Builder) buildMethodFromPathItem(
	path string,
	pathItem *openapiv3.PathItem) (err error) {
	for _, method := range []string{"GET", "PUT", "POST", "DELETE", "OPTIONS", "HEAD", "PATCH", "TRACE"} {
		var op *openapiv3.Operation
		switch method {
		case "GET":
			op = pathItem.Get
		case "PUT":
			op = pathItem.Put
		case "POST":
			op = pathItem.Post
		case "DELETE":
			op = pathItem.Delete
		case "OPTIONS":
			op = pathItem.Options
		case "HEAD":
			op = pathItem.Head
		case "PATCH":
			op = pathItem.Patch
		case "TRACE":
			op = pathItem.Trace
		}
		if op != nil {
			var m Method
			m.Name = sanitizeOperationName(op.OperationId)
			m.Path = path
			m.Method = method
			if m.Name == "" {
				m.Name = generateOperationName(method, path)
			}
			m.Description = op.Description
			m.ParametersTypeName, err = b.buildTypeFromParameters(m.Name, op.Parameters, op.RequestBody)
			m.ResponsesTypeName, err = b.buildTypeFromResponses(&m, m.Name, op.Responses)
			b.model.addMethod(&m)
		}
	}
	return err
}

// buildTypeFromParameters builds a service type description from the parameters of an API method
func (b *OpenAPI3Builder) buildTypeFromParameters(
	name string,
	parameters []*openapiv3.ParameterOrReference,
	requestBody *openapiv3.RequestBodyOrReference) (typeName string, err error) {
	t := &Type{}
	t.Name = name + "Parameters"
	t.Description = t.Name + " holds parameters to " + name
	t.Kind = Kind_STRUCT
	t.Fields = make([]*Field, 0)
	for _, parametersItem := range parameters {
		var f Field
		f.Type = fmt.Sprintf("%+v", parametersItem)
		parameter := parametersItem.GetParameter()
		if parameter != nil {
			switch parameter.In {
			case "body":
				f.Position = Position_BODY
			case "header":
				f.Position = Position_HEADER
			case "formdata":
				f.Position = Position_FORMDATA
			case "query":
				f.Position = Position_QUERY
			case "path":
				f.Position = Position_PATH
			}
			f.Name = parameter.Name
			if parameter.GetSchema() != nil && parameter.GetSchema() != nil {
				f.Type = b.typeForSchemaOrReference(parameter.GetSchema())
			}
			f.Serialize = true
			t.addField(&f)
		}
	}
	if requestBody != nil {
		content := requestBody.GetRequestBody().GetContent()
		if content != nil {
			for _, pair2 := range content.GetAdditionalProperties() {
				var f Field
				f.Position = Position_BODY
				f.Name = "resource"
				f.ValueType = b.typeForSchemaOrReference(pair2.GetValue().GetSchema())
				f.Type = "*" + f.ValueType
				f.Serialize = true
				t.addField(&f)
			}
		}
	}
	if len(t.Fields) > 0 {
		b.model.addType(t)
		return t.Name, err
	}
	return "", err
}

// buildTypeFromResponses builds a service type description from the responses of an API method
func (b *OpenAPI3Builder) buildTypeFromResponses(
	m *Method,
	name string,
	responses *openapiv3.Responses) (typeName string, err error) {
	t := &Type{}
	t.Name = name + "Responses"
	t.Description = t.Name + " holds responses of " + name
	t.Kind = Kind_STRUCT
	t.Fields = make([]*Field, 0)

	for _, pair := range responses.ResponseOrReference {
		var f Field
		f.Name = pair.Name
		f.Serialize = false
		response := pair.Value.GetResponse()
		if response != nil && response.GetContent() != nil {
			for _, pair2 := range response.GetContent().GetAdditionalProperties() {
				f.ValueType = b.typeForSchemaOrReference(pair2.GetValue().GetSchema())
				f.Type = "*" + f.ValueType
				t.addField(&f)
			}
		}
	}

	if len(t.Fields) > 0 {
		b.model.addType(t)
		return t.Name, err
	}
	return "", err
}

// typeForSchemaOrReference determines the language-specific type of a schema or reference
func (b *OpenAPI3Builder) typeForSchemaOrReference(value *openapiv3.SchemaOrReference) (typeName string) {
	if value.GetSchema() != nil {
		return b.typeForSchema(value.GetSchema())
	}
	if value.GetReference() != nil {
		return typeForRef(value.GetReference().XRef)
	}
	return "todo"
}

// typeForSchema determines the language-specific type of a schema
func (b *OpenAPI3Builder) typeForSchema(schema *openapiv3.Schema) (typeName string) {
	if schema.Type != "" {
		format := schema.Format
		switch schema.Type {
		case "string":
			return "string"
		case "integer":
			if format == "int32" {
				return "int32"
			}
			return "int"
		case "number":
			return "int"
		case "boolean":
			return "bool"
		case "array":
			if schema.Items != nil {
				// we have an array.., but of what?
				items := schema.Items.SchemaOrReference
				if len(items) == 1 {
					if items[0].GetReference().GetXRef() != "" {
						return "[]" + typeForRef(items[0].GetReference().GetXRef())
					} else if items[0].GetSchema().Type == "string" {
						return "[]string"
					} else if items[0].GetSchema().Type == "object" {
						return "[]interface{}"
					}
				}
			}
		case "object":
			if schema.AdditionalProperties == nil {
				return "map[string]interface{}"
			}
		default:

		}
	}
	if schema.AdditionalProperties != nil {
		additionalProperties := schema.AdditionalProperties
		if propertySchema := additionalProperties.GetSchemaOrReference().GetReference(); propertySchema != nil {
			if ref := propertySchema.XRef; ref != "" {
				return "map[string]" + typeForRef(ref)
			}
		}
	}
	// this function is incomplete... return a string representing anything that we don't handle
	log.Printf("unimplemented: %v", schema)
	return fmt.Sprintf("unimplemented: %v", schema)
}
