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
)

// mapDifference finds the difference between two Vocabularies.
// This function takes a Vocabulary and checks if the words within
// the current Vocabulary already exist within the global Vocabulary.
// If the word exists in both structures it is removed from the
// Vocabulary structure.
func mapDifference(v *metrics.Vocabulary) {
	for _, s := range v.Schemas {
		_, ok := schemas[s.Word]
		if ok {
			delete(schemas, s.Word)
		}
	}
	for _, op := range v.Operations {
		_, ok := operationID[op.Word]
		if ok {
			delete(operationID, op.Word)
		}
	}
	for _, param := range v.Parameters {
		_, ok := parameters[param.Word]
		if ok {
			delete(parameters, param.Word)
		}
	}
	for _, prop := range v.Properties {
		_, ok := properties[prop.Word]
		if ok {
			delete(properties, prop.Word)
		}
	}
}

// Difference implements the difference operation between multiple Vocabularies.
// The function accepts a slice of Vocabularies and returns a single Vocabulary
// struct which that contains words that were unique to the first Vocabulary in the slice.
func Difference(vocab []*metrics.Vocabulary) *metrics.Vocabulary {
	schemas = make(map[string]int)
	operationID = make(map[string]int)
	parameters = make(map[string]int)
	properties = make(map[string]int)

	unpackageVocabulary(vocab[0])
	for i := 1; i < len(vocab); i++ {
		mapDifference(vocab[i])
	}

	v := &metrics.Vocabulary{
		Properties: fillProtoStructure(properties),
		Schemas:    fillProtoStructure(schemas),
		Operations: fillProtoStructure(operationID),
		Parameters: fillProtoStructure(parameters),
	}
	return v
}
