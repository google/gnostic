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
	"sort"

	metrics "github.com/googleapis/gnostic/metrics"
	openapi_v3 "github.com/googleapis/gnostic/openapiv3"
)

func fillProtoStructures(m map[string]int) []*metrics.WordCount {
	keyNames := make([]string, 0, len(m))
	for key := range m {
		keyNames = append(keyNames, key)
	}
	sort.Strings(keyNames)

	counts := make([]*metrics.WordCount, 0)
	for _, k := range keyNames {
		temp := &metrics.WordCount{
			Word:  k,
			Count: int32(m[k]),
		}
		counts = append(counts, temp)
	}
	return counts
}

func processOperationV3(operation *openapi_v3.Operation, vocab *Vocabulary) {
	if operation.OperationId != "" {
		vocab.operationID[operation.OperationId]++
	}
	for _, item := range operation.Parameters {
		switch t := item.Oneof.(type) {
		case *openapi_v3.ParameterOrReference_Parameter:
			vocab.parameters[t.Parameter.Name]++
		}
	}
}

func processComponentsV3(components *openapi_v3.Components, vocab *Vocabulary) {
	processParametersV3(components, vocab)
	processSchemasV3(components, vocab)
	processResponsesV3(components, vocab)
}

func processParametersV3(components *openapi_v3.Components, vocab *Vocabulary) {
	if components.Parameters == nil {
		return
	}
	for _, pair := range components.Parameters.AdditionalProperties {
		switch t := pair.Value.Oneof.(type) {
		case *openapi_v3.ParameterOrReference_Parameter:
			vocab.parameters[t.Parameter.Name]++
		}
	}
}

func processSchemasV3(components *openapi_v3.Components, vocab *Vocabulary) {
	if components.Schemas == nil {
		return
	}
	for _, pair := range components.Schemas.AdditionalProperties {
		vocab.schemas[pair.Name]++
		processSchemaV3(pair.Value, vocab)
	}
}

func processSchemaV3(schema *openapi_v3.SchemaOrReference, vocab *Vocabulary) {
	if schema == nil {
		return
	}
	switch t := schema.Oneof.(type) {
	case *openapi_v3.SchemaOrReference_Reference:
		return
	case *openapi_v3.SchemaOrReference_Schema:
		if t.Schema.Properties != nil {
			for _, pair := range t.Schema.Properties.AdditionalProperties {
				vocab.properties[pair.Name]++
			}
		}
	}
}

func processResponsesV3(components *openapi_v3.Components, vocab *Vocabulary) {
	if components.Responses == nil {
		return
	}
	for _, pair := range components.Responses.AdditionalProperties {
		vocab.schemas[pair.Name]++
	}
}

func NewVocabularyFromOpenAPIv3(document *openapi_v3.Document) *metrics.Vocabulary {
	var vocab Vocabulary
	vocab.schemas = make(map[string]int)
	vocab.operationID = make(map[string]int)
	vocab.parameters = make(map[string]int)
	vocab.properties = make(map[string]int)

	if document.Components != nil {
		processComponentsV3(document.Components, &vocab)

	}
	for _, pair := range document.Paths.Path {
		v := pair.Value
		if v.Get != nil {
			processOperationV3(v.Get, &vocab)
		}
		if v.Post != nil {
			processOperationV3(v.Post, &vocab)
		}
		if v.Put != nil {
			processOperationV3(v.Put, &vocab)
		}
		if v.Patch != nil {
			processOperationV3(v.Patch, &vocab)
		}
		if v.Delete != nil {
			processOperationV3(v.Delete, &vocab)
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
