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

// ServiceMethod is an operation of an API and typically
// has associated client and server code.
type ServiceMethod struct {
	Name               string       // Operation name, possibly generated from method and path
	Path               string       // HTTP path
	Method             string       // HTTP method name
	Description        string       // description of method
	HandlerName        string       // name of the generated handler
	ProcessorName      string       // name of the processing function in the service interface
	ClientName         string       // name of client
	ResultTypeName     string       // native type name for the result structure
	ParametersTypeName string       // native type name for the input parameters structure
	ResponsesTypeName  string       // native type name for the responses
	ParametersType     *ServiceType // parameters (input), with fields corresponding to input parameters
	ResponsesType      *ServiceType // responses (output), with fields corresponding to possible response values
}

func (m *ServiceMethod) hasParameters() bool {
	return m.ParametersType != nil
}

func (m *ServiceMethod) hasResponses() bool {
	return m.ResponsesType != nil
}

func (m *ServiceMethod) parameterList() string {
	result := ""
	if m.ParametersType != nil {
		for _, field := range m.ParametersType.Fields {
			result += field.ParameterName + " " + field.NativeType + "," + newline
		}
	}
	return result
}

func (m *ServiceMethod) bodyParameterName() string {
	if m.ParametersType != nil {
		for _, field := range m.ParametersType.Fields {
			if field.Position == "body" {
				return field.JSONName
			}
		}
	}
	return ""
}

func (m *ServiceMethod) bodyParameterFieldName() string {
	for _, field := range m.ParametersType.Fields {
		if field.Position == "body" {
			return field.FieldName
		}
	}
	return ""
}

func (m *ServiceMethod) hasParametersWithPosition(position string) bool {
	if m.ParametersType != nil {
		for _, field := range m.ParametersType.Fields {
			if field.Position == position {
				return true
			}
		}
	}
	return false
}