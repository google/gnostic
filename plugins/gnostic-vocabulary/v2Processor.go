package main

import (
	metrics "github.com/googleapis/gnostic/metrics"
	openapi_v2 "github.com/googleapis/gnostic/openapiv2"
)

func processOperationV2(operation *openapi_v2.Operation, operationId, names map[string]int) {
	if operation.OperationId != "" {
		operationId[operation.OperationId] += 1
	}
	for _, item := range operation.Parameters {
		switch t := item.Oneof.(type) {
		case *openapi_v2.ParametersItem_Parameter:
			switch t2 := t.Parameter.Oneof.(type) {
			case *openapi_v2.Parameter_BodyParameter:
				names[t2.BodyParameter.Name] += 1
			case *openapi_v2.Parameter_NonBodyParameter:
				nonBodyParam := t2.NonBodyParameter
				processOperationParamatersV2(operation, names, nonBodyParam)
			}
		}
	}
}

func processOperationParamatersV2(operation *openapi_v2.Operation, names map[string]int, nonBodyParam *openapi_v2.NonBodyParameter) {
	switch t3 := nonBodyParam.Oneof.(type) {
	case *openapi_v2.NonBodyParameter_FormDataParameterSubSchema:
		names[t3.FormDataParameterSubSchema.Name] += 1
	case *openapi_v2.NonBodyParameter_HeaderParameterSubSchema:
		names[t3.HeaderParameterSubSchema.Name] += 1
	case *openapi_v2.NonBodyParameter_PathParameterSubSchema:
		names[t3.PathParameterSubSchema.Name] += 1
	case *openapi_v2.NonBodyParameter_QueryParameterSubSchema:
		names[t3.QueryParameterSubSchema.Name] += 1
	}
}

func processSchemaV2(schema *openapi_v2.Schema, properties map[string]int) {
	if schema.Properties == nil {
		return
	}
	for _, pair := range schema.Properties.AdditionalProperties {
		properties[pair.Name] += 1
	}
}

func processDocumentV2(document *openapi_v2.Document) *metrics.Vocabulary {
	var schemas map[string]int
	schemas = make(map[string]int)

	var operationId map[string]int
	operationId = make(map[string]int)

	var names map[string]int
	names = make(map[string]int)

	var properties map[string]int
	properties = make(map[string]int)

	if document.Definitions != nil {
		for _, pair := range document.Definitions.AdditionalProperties {
			schemas[pair.Name] += 1
			processSchemaV2(pair.Value, properties)
		}
	}
	for _, pair := range document.Paths.Path {
		v := pair.Value
		if v.Get != nil {
			processOperationV2(v.Get, operationId, names)
		}
		if v.Post != nil {
			processOperationV2(v.Post, operationId, names)
		}
		if v.Put != nil {
			processOperationV2(v.Put, operationId, names)
		}
		if v.Patch != nil {
			processOperationV2(v.Patch, operationId, names)
		}
		if v.Delete != nil {
			processOperationV2(v.Delete, operationId, names)
		}
	}

	vocab := &metrics.Vocabulary{
		Schemas:    fillProtoStructures(schemas),
		Operations: fillProtoStructures(operationId),
		Paramaters: fillProtoStructures(names),
		Properties: fillProtoStructures(properties),
	}

	return vocab
}
