package main

import (
	"sort"

	metrics "github.com/googleapis/gnostic/metrics"
	openapi_v3 "github.com/googleapis/gnostic/openapiv3"
)

func fillProtoStructures(m map[string]int) []*metrics.WordCount {
	key_names := make([]string, 0, len(m))
	for key := range m {
		key_names = append(key_names, key)
	}
	sort.Strings(key_names)

	counts := make([]*metrics.WordCount, 0)
	for _, k := range key_names {
		temp := &metrics.WordCount{
			Word:  k,
			Count: int32(m[k]),
		}
		counts = append(counts, temp)
	}
	return counts
}

func processOperationV3(operation *openapi_v3.Operation, operationId, parameters map[string]int) {
	if operation.OperationId != "" {
		operationId[operation.OperationId] += 1
	}
	for _, item := range operation.Parameters {
		switch t := item.Oneof.(type) {
		case *openapi_v3.ParameterOrReference_Parameter:
			parameters[t.Parameter.Name] += 1
		}
	}
}

func processComponentsV3(components *openapi_v3.Components, schemas, properties map[string]int) {
	processParametersV3(components, schemas, properties)
	processSchemasV3(components, schemas)
	processResponsesV3(components, schemas)
}

func processParametersV3(components *openapi_v3.Components, schemas, properties map[string]int) {
	if components.Parameters == nil {
		return
	}
	for _, pair := range components.Parameters.AdditionalProperties {
		schemas[pair.Name] += 1
		switch t := pair.Value.Oneof.(type) {
		case *openapi_v3.ParameterOrReference_Parameter:
			properties[t.Parameter.Name] += 1
		}
	}
}

func processSchemasV3(components *openapi_v3.Components, schemas map[string]int) {
	if components.Schemas == nil {
		return
	}
	for _, pair := range components.Schemas.AdditionalProperties {
		schemas[pair.Name] += 1
	}
}

func processResponsesV3(components *openapi_v3.Components, schemas map[string]int) {
	if components.Responses == nil {
		return
	}
	for _, pair := range components.Responses.AdditionalProperties {
		schemas[pair.Name] += 1
	}
}

func processDocumentV3(document *openapi_v3.Document) *metrics.Vocabulary {
	var schemas map[string]int
	schemas = make(map[string]int)

	var operationId map[string]int
	operationId = make(map[string]int)

	var parameters map[string]int
	parameters = make(map[string]int)

	var properties map[string]int
	properties = make(map[string]int)

	if document.Components != nil {
		processComponentsV3(document.Components, schemas, properties)

	}
	for _, pair := range document.Paths.Path {
		v := pair.Value
		if v.Get != nil {
			processOperationV3(v.Get, operationId, parameters)
		}
		if v.Post != nil {
			processOperationV3(v.Post, operationId, parameters)
		}
		if v.Put != nil {
			processOperationV3(v.Put, operationId, parameters)
		}
		if v.Patch != nil {
			processOperationV3(v.Patch, operationId, parameters)
		}
		if v.Delete != nil {
			processOperationV3(v.Delete, operationId, parameters)
		}
	}

	vocab := &metrics.Vocabulary{
		Properties: fillProtoStructures(properties),
		Schemas:    fillProtoStructures(schemas),
		Operations: fillProtoStructures(operationId),
		Parameters: fillProtoStructures(parameters),
	}
	return vocab
}
