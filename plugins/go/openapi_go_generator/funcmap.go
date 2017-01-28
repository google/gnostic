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
	"text/template"
)

// This file contains support functions that are passed into template
// evaluation for use within templates.

func hasFieldNamedOK(s *ServiceType) bool {
	return s.hasFieldNamed("OK")
}

func parameterList(m *ServiceMethod) string {
	result := ""
	for i, field := range m.ParametersType.Fields {
		if i > 0 {
			result += ", "
		}
		result += field.JSONName + " " + field.Type
	}
	return result
}

func bodyParameterName(m *ServiceMethod) string {
	for _, field := range m.ParametersType.Fields {
		if field.Position == "body" {
			return field.JSONName
		}
	}
	return ""
}

func bodyParameterFieldName(m *ServiceMethod) string {
	for _, field := range m.ParametersType.Fields {
		if field.Position == "body" {
			return field.Name
		}
	}
	return ""
}

func helpers() template.FuncMap {
	return template.FuncMap{
		"hasFieldNamedOK":        hasFieldNamedOK,
		"parameterList":          parameterList,
		"bodyParameterName":      bodyParameterName,
		"bodyParameterFieldName": bodyParameterFieldName,
	}
}
