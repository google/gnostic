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

// ServiceTypeField is a field in a definition and can be
// associated with a position in a request structure.
type ServiceTypeField struct {
	Name          string // the name as specified in the API description
	Type          string // the specified type of the field
	ValueType     string // if Type is a pointer, this is the type of its value
	NativeType    string // the programming-language native type of the field
	FieldName     string // the name to use for a data structure field
	ParameterName string // the name to use for a function parameter
	JSONName      string // the name to use in JSON serialization
	Position      string // "body", "header", "formdata", "query", or "path"
}

// serviceType returns the ServiceType associated with a field.
func (f *ServiceTypeField) serviceType(m *ServiceModel) *ServiceType {
	return m.typeWithName(f.ValueType)
}