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
	"log"
	"strings"

	openapiv3 "github.com/googleapis/gnostic/OpenAPIv3"
)

// NewServiceModelV3 builds a model of an API service for use in code generation.
func NewServiceModelV3(document *openapiv3.Document, packageName string) (*ServiceModel, error) {
	// Set model properties from passed-in document.
	model := &ServiceModel{}
	model.Name = document.Info.Title
	model.Package = packageName // Set package name from argument.
	model.Types = make([]*ServiceType, 0)
	model.Methods = make([]*ServiceMethod, 0)
	err := model.buildServiceV3(document)
	return model, err
}

// buildServiceV3 builds an API service description, preprocessing its types and methods for code generation.
func (model *ServiceModel) buildServiceV3(document *openapiv3.Document) (err error) {
	// Collect service type descriptions from Components/Schemas section.
	if document.Components != nil && document.Components.Schemas != nil {
		for _, pair := range document.Components.Schemas.AdditionalProperties {
			if schema := pair.Value.GetSchema(); schema != nil {
				t, err := model.buildServiceTypeFromSchemaV3(pair.Name, schema)
				if err != nil {
					return err
				}
				model.Types = append(model.Types, t)
			}
		}
	}
	// Collect service method descriptions from Paths section.
	for _, pair := range document.Paths.Path {
		v := pair.Value
		if v.Get != nil {
			model.buildServiceMethodFromOperationV3(v.Get, "GET", pair.Name)
		}
		if v.Post != nil {
			model.buildServiceMethodFromOperationV3(v.Post, "POST", pair.Name)
		}
		if v.Put != nil {
			model.buildServiceMethodFromOperationV3(v.Put, "PUT", pair.Name)
		}
		if v.Delete != nil {
			model.buildServiceMethodFromOperationV3(v.Delete, "DELETE", pair.Name)
		}
	}
	return err
}

// buildServiceTypeFromSchemaV3 builds a service type description from a schema in the API description
func (model *ServiceModel) buildServiceTypeFromSchemaV3(
	name string,
	schema *openapiv3.Schema) (t *ServiceType, err error) {
	t = &ServiceType{}
	t.Name = strings.Title(filteredTypeName(name))
	t.Description = t.Name + " implements the service definition of " + name
	t.Fields = make([]*ServiceTypeField, 0)
	if schema.Properties != nil {
		if len(schema.Properties.AdditionalProperties) > 0 {
			// If the schema has properties, generate a struct.
			t.Kind = "struct"
		}
		for _, pair2 := range schema.Properties.AdditionalProperties {
			if schema := pair2.Value; schema != nil {
				var f ServiceTypeField
				f.Name = strings.Title(pair2.Name)
				f.FieldName = strings.Replace(f.Name, "-", "_", -1)
				f.Type = typeForSchemaOrReferenceV3(schema)
				f.JSONName = pair2.Name
				t.Fields = append(t.Fields, &f)
			}
		}
	}
	if len(t.Fields) == 0 {
		if schema.AdditionalProperties != nil {
			// If the schema has no fixed properties and additional properties of a specified type,
			// generate a map pointing to objects of that type.
			mapType := typeForRef(schema.AdditionalProperties.GetSchemaOrReference().GetReference().GetXRef())
			t.Kind = "map[string]" + mapType
		}
	}
	return t, err
}

// buildServiceMethodFromOperationV3 builds a service method description
func (model *ServiceModel) buildServiceMethodFromOperationV3(
	op *openapiv3.Operation,
	method string,
	path string) (err error) {
	var m ServiceMethod
	m.Name = cleanupOperationName(op.OperationId)
	m.Path = path
	m.Method = method
	if m.Name == "" {
		m.Name = generateOperationName(method, path)
	}
	m.Description = op.Description
	m.HandlerName = "Handle" + m.Name
	m.ProcessorName = m.Name
	m.ClientName = m.Name
	m.ParametersType, err = model.buildServiceTypeFromParametersV3(m.Name, op.Parameters, op.RequestBody)
	m.ResponsesType, err = model.buildServiceTypeFromResponsesV3(&m, m.Name, op.Responses)
	model.Methods = append(model.Methods, &m)
	return err
}

// buildServiceTypeFromParametersV3 builds a service type description from the parameters of an API method
func (model *ServiceModel) buildServiceTypeFromParametersV3(
	name string,
	parameters []*openapiv3.ParameterOrReference,
	requestBody *openapiv3.RequestBodyOrReference) (t *ServiceType, err error) {
	t = &ServiceType{}
	t.Name = name + "Parameters"
	t.Description = t.Name + " holds parameters to " + name
	t.Kind = "struct"
	t.Fields = make([]*ServiceTypeField, 0)
	for _, parametersItem := range parameters {
		var f ServiceTypeField
		f.Type = fmt.Sprintf("%+v", parametersItem)
		parameter := parametersItem.GetParameter()
		if parameter != nil {
			f.Position = parameter.In
			f.Name = parameter.Name
			f.FieldName = goFieldName(f.Name)
			if parameter.GetSchema() != nil && parameter.GetSchema() != nil {
				f.Type = typeForSchemaOrReferenceV3(parameter.GetSchema())
				f.NativeType = f.Type
			}
			f.JSONName = f.Name
			f.ParameterName = goParameterName(f.FieldName)
			t.Fields = append(t.Fields, &f)
			if f.NativeType == "integer" {
				f.NativeType = "int64"
			}
		}
	}
	if requestBody != nil {
		content := requestBody.GetRequestBody().GetContent()
		if content != nil {
			for _, pair2 := range content.GetAdditionalProperties() {
				var f ServiceTypeField
				f.Position = "body"
				f.Name = "resource"
				f.FieldName = goFieldName(f.Name)
				f.ValueType = typeForSchemaOrReferenceV3(pair2.GetValue().GetSchema())
				f.Type = "*" + f.ValueType
				f.NativeType = f.Type
				f.JSONName = f.Name
				f.ParameterName = goParameterName(f.FieldName)
				t.Fields = append(t.Fields, &f)
			}
		}
	}
	if len(t.Fields) > 0 {
		model.Types = append(model.Types, t)
		return t, err
	}
	return nil, err
}

// buildServiceTypeFromResponsesV3 builds a service type description from the responses of an API method
func (model *ServiceModel) buildServiceTypeFromResponsesV3(
	m *ServiceMethod,
	name string,
	responses *openapiv3.Responses) (t *ServiceType, err error) {
	t = &ServiceType{}
	t.Name = name + "Responses"
	t.Description = t.Name + " holds responses of " + name
	t.Kind = "struct"
	t.Fields = make([]*ServiceTypeField, 0)

	m.ResultTypeName = t.Name

	for _, pair := range responses.ResponseOrReference {
		var f ServiceTypeField
		f.Name = pair.Name
		f.FieldName = propertyNameForResponseCode(pair.Name)
		f.JSONName = ""
		response := pair.Value.GetResponse()
		if response != nil && response.GetContent() != nil {
			for _, pair2 := range response.GetContent().GetAdditionalProperties() {
				f.ValueType = typeForSchemaOrReferenceV3(pair2.GetValue().GetSchema())
				f.Type = "*" + f.ValueType
				t.Fields = append(t.Fields, &f)
			}
		}
	}

	if len(t.Fields) > 0 {
		model.Types = append(model.Types, t)
		return t, err
	}
	return nil, err
}

// typeForSchemaOrReferenceV3 determines the language-specific type of a schema or reference
func typeForSchemaOrReferenceV3(value *openapiv3.SchemaOrReference) (typeName string) {
	if value.GetSchema() != nil {
		return typeForSchemaV3(value.GetSchema())
	}
	if value.GetReference() != nil {
		return typeForRef(value.GetReference().XRef)
	}
	return "todo"
}

// typeForSchema determines the language-specific type of a schema
func typeForSchemaV3(schema *openapiv3.Schema) (typeName string) {
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
