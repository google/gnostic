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
	"path"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ServiceModel represents an API for code generation.
type ServiceModel struct {
	Name    string           // a free-form title for the API
	Package string           // the name to use for the generated Go package
	Types   []*ServiceType   // the types used by the service
	Methods []*ServiceMethod // the methods (functions) of the service
}

func (m *ServiceModel) typeWithName(name string) *ServiceType {
	for _, t := range m.Types {
		if t.Name == name {
			return t
		}
	}
	return nil
}

func generateOperationName(method, path string) string {
	filteredPath := strings.Replace(path, "/", "_", -1)
	filteredPath = strings.Replace(filteredPath, ".", "_", -1)
	filteredPath = strings.Replace(filteredPath, "{", "", -1)
	filteredPath = strings.Replace(filteredPath, "}", "", -1)
	return upperFirst(method) + filteredPath
}

func cleanupOperationName(name string) string {
	name = strings.Title(name)
	name = strings.Replace(name, ".", "_", -1)
	return name
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

func typeForRef(ref string) (typeName string) {
	return strings.Replace(strings.Title(path.Base(ref)), "-", "_", -1)
}

func propertyNameForResponseCode(code string) string {
	if code == "200" {
		return "OK"
	}
	name := strings.Title(code)
	name = strings.Replace(name, "-", "_", -1)
	return name
}

func snakeCaseToCamelCase(snakeCase string) (camelCase string) {
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
	name = snakeCaseToCamelCase(name)
	return name
}

// convert the first character of a string to upper case
func upperFirst(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + strings.ToLower(s[n:])
}
