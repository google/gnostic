package main

import (
	"io/ioutil"

	"github.com/golang/protobuf/proto"

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

func processDocumentV3(document *openapi_v3.Document, schemas map[string]int, operationId map[string]int, names map[string]int, properties map[string]int) {
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
}

func processOperationV3(operation *openapi_v3.Operation, operationId map[string]int, names map[string]int) {
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

func processComponentsV3(components *openapi_v3.Components, schemas map[string]int, properties map[string]int) {
	processParametersV3(components, schemas, properties)
	processSchemasV3(components, schemas)
	processResponsesV3(components, schemas)

}

func processParametersV3(components *openapi_v3.Components, schemas map[string]int, properties map[string]int) {
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
