// Copyright 2016 Google Inc. All Rights Reserved.
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

//go:generate ./COMPILE-PROTOS.sh

package main

import (
	// "log"
	"regexp"
	"sort"
)

// compiler helper functions, usually called from generated code

func unpackMap(in interface{}) (map[string]interface{}, []string, bool) {
	m, ok := in.(map[string]interface{})
	if !ok {
		return nil, nil, ok
	}
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return m, keys, ok
}

func mapHasKey(m map[string]interface{}, key string) bool {
	_, ok := m[key]
	return ok
}

func convertInterfaceArrayToStringArray(interfaceArray []interface{}) []string {
	stringArray := make([]string, 0)
	for _, item := range interfaceArray {
		v, ok := item.(string)
		if ok {
			stringArray = append(stringArray, v)
		}
	}
	return stringArray
}

func patternMatches(pattern string, value string) bool {
	matched, err := regexp.Match(pattern, []byte(value))
	if err != nil {
		panic(err)
	}
	return matched
}

func mapContainsAllKeys(m map[string]interface{}, keys []string) bool {
	for _, k := range keys {
		_, found := m[k]
		if !found {
			//log.Printf("ERROR: map does not contain required key %s (%+v)", k, m)
			return false
		}
	}
	return true
}

func mapContainsOnlyKeysAndPatterns(m map[string]interface{}, keys []string, patterns []string) bool {
	for k, _ := range m {
		found := false
		// does the key match an allowed key
		for _, k2 := range keys {
			if k == k2 {
				found = true
				break
			}
		}
		if !found {
			// does the key match an allowed pattern?
			for _, pattern := range patterns {
				if patternMatches(pattern, k) {
					//log.Printf("pattern %s matched %s", pattern, k)
					found = true
					break
				}
			}
			if !found {
				//log.Printf("ERROR: map contains unhandled key %s (allowed=%+v) (%+v)", k, keys, m)
				return false
			}
		}
	}
	return true
}
