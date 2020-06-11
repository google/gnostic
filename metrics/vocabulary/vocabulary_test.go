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

func testVocabulary(t *testing.T, outputVocab *metrics.Vocabulary, referencePb *metrics.Vocabulary) {
	results := Difference([]*metrics.Vocabulary{outputVocab, referencePb})
	results2 := Difference([]*metrics.Vocabulary{referencePb, outputVocab})

	if !isEmpty(results) && !isEmpty(results2) {
		t.Logf("Difference failed: Output does not match")
		t.FailNow()
	} else {
		// if the test succeeded, clean up
		os.Remove("vocabulary-operation.pb")
	}
}

func TestSampleVocabularyUnion(t *testing.T) {
	v1 := metrics.Vocabulary{
		Schemas:    fillTestProtoStructue([]string{"heelo", "random", "funcName", "google"}, []int{1, 2, 3, 4}),
		Properties: fillTestProtoStructue([]string{"Hello", "dog", "funcName", "cat"}, []int{4, 3, 2, 1}),
		Operations: fillTestProtoStructue([]string{"countGreetings", "print", "funcName"}, []int{12, 11, 4}),
		Parameters: fillTestProtoStructue([]string{"name", "id", "tag", "suggester"}, []int{5, 1, 1, 15}),
	}

	v2 := metrics.Vocabulary{
		Schemas:    fillTestProtoStructue([]string{"Hello", "random", "status", "google"}, []int{5, 6, 1, 4}),
		Properties: fillTestProtoStructue([]string{"cat", "dog", "thing"}, []int{4, 3, 2}),
		Operations: fillTestProtoStructue([]string{"countPrint", "print", "funcName"}, []int{17, 12, 19}),
		Parameters: fillTestProtoStructue([]string{"name", "id", "tag", "suggester"}, []int{5, 1, 1, 15}),
	}

	vocabularies := make([]*metrics.Vocabulary, 0)
	vocabularies = append(vocabularies, &v1, &v2)

	reference := metrics.Vocabulary{
		Schemas:    fillTestProtoStructue([]string{"Hello", "funcName", "google", "heelo", "random", "status"}, []int{5, 3, 8, 1, 8, 1}),
		Properties: fillTestProtoStructue([]string{"Hello", "cat", "dog", "funcName", "thing"}, []int{4, 5, 6, 2, 2}),
		Operations: fillTestProtoStructue([]string{"countGreetings", "countPrint", "funcName", "print"}, []int{12, 17, 23, 23}),
		Parameters: fillTestProtoStructue([]string{"id", "name", "suggester", "tag"}, []int{2, 10, 30, 2}),
	}

	unionResult := Union(vocabularies)

	testVocabulary(t,
		unionResult,
		&reference,
	)
}

func TestSampleVocabularyIntersection(t *testing.T) {
	v1 := metrics.Vocabulary{
		Schemas:    fillTestProtoStructue([]string{"heelo", "random", "funcName", "google"}, []int{1, 2, 3, 4}),
		Properties: fillTestProtoStructue([]string{"Hello", "dog", "funcName", "cat"}, []int{4, 3, 2, 1}),
		Operations: fillTestProtoStructue([]string{"countGreetings", "print", "funcName"}, []int{12, 11, 4}),
		Parameters: fillTestProtoStructue([]string{"name", "id", "tag", "suggester"}, []int{5, 1, 1, 15}),
	}

	v2 := metrics.Vocabulary{
		Schemas:    fillTestProtoStructue([]string{"Hello", "random", "status", "google"}, []int{5, 6, 1, 4}),
		Properties: fillTestProtoStructue([]string{"cat", "dog", "thing"}, []int{4, 3, 2}),
		Operations: fillTestProtoStructue([]string{"countPrint", "print", "funcName"}, []int{17, 12, 19}),
		Parameters: fillTestProtoStructue([]string{"name", "id", "tag", "suggester"}, []int{5, 1, 1, 15}),
	}

	vocabularies := make([]*metrics.Vocabulary, 0)
	vocabularies = append(vocabularies, &v1, &v2)

	reference := metrics.Vocabulary{
		Schemas:    fillTestProtoStructue([]string{"google", "random"}, []int{8, 8}),
		Properties: fillTestProtoStructue([]string{"cat", "dog"}, []int{5, 6}),
		Operations: fillTestProtoStructue([]string{"funcName", "print"}, []int{23, 23}),
		Parameters: fillTestProtoStructue([]string{"id", "name", "suggester", "tag"}, []int{2, 10, 30, 2}),
	}

	intersectionResult := Intersection(vocabularies)

	testVocabulary(t,
		intersectionResult,
		&reference,
	)
}
func TestSampleVocabularyDifference(t *testing.T) {
	v1 := metrics.Vocabulary{
		Schemas:    fillTestProtoStructue([]string{"heelo", "random", "funcName", "google"}, []int{1, 2, 3, 4}),
		Properties: fillTestProtoStructue([]string{"Hello", "dog", "funcName", "cat"}, []int{4, 3, 2, 1}),
		Operations: fillTestProtoStructue([]string{"countGreetings", "print", "funcName"}, []int{12, 11, 4}),
		Parameters: fillTestProtoStructue([]string{"name", "id", "tag", "suggester"}, []int{5, 1, 1, 15}),
	}

	v2 := metrics.Vocabulary{
		Schemas:    fillTestProtoStructue([]string{"Hello", "random", "status", "google"}, []int{5, 6, 1, 4}),
		Properties: fillTestProtoStructue([]string{"cat", "dog", "thing"}, []int{4, 3, 2}),
		Operations: fillTestProtoStructue([]string{"countPrint", "print", "funcName"}, []int{17, 12, 19}),
		Parameters: fillTestProtoStructue([]string{"name", "id", "tag", "suggester"}, []int{5, 1, 1, 15}),
	}

	vocabularies := make([]*metrics.Vocabulary, 0)
	vocabularies = append(vocabularies, &v1, &v2)

	reference := metrics.Vocabulary{
		Schemas:    fillTestProtoStructue([]string{"funcName", "heelo"}, []int{3, 1}),
		Properties: fillTestProtoStructue([]string{"Hello", "funcName"}, []int{4, 2}),
		Operations: fillTestProtoStructue([]string{"countGreetings"}, []int{12}),
	}

	differenceResult := Difference(vocabularies)

	testVocabulary(t,
		differenceResult,
		&reference,
	)
}
