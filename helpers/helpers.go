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

package helpers

import (
	// "log"
	"gopkg.in/yaml.v2"
	"regexp"
)

// compiler helper functions, usually called from generated code

func UnpackMap(in interface{}) (yaml.MapSlice, bool) {
	m, ok := in.(yaml.MapSlice)
	if !ok {
		return nil, ok
	}
	return m, ok
}

func MapHasKey(m yaml.MapSlice, key string) bool {
	for _, item := range m {
		if key == item.Key.(string) {
			return true
		}
	}
	return false
}

func MapValueForKey(m yaml.MapSlice, key string) interface{} {
	for _, item := range m {
		if key == item.Key.(string) {
			return item.Value
		}
	}
	return nil
}

func ConvertInterfaceArrayToStringArray(interfaceArray []interface{}) []string {
	stringArray := make([]string, 0)
	for _, item := range interfaceArray {
		v, ok := item.(string)
		if ok {
			stringArray = append(stringArray, v)
		}
	}
	return stringArray
}

func PatternMatches(pattern string, value string) bool {
	matched, err := regexp.Match(pattern, []byte(value))
	if err != nil {
		panic(err)
	}
	return matched
}

func MapContainsAllKeys(m yaml.MapSlice, keys []string) bool {
	for _, k := range keys {
		if !MapHasKey(m, k) {
			//log.Printf("ERROR: map does not contain required key %s (%+v)", k, m)
			return false
		}
	}
	return true
}

func MapContainsOnlyKeysAndPatterns(m yaml.MapSlice, keys []string, patterns []string) bool {
	for _, item := range m {
		k := item.Key.(string)
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
				if PatternMatches(pattern, k) {
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
