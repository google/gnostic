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

package jsonschema

import (
	"fmt"
	"gopkg.in/yaml.v2"
)

const INDENT = "  "

func renderMap(info interface{}, indent string) (result string) {
	result = "{\n"
	inner_indent := indent + INDENT
	switch pairs := info.(type) {
	case yaml.MapSlice:
		for i, pair := range pairs {
			// first print the key
			result += fmt.Sprintf("%s\"%+v\": ", inner_indent, pair.Key)
			// then the value
			switch value := pair.Value.(type) {
			case string:
				result += "\"" + value + "\""
			case bool:
				if value {
					result += "true"
				} else {
					result += "false"
				}
			case []interface{}:
				result += renderArray(value, inner_indent)
			case yaml.MapSlice:
				result += renderMap(value, inner_indent)
			case int:
				result += fmt.Sprintf("%d", value)
			default:
				result += fmt.Sprintf("???MapItem(%+v)", value)
			}
			if i < len(pairs)-1 {
				result += ","
			}
			result += "\n"
		}
	default:
		// t is some other type that we didn't name.
	}

	result += indent + "}"
	return result
}

func renderArray(array []interface{}, indent string) (result string) {
	result = "[\n"
	inner_indent := indent + INDENT
	for i, item := range array {
		switch item := item.(type) {
		case string:
			result += inner_indent + "\"" + item + "\""
		case bool:
			if item {
				result += inner_indent + "true"
			} else {
				result += inner_indent + "false"
			}
		case yaml.MapSlice:
			result += inner_indent + renderMap(item, inner_indent) + ""
		default:
			result += inner_indent + fmt.Sprintf("???ArrayItem(%+v)", item)
		}
		if i < len(array)-1 {
			result += ","
		}
		result += "\n"
	}
	result += indent + "]"
	return result
}

func Render(info yaml.MapSlice) string {
	return renderMap(info, "") + "\n"
}
