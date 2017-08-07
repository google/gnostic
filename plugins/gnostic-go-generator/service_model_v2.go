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

	openapiv2 "github.com/googleapis/gnostic/OpenAPIv2"
)

// NewServiceModelV2 builds a model of an API service for use in code generation.
func NewServiceModelV2(document *openapiv2.Document, packageName string) (*ServiceModel, error) {
	// Set model properties from passed-in document.
	model := &ServiceModel{}
	model.Name = document.Info.Title
	model.Package = packageName // Set package name from argument.
	model.Types = make([]*ServiceType, 0)
	model.Methods = make([]*ServiceMethod, 0)
	err := model.buildServiceV2(document)
	if err != nil {
		return nil, err
	}
	return model, nil
}

// buildServiceV2 builds an API service description, preprocessing its types and methods for code generation.
func (model *ServiceModel) buildServiceV2(document *openapiv2.Document) (err error) {
	// Collect service type descriptions from Definitions section.
	if document.Definitions != nil {
		for _, pair := range document.Definitions.AdditionalProperties {
			t, err := model.buildServiceTypeFromDefinitionV2(pair.Name, pair.Value)
			if err != nil {
				return err
			}
			model.Types = append(model.Types, t)
		}
	}
	// Collect service method descriptions from Paths section.
	for _, pair := range document.Paths.Path {
		v := pair.Value
		if v.Get != nil {
			model.buildServiceMethodFromOperationV2(v.Get, "GET", pair.Name)
		}
		if v.Post != nil {
			model.buildServiceMethodFromOperationV2(v.Post, "POST", pair.Name)
		}
		if v.Put != nil {
			model.buildServiceMethodFromOperationV2(v.Put, "PUT", pair.Name)
		}
		if v.Delete != nil {
			model.buildServiceMethodFromOperationV2(v.Delete, "DELETE", pair.Name)
		}
	}
	return err
}

func (model *ServiceModel) buildServiceTypeFromDefinitionV2(name string, schema *openapiv2.Schema) (t *ServiceType, err error) {
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
			var f ServiceTypeField
			f.Name = strings.Title(pair2.Name)
			f.FieldName = strings.Replace(f.Name, "-", "_", -1)
			f.Type = typeForSchemaV2(pair2.Value)
			f.JSONName = pair2.Name
			t.Fields = append(t.Fields, &f)
		}
	}
	if len(t.Fields) == 0 {
		if schema.AdditionalProperties != nil {
			// If the schema has no fixed properties and additional properties of a specified type,
			// generate a map pointing to objects of that type.
			mapType := typeForRef(schema.AdditionalProperties.GetSchema().XRef)
			t.Kind = "map[string]" + mapType
		}
	}
	return t, err
}

func (model *ServiceModel) buildServiceMethodFromOperationV2(op *openapiv2.Operation, method string, path string) (err error) {
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
	m.ParametersType, err = model.buildServiceTypeFromParametersV2(m.Name, op.Parameters)
	m.ResponsesType, err = model.buildServiceTypeFromResponsesV2(&m, m.Name, op.Responses)
	model.Methods = append(model.Methods, &m)
	return err
}

func (model *ServiceModel) buildServiceTypeFromParametersV2(name string, parameters []*openapiv2.ParametersItem) (t *ServiceType, err error) {
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
			bodyParameter := parameter.GetBodyParameter()
			if bodyParameter != nil {
				f.Name = bodyParameter.Name
				f.FieldName = snakeCaseToCamelCase(strings.Replace(f.Name, "-", "_", -1))
				if bodyParameter.Schema != nil {
					f.Type = typeForSchemaV2(bodyParameter.Schema)
					f.NativeType = f.Type
					f.Position = "body"
				}
			}
			nonBodyParameter := parameter.GetNonBodyParameter()
			if nonBodyParameter != nil {
				headerParameter := nonBodyParameter.GetHeaderParameterSubSchema()
				if headerParameter != nil {
					f.Name = headerParameter.Name
					f.FieldName = snakeCaseToCamelCase(strings.Replace(f.Name, "-", "_", -1))
					f.Type = headerParameter.Type
					f.NativeType = f.Type
					f.Position = "header"
				}
				formDataParameter := nonBodyParameter.GetFormDataParameterSubSchema()
				if formDataParameter != nil {
					f.Name = formDataParameter.Name
					f.FieldName = snakeCaseToCamelCase(strings.Replace(f.Name, "-", "_", -1))
					f.Type = formDataParameter.Type
					f.NativeType = f.Type
					f.Position = "formdata"
				}
				queryParameter := nonBodyParameter.GetQueryParameterSubSchema()
				if queryParameter != nil {
					f.Name = queryParameter.Name
					f.FieldName = snakeCaseToCamelCase(strings.Replace(f.Name, "-", "_", -1))
					f.Type = queryParameter.Type
					f.NativeType = f.Type
					f.Position = "query"
				}
				pathParameter := nonBodyParameter.GetPathParameterSubSchema()
				if pathParameter != nil {
					f.Name = pathParameter.Name
					f.FieldName = snakeCaseToCamelCase(strings.Replace(f.Name, "-", "_", -1))
					f.Type = pathParameter.Type
					f.NativeType = f.Type
					f.Position = "path"
					f.Type = typeForName(pathParameter.Type, pathParameter.Format)
				}
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

func (model *ServiceModel) buildServiceTypeFromResponsesV2(m *ServiceMethod, name string, responses *openapiv2.Responses) (t *ServiceType, err error) {
	t = &ServiceType{}
	t.Name = name + "Responses"
	t.Description = t.Name + " holds responses of " + name
	t.Kind = "struct"
	t.Fields = make([]*ServiceTypeField, 0)

	m.ResultTypeName = t.Name

	for _, responseCode := range responses.ResponseCode {
		var f ServiceTypeField
		f.Name = responseCode.Name
		f.FieldName = propertyNameForResponseCode(responseCode.Name)
		f.JSONName = ""
		response := responseCode.Value.GetResponse()
		if response != nil && response.Schema != nil && response.Schema.GetSchema() != nil {
			f.ValueType = typeForSchemaV2(response.Schema.GetSchema())
			f.Type = "*" + f.ValueType
			t.Fields = append(t.Fields, &f)
		}
	}

	if len(t.Fields) > 0 {
		model.Types = append(model.Types, t)
		return t, err
	}
	return nil, err
}

func typeForSchemaV2(schema *openapiv2.Schema) (typeName string) {
	ref := schema.XRef
	if ref != "" {
		return typeForRef(ref)
	}
	if schema.Type != nil {
		types := schema.Type.Value
		format := schema.Format
		if len(types) == 1 && types[0] == "string" {
			return "string"
		}
		if len(types) == 1 && types[0] == "integer" && format == "int32" {
			return "int32"
		}
		if len(types) == 1 && types[0] == "integer" {
			return "int"
		}
		if len(types) == 1 && types[0] == "number" {
			return "int"
		}
		if len(types) == 1 && types[0] == "array" && schema.Items != nil {
			// we have an array.., but of what?
			items := schema.Items.Schema
			if len(items) == 1 && items[0].XRef != "" {
				return "[]" + typeForRef(items[0].XRef)
			}
		}
		if len(types) == 1 && types[0] == "object" && schema.AdditionalProperties == nil {
			return "map[string]interface{}"
		}
	}
	if schema.AdditionalProperties != nil {
		additionalProperties := schema.AdditionalProperties
		if propertySchema := additionalProperties.GetSchema(); propertySchema != nil {
			if ref := propertySchema.XRef; ref != "" {
				return "map[string]" + typeForRef(ref)
			}
		}
	}
	// this function is incomplete... so return a string representing anything that we don't handle
	return fmt.Sprintf("%v", schema)
}
