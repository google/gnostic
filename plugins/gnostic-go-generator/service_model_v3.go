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
	"strings"

	openapiv3 "github.com/googleapis/gnostic/OpenAPIv3"
	"log"
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
	if err != nil {
		return nil, err
	}
	return model, nil
}

// buildServiceV3 builds an API service description, preprocessing its types and methods for code generation.
func (model *ServiceModel) buildServiceV3(document *openapiv3.Document) (err error) {
	// Collect service type descriptions from Definitions section.
	if document.Components != nil && document.Components.Schemas != nil {
		for _, pair := range document.Components.Schemas.AdditionalProperties {
			if schema := pair.Value.GetSchema(); schema != nil {
				t, err := model.buildServiceTypeFromDefinitionV3(pair.Name, schema)
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

func (model *ServiceModel) buildServiceTypeFromDefinitionV3(name string, schema *openapiv3.Schema) (t *ServiceType, err error) {
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

func (model *ServiceModel) buildServiceMethodFromOperationV3(op *openapiv3.Operation, method string, path string) (err error) {
	log.Printf("build service method %s %s", method, path)
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
	m.ParametersType, err = model.buildServiceTypeFromParametersV3(m.Name, op.Parameters)
	if m.ParametersType != nil {
		m.ParametersTypeName = m.ParametersType.Name
	}
	m.ResponsesType, err = model.buildServiceTypeFromResponsesV3(&m, m.Name, op.Responses)
	if m.ResponsesType != nil {
		m.ResponsesTypeName = m.ResponsesType.Name
	}
	model.Methods = append(model.Methods, &m)
	return err
}

func (model *ServiceModel) buildServiceTypeFromParametersV3(name string, parameters []*openapiv3.ParameterOrReference) (t *ServiceType, err error) {
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
			f.FieldName = snakeCaseToCamelCase(strings.Replace(f.Name, "-", "_", -1))
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
	if len(t.Fields) > 0 {
		model.Types = append(model.Types, t)
		return t, err
	}
	return nil, err
}

func (model *ServiceModel) buildServiceTypeFromResponsesV3(m *ServiceMethod, name string, responses *openapiv3.Responses) (t *ServiceType, err error) {
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
				log.Printf("pair: %+v", pair2)
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

func typeForSchemaOrReferenceV3(value *openapiv3.SchemaOrReference) (typeName string) {
	if value.GetSchema() != nil {
		return typeForSchemaV3(value.GetSchema())
	}
	if value.GetReference() != nil {
		return typeForRef(value.GetReference().XRef)
	}
	return "todo"
}

func typeForSchemaV3(schema *openapiv3.Schema) (typeName string) {
	if schema.Type != "" {
		format := schema.Format
		if schema.Type == "string" {
			return "string"
		}
		if schema.Type == "integer" && format == "int32" {
			return "int32"
		}
		if schema.Type == "integer" {
			return "int"
		}
		if schema.Type == "number" {
			return "int"
		}
		if schema.Type == "array" && schema.Items != nil {
			// we have an array.., but of what?
			items := schema.Items.SchemaOrReference
			if len(items) == 1 && items[0].GetReference().GetXRef() != "" {
				return "[]" + typeForRef(items[0].GetReference().GetXRef())
			}
		}
		if schema.Type == "object" && schema.AdditionalProperties == nil {
			return "map[string]interface{}"
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
	// this function is incomplete... so return a string representing anything that we don't handle
	return fmt.Sprintf("%v", schema)
}
