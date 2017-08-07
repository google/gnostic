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
)

func (renderer *ServiceRenderer) GenerateTypes() ([]byte, error) {
	f := NewLineWriter()
	f.WriteLine(`// GENERATED FILE: DO NOT EDIT!`)
	f.WriteLine(``)
	f.WriteLine(`package ` + renderer.Model.Package)
	f.WriteLine(`// Types used by the API.`)
	for _, modelType := range renderer.Model.Types {
		f.WriteLine(`// ` + modelType.Description)
		if modelType.Kind == "struct" {
			f.WriteLine(`type ` + modelType.Name + ` struct {`)
			for _, field := range modelType.Fields {
				f.WriteLine(field.FieldName + ` ` + goType(field.Type) + jsonTag(field))
			}
			f.WriteLine(`}`)
		} else if modelType.Kind != "" {
			f.WriteLine(`type ` + modelType.Name + ` ` + modelType.Kind)
		} else {
			f.WriteLine(`type ` + modelType.Name + ` struct {}`)
		}
	}
	return f.Bytes(), nil
}

func jsonTag(field *surface.Field) string {
	if field.JSONName != "" {
		return " `json:" + `"` + field.JSONName + `,omitempty"` + "`"
	}
	return ""
}

func goType(openapiType string) string {
	switch openapiType {
	case "number":
		return "int"
	default:
		return openapiType
	}
}