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

// ServiceType typically corresponds to a definition, parameter,
// or response in the API and is represented by a type in generated code.
type ServiceType struct {
	Name        string              // the name to use for the type
	Kind        string              // a "meta" description of the type (struct, map, etc)
	Description string              // a comment describing the type
	Fields      []*ServiceTypeField // the fields of the type
}

func (s *ServiceType) hasFieldWithName(name string) bool {
	if s == nil || s.Fields == nil {
		return false
	}
	for _, f := range s.Fields {
		if f.FieldName == name {
			return true
		}
	}
	return false
}

func (s *ServiceType) fieldWithName(name string) *ServiceTypeField {
	if s == nil || s.Fields == nil {
		return nil
	}
	for _, f := range s.Fields {
		if f.FieldName == name {
			return f
		}
	}
	return nil
}
