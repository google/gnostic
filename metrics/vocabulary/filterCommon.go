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

// Difference implements the difference operation between multiple Vocabularies.
// The function accepts a slice of Vocabularies and returns a single Vocabulary
// struct which that contains words that were unique to the first Vocabulary in the slice.

// FilterCommon implements the difference operation among a slice of Vocabularies.
// The function returns a slice of Vocabrularies that contains the unique terms
// for each pb file.
func FilterCommon(v []*metrics.Vocabulary) []*metrics.Vocabulary {
	uniqueVocabularies := make([]*metrics.Vocabulary, 0)
	n := len(v)
	for x := 0; x < n; x++ {
		temp := make([]*metrics.Vocabulary, 0)
		temp = append(temp, v[x])
		for y := 0; y < n; y++ {
			if x == y {
				continue
			}
			temp = append(temp, v[y])
		}
		vocab := Difference(temp)
		uniqueVocabularies = append(uniqueVocabularies, vocab)
	}
	return uniqueVocabularies
}
