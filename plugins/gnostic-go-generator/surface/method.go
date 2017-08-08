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

// HasParameters returns true if a method has parameters
func (m *Method) HasParameters() bool {
	return m.ParametersType != nil
}

// HasResponses returns true if a method has responses
func (m *Method) HasResponses() bool {
	return m.ResponsesType != nil
}

// BodyParameterField returns the body parameter field of a method, if one is present
func (m *Method) BodyParameterField() *Field {
	if m.ParametersType != nil {
		for _, field := range m.ParametersType.Fields {
			if field.Position == Position_BODY {
				return field
			}
		}
	}
	return nil
}

// HasParametersWithPosition returns true if the method has parameters in a specified position.
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