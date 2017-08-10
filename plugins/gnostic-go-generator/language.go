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

package main

import (
	surface "github.com/googleapis/gnostic/plugins/gnostic-go-generator/surface"
	"unicode"
	"strings"
)

type GoLanguageModel struct{}

func NewGoLanguageModel() *GoLanguageModel {
	return &GoLanguageModel{}
}

// Prepare sets language-specific properties for all types and methods.
func (language *GoLanguageModel) Prepare(model *surface.Model) {

	for _, t := range model.Types {
		// determine the type used for Go language implementation of the type
		t.TypeName = strings.Title(filteredTypeName(t.Name))

		for _, f := range t.Fields {
			f.FieldName = goFieldName(f.Name)
			f.ParameterName = goParameterName(f.Name)
			switch f.Type {
			case "number":
				f.NativeType = "int"
			case "integer":
				if f.Format == "int32" {
					f.NativeType = "int32"
				} else if f.Format == "int64" {
					f.NativeType = "int64"
				} else {
					f.NativeType = "int64"
				}
			default:
				f.NativeType = f.Type
			}
		}
	}

	for _, m := range model.Methods {
		m.HandlerName = "Handle" + m.Name
		m.ProcessorName = m.Name
		m.ClientName = m.Name
	}
}

func goParameterName(name string) string {
	// lowercase first letter
	a := []rune(name)
	a[0] = unicode.ToLower(a[0])
	name = string(a)
	// replace dots with underscores
	name = strings.Replace(name, ".", "_", -1)
	// avoid reserved words
	if name == "type" {
		return "ttttype"
	}
	return name
}

func goFieldName(name string) string {
	name = strings.Replace(name, ".", "_", -1)
	name = strings.Replace(name, "-", "_", -1)
	name = snakeCaseToCamelCaseWithCapitalizedFirstLetter(name)
	// avoid integers
	if name == "200" {
		return "OK"
	}
	return name
}

func snakeCaseToCamelCaseWithCapitalizedFirstLetter(snakeCase string) (camelCase string) {
	isToUpper := false
	for _, runeValue := range snakeCase {
		if isToUpper {
			camelCase += strings.ToUpper(string(runeValue))
			isToUpper = false
		} else {
			if runeValue == '_' {
				isToUpper = true
			} else {
				camelCase += string(runeValue)
			}
		}
	}
	camelCase = strings.Title(camelCase)
	return
}

func filteredTypeName(typeName string) (name string) {
	// first take the last path segment
	parts := strings.Split(typeName, "/")
	name = parts[len(parts)-1]
	// then take the last part of a dotted name
	parts = strings.Split(name, ".")
	name = parts[len(parts)-1]
	return name
}

func typeForName(name string, format string) (typeName string) {
	switch name {
	case "integer":
		if format == "int32" {
			return "int32"
		} else if format == "int64" {
			return "int64"
		} else {
			return "int32"
		}
	default:
		return name
	}
}
