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
	"strings"

	rules "github.com/googleapis/gnostic/metrics/rules"
	pb "github.com/googleapis/gnostic/openapiv2"
)

func fillProtoStructure(m []rules.MessageType) []*Message {
	messages := make([]*Message, 0)
	for _, message := range m {
		temp := &Message{
			Type:    message.Message[0],
			Message: message.Message[1],
			Keys:    message.Path,
		}
		if message.Message[2] != "" {
			temp.Suggestion = message.Message[2]
		}
		messages = append(messages, temp)
	}
	return messages
}

func gatherParametersV2(document *pb.Document) []rules.Field {
	p := make([]rules.Field, 0)
	if document.Paths != nil {
		for _, pair := range document.Paths.Path {
			v := pair.Value
			path := strings.Split("paths"+pair.Name, "/")
			if v.Get != nil {
				p = append(p, processParametersV2(v.Get, path)...)
			}
			if v.Put != nil {
				p = append(p, processParametersV2(v.Put, path)...)
			}
			if v.Post != nil {
				p = append(p, processParametersV2(v.Post, path)...)
			}
			if v.Delete != nil {
				p = append(p, processParametersV2(v.Delete, path)...)
			}
			if v.Patch != nil {
				p = append(p, processParametersV2(v.Patch, path)...)
			}
		}
	}
	return p
}

func processParametersV2(operation *pb.Operation, path []string) []rules.Field {
	parameters := make([]rules.Field, 0)
	for _, item := range operation.Parameters {
		switch t := item.Oneof.(type) {
		case *pb.ParametersItem_Parameter:
			switch t2 := t.Parameter.Oneof.(type) {
			case *pb.Parameter_BodyParameter:
				parameters = append(parameters, rules.Field{Name: t2.BodyParameter.Name, Path: path})
			case *pb.Parameter_NonBodyParameter:
				switch t3 := t2.NonBodyParameter.Oneof.(type) {
				case *pb.NonBodyParameter_FormDataParameterSubSchema:
					parameters = append(parameters, rules.Field{Name: t3.FormDataParameterSubSchema.Name, Path: path})
				case *pb.NonBodyParameter_HeaderParameterSubSchema:
					parameters = append(parameters, rules.Field{Name: t3.HeaderParameterSubSchema.Name, Path: path})
				case *pb.NonBodyParameter_PathParameterSubSchema:
					parameters = append(parameters, rules.Field{Name: t3.PathParameterSubSchema.Name, Path: path})
				case *pb.NonBodyParameter_QueryParameterSubSchema:
					parameters = append(parameters, rules.Field{Name: t3.QueryParameterSubSchema.Name, Path: path})
				}
			}
		}
	}
	return parameters
}

//AIPLintV2 accepts an OpenAPI v2 document and will call the individual AIP rules
//on the document.
func AIPLintV2(document *pb.Document) (*Linter, int) {
	fields := gatherParametersV2(document)
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
