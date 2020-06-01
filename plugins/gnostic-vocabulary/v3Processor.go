package main

import (
	"io/ioutil"

	"github.com/golang/protobuf/proto"

	metrics "github.com/googleapis/gnostic/metrics"
	openapi_v3 "github.com/googleapis/gnostic/openapiv3"
)

func readDocumentFromFileWithNameV3(filename string) (*openapi_v3.Document, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	document := &openapi_v3.Document{}
	err = proto.Unmarshal(data, document)
	if err != nil {
		return nil, err
	}
	return document, nil

}

func fillProtoStructuresV3(m map[string]int) []*metrics.WordCount {
	counts := make([]*metrics.WordCount, 0)
	for k, v := range m {
		temp := &metrics.WordCount{
			Word:  k,
			Count: int32(v),
		}
		counts = append(counts, temp)
	}
	return counts
}

func processOperationV3(operation *openapi_v3.Operation, operationId, names map[string]int) {
	if operation.OperationId != "" {
		operationId[operation.OperationId] += 1
	}
	for _, item := range operation.Parameters {
		switch t := item.Oneof.(type) {
		case *openapi_v3.ParameterOrReference_Parameter:
			names[t.Parameter.Name] += 1
		}
	}
}

func processComponentsV3(components *openapi_v3.Components, schemas, properties map[string]int) {
	processParametersV3(components, schemas, properties)
	processSchemasV3(components, schemas)
	processResponsesV3(components, schemas)

}

func processParametersV3(components *openapi_v3.Components, schemas, properties map[string]int) {
	for _, pair := range components.Parameters.AdditionalProperties {
		schemas[pair.Name] += 1
		switch t := pair.Value.Oneof.(type) {
		case *openapi_v3.ParameterOrReference_Parameter:
			properties[t.Parameter.Name] += 1
		}
	}
}

func processSchemasV3(components *openapi_v3.Components, schemas map[string]int) {
	for _, pair := range components.Schemas.AdditionalProperties {
		schemas[pair.Name] += 1
	}
}

func processResponsesV3(components *openapi_v3.Components, schemas map[string]int) {
	for _, pair := range components.Responses.AdditionalProperties {
		schemas[pair.Name] += 1
	}
}

func processDocumentV3(document *openapi_v3.Document, schemas, operationId, names, properties map[string]int) *metrics.Vocabulary {
	if document.Components != nil {
		processComponentsV3(document.Components, schemas, properties)

	}
	for _, pair := range document.Paths.Path {
		v := pair.Value
		if v.Get != nil {
			processOperationV3(v.Get, operationId, names)
		}
		if v.Post != nil {
			processOperationV3(v.Post, operationId, names)
		}
		if v.Put != nil {
			processOperationV3(v.Put, operationId, names)
		}
		if v.Patch != nil {
			processOperationV3(v.Patch, operationId, names)
		}
		if v.Delete != nil {
			processOperationV3(v.Delete, operationId, names)
		}
	}

	vocab := &metrics.Vocabulary{
		Schemas:    fillProtoStructuresV3(schemas),
		Operations: fillProtoStructuresV3(operationId),
		Paramaters: fillProtoStructuresV3(names),
		Properties: fillProtoStructuresV3(properties),
	}

	return vocab
}
