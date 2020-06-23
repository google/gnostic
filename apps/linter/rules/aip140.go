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

package rules

import (
	"regexp"
	"sort"
	"strings"

	"github.com/stoewer/go-strcase"
)

func snakeCase(field string) (bool, string) {
	snake := strcase.SnakeCase(field)
	snake = strings.ToLower(snake)

	return snake == field, snake
}

func abbreviation(field string) (bool, string) {
	var expectedAbbreviations = map[string]string{
		"configuration": "config",
		"identifier":    "id",
		"information":   "info",
		"specification": "spec",
		"statistics":    "stats",
	}

	if suggestion, exists := expectedAbbreviations[field]; exists {
		return true, suggestion
	}
	return false, field
}

func numbers(field string) bool {
	var numberStart = regexp.MustCompile("^[0-9]")
	for _, segment := range strings.Split(field, "_") {
		if numberStart.MatchString(segment) {
			return true
		}
	}
	return false
}

func reservedWords(field string) bool {
	reservedWordsSet := []string{"abstract", "and", "arguments", "as", "assert", "async", "await", "boolean", "break", "byte",
		"case", "catch", "char", "class", "const", "continue", "debugger", "def", "default", "del", "delete", "do", "double", "elif",
		"else", "enum", "eval", "except", "export", "extends", "false", "final", "finally", "float", "for", "from", "function", "global",
		"goto", "if", "implements", "import", "in", "instanceof", "int", "interface", "is", "lambda", "let", "long", "native", "new", "nonlocal",
		"not", "null", "or", "package", "pass", "private", "protected", "public", "raise", "return", "short", "static", "strictfp",
		"super", "switch", "synchronized", "this", "throw", "throws", "transient", "true", "try", "typeof", "var", "void", "volatile",
		"while", "with", "yield"}

	for _, segment := range strings.Split(field, "_") {
		result := sort.SearchStrings(reservedWordsSet, segment)
		if result < len(reservedWordsSet) && reservedWordsSet[result] == segment {
			return true
		}
	}
	return false
}

func prepositions(field string) bool {
	preps := []string{"after", "at", "before", "between", "but", "by", "except",
		"for", "from", "in", "including", "into", "of", "over", "since", "to",
		"toward", "under", "upon", "with", "within", "without"}
	for _, segment := range strings.Split(field, "_") {
		result := sort.SearchStrings(preps, segment)
		if result < len(preps) && preps[result] == segment {
			return true
		}
	}
	return false
}

func AIP140Driver(field string) {

}
