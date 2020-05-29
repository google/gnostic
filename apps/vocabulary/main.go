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

package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	metrics "github.com/googleapis/gnostic/metrics"
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

func fillProtoStructures(m map[string]int) []*metrics.WordCount {
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

func main() {
	flag.Parse()
	args := flag.Args()

	document, err := readDocumentFromFileWithName(args[0])

	if err != nil {
		log.Printf("Error reading %s. This sample expects OpenAPI v2.", args[0])
		os.Exit(1)
	}

	var schemas map[string]int
	schemas = make(map[string]int)

	var operationId map[string]int
	operationId = make(map[string]int)

	var names map[string]int
	names = make(map[string]int)

	var properties map[string]int
	properties = make(map[string]int)

	processDocument(document, schemas, operationId, names, properties)

	vocab := &metrics.Vocabulary{
		Schemas:    fillProtoStructures(schemas),
		Operations: fillProtoStructures(operationId),
		Paramaters: fillProtoStructures(names),
		Properties: fillProtoStructures(properties),
	}

	bytes, err := proto.Marshal(vocab)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile("vocabulary.pb", bytes, 0644)
	if err != nil {
		panic(err)
	}

	m := jsonpb.Marshaler{Indent: " "}
	s, err := m.MarshalToString(vocab)
	jsonOutput := []byte(s)
	err = ioutil.WriteFile("vocabulary.json", jsonOutput, os.ModePerm)
}
