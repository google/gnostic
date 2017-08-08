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
	"log"
)

type GoLanguageModel struct{}

func NewGoLanguageModel() *GoLanguageModel {
	return &GoLanguageModel{}
}

// Prepare sets language-specific properties for all types and methods.
func (language *GoLanguageModel) Prepare(model *surface.Model) {

	for _, t := range model.Types {
		log.Printf("%s", t.Name)
		for _, f := range t.Fields {
			log.Printf("  %s %s %s", f.Name, f.Type, f.Format)

			f.FieldName = goFieldName(f.Name)
			f.ParameterName = goParameterName(f.Name)

			f.NativeType = f.Type
			if f.NativeType == "number" {
				f.NativeType = "int"
			} else if f.NativeType == "integer" {
				f.NativeType = "int64"
			}
		}
	}

	for _, m := range model.Methods {
		log.Printf("%s %s", m.Method, m.Path)

		p1 := "?"
		p2 := "?"
		if m.ParametersType != nil {
			p1 = m.ParametersType.Name
		}
		if m.ResponsesType != nil {
			p2 = m.ResponsesType.Name
		}

		log.Printf("  %s %s", p1, p2)
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

// unused
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
