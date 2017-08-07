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

package gnostic_surface_v1

const newline = "\n"

func (m *Method) HasParameters() bool {
	return m.ParametersType != nil
}

func (m *Method) HasResponses() bool {
	return m.ResponsesType != nil
}

func (m *Method) ParameterList() string {
	result := ""
	if m.ParametersType != nil {
		for _, field := range m.ParametersType.Fields {
			result += field.ParameterName + " " + field.NativeType + "," + newline
		}
	}
	return result
}

func (m *Method) BodyParameterName() string {
	if m.ParametersType != nil {
		for _, field := range m.ParametersType.Fields {
			if field.Position == Position_BODY {
				return field.JSONName
			}
		}
	}
	return ""
}

func (m *Method) BodyParameterFieldName() string {
	for _, field := range m.ParametersType.Fields {
		if field.Position == Position_BODY {
			return field.FieldName
		}
	}
	return ""
}

func (m *Method) HasParametersWithPosition(position Position) bool {
	if m.ParametersType != nil {
		for _, field := range m.ParametersType.Fields {
			if field.Position == position {
				return true
			}
		}
	}
	return false
}