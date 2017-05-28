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

package jsonwriter

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
)

const INDENT = "  "

// basic escaping, will need to be improved or replaced
func escape(s string) string {
	s = strings.Replace(s, "\n", "\\n", -1)
	s = strings.Replace(s, "\"", "\\\"", -1)
	return s
}

type Writer struct {
	b bytes.Buffer
}

func (w *Writer) bytes() []byte {
	return w.b.Bytes()
}

func (w *Writer) writeString(s string) {
	w.b.Write([]byte(s))
}

func (w *Writer) writeMap(info interface{}, indent string) {
	w.writeString("{\n")
	inner_indent := indent + INDENT
	switch pairs := info.(type) {
	case yaml.MapSlice:
		for i, pair := range pairs {
			// first print the key
			w.writeString(fmt.Sprintf("%s\"%+v\": ", inner_indent, pair.Key))
			// then the value
			switch value := pair.Value.(type) {
			case string:
				w.writeString("\"")
				w.writeString(escape(value))
				w.writeString("\"")
			case bool:
				if value {
					w.writeString("true")
				} else {
					w.writeString("false")
				}
			case []interface{}:
				w.writeArray(value, inner_indent)
			case yaml.MapSlice:
				w.writeMap(value, inner_indent)
			case int:
				w.writeString(fmt.Sprintf("%d", value))
			case int64:
				w.writeString(fmt.Sprintf("%d", value))
			case []string:
				w.writeStringArray(value, inner_indent)
			case float64:
				w.writeString(fmt.Sprintf("%f", value))
			case []yaml.MapSlice:
				w.writeMapSliceArray(value, inner_indent)
			default:
				w.writeString(fmt.Sprintf("???MapItem(%+v, %T)", value, value))
			}
			if i < len(pairs)-1 {
				w.writeString(",")
			}
			w.writeString("\n")
		}
	default:
		// t is some other type that we didn't name.
	}
	w.writeString(indent)
	w.writeString("}")
}

func (w *Writer) writeArray(array []interface{}, indent string) {
	w.writeString("[\n")
	inner_indent := indent + INDENT
	for i, item := range array {
		w.writeString(inner_indent)
		switch item := item.(type) {
		case string:
			w.writeString("\"")
			w.writeString(item)
			w.writeString("\"")
		case bool:
			if item {
				w.writeString("true")
			} else {
				w.writeString("false")
			}
		case yaml.MapSlice:
			w.writeMap(item, inner_indent)
		default:
			w.writeString(fmt.Sprintf("???ArrayItem(%+v)", item))
		}
		if i < len(array)-1 {
			w.writeString(",")
		}
		w.writeString("\n")
	}
	w.writeString(indent)
	w.writeString("]")
}

func (w *Writer) writeStringArray(array []string, indent string) {
	w.writeString("[\n")
	inner_indent := indent + INDENT
	for i, item := range array {
		w.writeString(inner_indent)
		w.writeString("\"")
		w.writeString(escape(item))
		w.writeString("\"")
		if i < len(array)-1 {
			w.writeString(",")
		}
		w.writeString("\n")
	}
	w.writeString(indent)
	w.writeString("]")
}

func (w *Writer) writeMapSliceArray(array []yaml.MapSlice, indent string) {
	w.writeString("[\n")
	inner_indent := indent + INDENT
	for i, item := range array {
		w.writeString(inner_indent)
		w.writeMap(item, inner_indent)
		if i < len(array)-1 {
			w.writeString(",")
		}
		w.writeString("\n")
	}
	w.writeString(indent)
	w.writeString("]")
}

func Marshal(in interface{}) (out []byte, err error) {
	var w Writer
	m, ok := in.(yaml.MapSlice)
	if !ok {
		return nil, errors.New("invalid type passed to Marshal")
	}
	w.writeMap(m, "")
	w.writeString("\n")
	return w.bytes(), err
}
