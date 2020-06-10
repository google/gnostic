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

package vocabulary

import (
	"os"
	"testing"

	metrics "github.com/googleapis/gnostic/metrics"
)

func fillTestProtoStructue(words []string, count []int) []*metrics.WordCount {
	counts := make([]*metrics.WordCount, 0)
	for i := 0; i < len(words); i++ {
		temp := &metrics.WordCount{
			Word:  words[i],
			Count: int32(count[i]),
		}
		counts = append(counts, temp)
	}
	return counts
}

func testVocabulary(t *testing.T, operation func(a []*metrics.Vocabulary) *metrics.Vocabulary, inputVocab []*metrics.Vocabulary, referencePb *metrics.Vocabulary) {
	// remove any preexisting output files
	// run the compiler
	var err error
	output := operation(inputVocab)
	if err != nil {
		t.Logf("Compile failed: %+v", err)
		t.FailNow()
	}
	compare := make([]*metrics.Vocabulary, 0)
	compare = append(compare, output)
	compare = append(compare, referencePb)
	results := VocabularyDifference(compare)

	if !isEmpty(results) {
		t.Logf("Difference failed: Output does not match")
		t.FailNow()
	} else {
		// if the test succeeded, clean up
		os.Remove("vocabulary-operation.pb")
	}
}

func TestSampleVocabularyUnion(t *testing.T) {
	schemaWords := []string{"heelo", "random", "funcName", "google"}
	schemaCount := []int{1, 2, 3, 4}
	propWords := []string{"Hello", "dog", "funcName", "cat"}
	propCount := []int{4, 3, 2, 1}
	opsWords := []string{"countGreetings", "print", "funcName"}
	opsCount := []int{12, 11, 4}
	paramWords := []string{"name", "id", "tag", "suggester"}
	paramCount := []int{5, 1, 1, 15}

	v1 := metrics.Vocabulary{
		Schemas:    fillTestProtoStructue(schemaWords, schemaCount),
		Properties: fillTestProtoStructue(propWords, propCount),
		Operations: fillTestProtoStructue(opsWords, opsCount),
		Parameters: fillTestProtoStructue(paramWords, paramCount),
	}

	schemaWords2 := []string{"Hello", "random", "status", "google"}
	schemaCount2 := []int{5, 6, 1, 4}
	propWords2 := []string{"cat", "dog", "thing"}
	propCount2 := []int{4, 3, 2}
	opsWords2 := []string{"countPrint", "print", "funcName"}
	opsCount2 := []int{17, 12, 19}
	paramWords2 := []string{"name", "id", "tag", "suggester"}
	paramCount2 := []int{5, 1, 1, 15}

	v2 := metrics.Vocabulary{
		Schemas:    fillTestProtoStructue(schemaWords2, schemaCount2),
		Properties: fillTestProtoStructue(propWords2, propCount2),
		Operations: fillTestProtoStructue(opsWords2, opsCount2),
		Parameters: fillTestProtoStructue(paramWords2, paramCount2),
	}

	vocab := make([]*metrics.Vocabulary, 0)
	vocab = append(vocab, &v1, &v2)

	schemaWords3 := []string{"Hello", "funcName", "google", "heelo", "random", "status"}
	schemaCount3 := []int{5, 3, 8, 1, 8, 1}
	propWords3 := []string{"Hello", "cat", "dog", "funcName", "thing"}
	propCount3 := []int{4, 5, 6, 2, 2}
	opsWords3 := []string{"countGreetings", "countPrint", "funcName", "print"}
	opsCount3 := []int{12, 17, 23, 23}
	paramWords3 := []string{"id", "name", "suggester", "tag"}
	paramCount3 := []int{2, 10, 30, 2}

	reference := metrics.Vocabulary{
		Schemas:    fillTestProtoStructue(schemaWords3, schemaCount3),
		Properties: fillTestProtoStructue(propWords3, propCount3),
		Operations: fillTestProtoStructue(opsWords3, opsCount3),
		Parameters: fillTestProtoStructue(paramWords3, paramCount3),
	}

	f := func(v []*metrics.Vocabulary) *metrics.Vocabulary {
		return VocabularyUnion(vocab)
	}

	testVocabulary(t,
		f,
		vocab,
		&reference,
	)
}

func TestSampleVocabularyIntersection(t *testing.T) {
	schemaWords := []string{"heelo", "random", "funcName", "google"}
	schemaCount := []int{1, 2, 3, 4}
	propWords := []string{"Hello", "dog", "funcName", "cat"}
	propCount := []int{4, 3, 2, 1}
	opsWords := []string{"countGreetings", "print", "funcName"}
	opsCount := []int{12, 11, 4}
	paramWords := []string{"name", "id", "tag", "suggester"}
	paramCount := []int{5, 1, 1, 15}

	v1 := metrics.Vocabulary{
		Schemas:    fillTestProtoStructue(schemaWords, schemaCount),
		Properties: fillTestProtoStructue(propWords, propCount),
		Operations: fillTestProtoStructue(opsWords, opsCount),
		Parameters: fillTestProtoStructue(paramWords, paramCount),
	}

	schemaWords2 := []string{"Hello", "random", "status", "google"}
	schemaCount2 := []int{5, 6, 1, 4}
	propWords2 := []string{"cat", "dog", "thing"}
	propCount2 := []int{4, 3, 2}
	opsWords2 := []string{"countPrint", "print", "funcName"}
	opsCount2 := []int{17, 12, 19}
	paramWords2 := []string{"name", "id", "tag", "suggester"}
	paramCount2 := []int{5, 1, 1, 15}

	v2 := metrics.Vocabulary{
		Schemas:    fillTestProtoStructue(schemaWords2, schemaCount2),
		Properties: fillTestProtoStructue(propWords2, propCount2),
		Operations: fillTestProtoStructue(opsWords2, opsCount2),
		Parameters: fillTestProtoStructue(paramWords2, paramCount2),
	}

	vocab := make([]*metrics.Vocabulary, 0)
	vocab = append(vocab, &v1, &v2)

	schemaWords3 := []string{"google", "random"}
	schemaCount3 := []int{8, 8}
	propWords3 := []string{"cat", "dog"}
	propCount3 := []int{5, 6}
	opsWords3 := []string{"funcName", "print"}
	opsCount3 := []int{23, 23}
	paramWords3 := []string{"id", "name", "suggester", "tag"}
	paramCount3 := []int{2, 10, 30, 2}

	reference := metrics.Vocabulary{
		Schemas:    fillTestProtoStructue(schemaWords3, schemaCount3),
		Properties: fillTestProtoStructue(propWords3, propCount3),
		Operations: fillTestProtoStructue(opsWords3, opsCount3),
		Parameters: fillTestProtoStructue(paramWords3, paramCount3),
	}

	f := func(v []*metrics.Vocabulary) *metrics.Vocabulary {
		return VocabularyIntersection(vocab)
	}

	testVocabulary(t,
		f,
		vocab,
		&reference,
	)
}
func TestSampleVocabularyDifference(t *testing.T) {
	schemaWords := []string{"heelo", "random", "funcName", "google"}
	schemaCount := []int{1, 2, 3, 4}
	propWords := []string{"Hello", "dog", "funcName", "cat"}
	propCount := []int{4, 3, 2, 1}
	opsWords := []string{"countGreetings", "print", "funcName"}
	opsCount := []int{12, 11, 4}
	paramWords := []string{"name", "id", "tag", "suggester"}
	paramCount := []int{5, 1, 1, 15}

	v1 := metrics.Vocabulary{
		Schemas:    fillTestProtoStructue(schemaWords, schemaCount),
		Properties: fillTestProtoStructue(propWords, propCount),
		Operations: fillTestProtoStructue(opsWords, opsCount),
		Parameters: fillTestProtoStructue(paramWords, paramCount),
	}

	schemaWords2 := []string{"Hello", "random", "status", "google"}
	schemaCount2 := []int{5, 6, 1, 4}
	propWords2 := []string{"cat", "dog", "thing"}
	propCount2 := []int{4, 3, 2}
	opsWords2 := []string{"countPrint", "print", "funcName"}
	opsCount2 := []int{17, 12, 19}
	paramWords2 := []string{"name", "id", "tag", "suggester"}
	paramCount2 := []int{5, 1, 1, 15}

	v2 := metrics.Vocabulary{
		Schemas:    fillTestProtoStructue(schemaWords2, schemaCount2),
		Properties: fillTestProtoStructue(propWords2, propCount2),
		Operations: fillTestProtoStructue(opsWords2, opsCount2),
		Parameters: fillTestProtoStructue(paramWords2, paramCount2),
	}

	vocab := make([]*metrics.Vocabulary, 0)
	vocab = append(vocab, &v1, &v2)

	schemaWords3 := []string{"funcName", "heelo"}
	schemaCount3 := []int{3, 1}
	propWords3 := []string{"Hello", "funcName"}
	propCount3 := []int{4, 2}
	opsWords3 := []string{"countGreetings"}
	opsCount3 := []int{12}

	reference := metrics.Vocabulary{
		Schemas:    fillTestProtoStructue(schemaWords3, schemaCount3),
		Properties: fillTestProtoStructue(propWords3, propCount3),
		Operations: fillTestProtoStructue(opsWords3, opsCount3),
	}

	f := func(v []*metrics.Vocabulary) *metrics.Vocabulary {
		return VocabularyDifference(vocab)
	}

	testVocabulary(t,
		f,
		vocab,
		&reference,
	)
}
