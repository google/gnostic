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

package vocabulary

import (
	metrics "github.com/googleapis/gnostic/metrics"
	openapi_v2 "github.com/googleapis/gnostic/openapiv2"
)

func processOperationV2(operation *openapi_v2.Operation, vocab *Vocabulary) {
	if operation.OperationId != "" {
		vocab.operationID[operation.OperationId]++
	}
	for _, item := range operation.Parameters {
		switch t := item.Oneof.(type) {
		case *openapi_v2.ParametersItem_Parameter:
			switch t2 := t.Parameter.Oneof.(type) {
			case *openapi_v2.Parameter_BodyParameter:
				vocab.parameters[t2.BodyParameter.Name]++
			case *openapi_v2.Parameter_NonBodyParameter:
				nonBodyParam := t2.NonBodyParameter
				processOperationParametersV2(operation, vocab, nonBodyParam)
			}
		}
	}
}

func processOperationParametersV2(operation *openapi_v2.Operation, vocab *Vocabulary, nonBodyParam *openapi_v2.NonBodyParameter) {
	switch t3 := nonBodyParam.Oneof.(type) {
	case *openapi_v2.NonBodyParameter_FormDataParameterSubSchema:
		vocab.parameters[t3.FormDataParameterSubSchema.Name]++
	case *openapi_v2.NonBodyParameter_HeaderParameterSubSchema:
		vocab.parameters[t3.HeaderParameterSubSchema.Name]++
	case *openapi_v2.NonBodyParameter_PathParameterSubSchema:
		vocab.parameters[t3.PathParameterSubSchema.Name]++
	case *openapi_v2.NonBodyParameter_QueryParameterSubSchema:
		vocab.parameters[t3.QueryParameterSubSchema.Name]++
	}
}

func processSchemaV2(schema *openapi_v2.Schema, vocab *Vocabulary) {
	if schema.Properties == nil {
		return
	}
	for _, pair := range schema.Properties.AdditionalProperties {
		vocab.properties[pair.Name]++
	}
}

//NewVocabularyFromOpenAPIv2 does
func NewVocabularyFromOpenAPIv2(document *openapi_v2.Document) *metrics.Vocabulary {
	var vocab Vocabulary
	vocab.schemas = make(map[string]int)
	vocab.operationID = make(map[string]int)
	vocab.parameters = make(map[string]int)
	vocab.properties = make(map[string]int)

	if document.Definitions != nil {
		for _, pair := range document.Definitions.AdditionalProperties {
			vocab.schemas[pair.Name]++
			processSchemaV2(pair.Value, &vocab)
		}
	}
	if document.Paths != nil {
		for _, pair := range document.Paths.Path {
			v := pair.Value
			if v.Get != nil {
				processOperationV2(v.Get, &vocab)
			}
			if v.Post != nil {
				processOperationV2(v.Post, &vocab)
			}
			if v.Put != nil {
				processOperationV2(v.Put, &vocab)
			}
			if v.Patch != nil {
				processOperationV2(v.Patch, &vocab)
			}
			if v.Delete != nil {
				processOperationV2(v.Delete, &vocab)
			}
		}
	}

	v := &metrics.Vocabulary{
		Schemas:    fillProtoStructures(vocab.schemas),
		Operations: fillProtoStructures(vocab.operationID),
		Parameters: fillProtoStructures(vocab.parameters),
		Properties: fillProtoStructures(vocab.properties),
	}

	return v
}
