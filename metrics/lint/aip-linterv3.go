// Copyright 2020 Google LLC. All Rights Reserved.
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

package linter

import (
	"io/ioutil"
	"strings"

	"github.com/golang/protobuf/proto"

	rules "github.com/googleapis/gnostic/metrics/rules"
	openapi_v3 "github.com/googleapis/gnostic/openapiv3"
)

func readDocumentFromFileWithName(filename string) (*openapi_v3.Document, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	document := &openapi_v3.Document{}
	err = proto.Unmarshal(data, document)
	if err != nil {
		return nil, err
	}
	return document, nil
}
func processParametersV3(components *openapi_v3.Components, path []string) []rules.Field {
	parameters := make([]rules.Field, 0)
	if components.Parameters != nil {
		for _, pair := range components.Parameters.AdditionalProperties {
			switch t := pair.Value.Oneof.(type) {
			case *openapi_v3.ParameterOrReference_Parameter:
				parameters = append(parameters, rules.Field{Name: t.Parameter.Name, Path: path})

			}
		}
	}
	return parameters
}

func processOperationV3(operation *openapi_v3.Operation, path []string) []rules.Field {
	parameters := make([]rules.Field, 0)
	for _, item := range operation.Parameters {
		switch t := item.Oneof.(type) {
		case *openapi_v3.ParameterOrReference_Parameter:
			parameters = append(parameters, rules.Field{Name: t.Parameter.Name, Path: path})

		}
	}
	return parameters
}

func gatherParameters(document *openapi_v3.Document) []rules.Field {
	p := make([]rules.Field, 0)

	if document.Components != nil {
		path := []string{"components", "parameters"}
		p = append(p, processParametersV3(document.Components, path)...)
	}

	if document.Paths != nil {
		for _, pair := range document.Paths.Path {
			v := pair.Value
			path := strings.Split("paths"+pair.Name, "/")
			if v.Get != nil {
				p = append(p, processOperationV3(v.Get, path)...)
			}
			if v.Post != nil {
				p = append(p, processOperationV3(v.Post, path)...)
			}
			if v.Put != nil {
				p = append(p, processOperationV3(v.Put, path)...)
			}
			if v.Patch != nil {
				p = append(p, processOperationV3(v.Patch, path)...)
			}
			if v.Delete != nil {
				p = append(p, processOperationV3(v.Delete, path)...)
			}
		}
	}
	return p
}

//AIPLintV3 accepts an OpenAPI v2 document and will call the individual AIP rules
//on the document.
func AIPLintV3(document *openapi_v3.Document) (*Linter, int) {
	fields := gatherParameters(document)
	messages := make([]rules.MessageType, 0)
	for _, field := range fields {
		messages = append(messages, rules.AIP122Driver(field)...)
		messages = append(messages, rules.AIP140Driver(field)...)
	}
	m := fillProtoStructure(messages)

	linterResult := &Linter{
		Messages: m,
	}
	return linterResult, len(messages)
}
