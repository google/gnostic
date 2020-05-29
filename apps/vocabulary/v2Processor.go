package main

import (
	"io/ioutil"

	"github.com/golang/protobuf/proto"

	openapi_v2 "github.com/googleapis/gnostic/openapiv2"
)

func readDocumentFromFileWithName(filename string) (*openapi_v2.Document, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	document := &openapi_v2.Document{}
	err = proto.Unmarshal(data, document)
	if err != nil {
		return nil, err
	}
	return document, nil

}

func processDocument(document *openapi_v2.Document, schemas map[string]int, operationId map[string]int, names map[string]int, properties map[string]int) {
	if document.Definitions != nil {
		for _, pair := range document.Definitions.AdditionalProperties {
			schemas[pair.Name] += 1
			processSchema(pair.Value, properties)
		}
	}
	for _, pair := range document.Paths.Path {
		v := pair.Value
		if v.Get != nil {
			processOperation(v.Get, operationId, names)
		}
		if v.Post != nil {
			processOperation(v.Post, operationId, names)
		}
		if v.Put != nil {
			processOperation(v.Put, operationId, names)
		}
		if v.Patch != nil {
			processOperation(v.Patch, operationId, names)
		}
		if v.Delete != nil {
			processOperation(v.Delete, operationId, names)
		}
	}
}

func processOperation(operation *openapi_v2.Operation, operationId map[string]int, names map[string]int) {
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
				processOperationParamaters(operation, names, nonBodyParam)

			}
		}
	}
}

func processOperationParamaters(operation *openapi_v2.Operation, names map[string]int, nonBodyParam *openapi_v2.NonBodyParameter) {
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

func processSchema(schema *openapi_v2.Schema, properties map[string]int) {
	if schema.Properties == nil {
		return
	}
	for _, pair := range schema.Properties.AdditionalProperties {
		properties[pair.Name] += 1
	}
}
