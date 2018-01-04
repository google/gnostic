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
	openapi "github.com/googleapis/gnostic/OpenAPIv2"
	plugins "github.com/googleapis/gnostic/plugins"
)

// DocumentLinter contains information collected about an API description.
type DocumentLinterV2 struct {
	document *openapi.Document `json:"-"`
}

func (d *DocumentLinterV2) Run() []*plugins.Message {
	return d.analyzeDocument(d.document)
}

// NewDocumentLinter builds a new DocumentLinter object.
func NewDocumentLinterV2(document *openapi.Document) *DocumentLinterV2 {
	return &DocumentLinterV2{document: document}
}

// Analyze an OpenAPI description.
// Collect information about types used in the API.
// This should be called exactly once per DocumentLinter object.
func (s *DocumentLinterV2) analyzeDocument(document *openapi.Document) []*plugins.Message {
	messages := make([]*plugins.Message, 0, 0)

	for _, pair := range document.Paths.Path {
		path := pair.Value
		if path.Get != nil {
			messages = append(messages, s.analyzeOperation("get", "paths"+pair.Name, path.Get)...)
		}
		if path.Post != nil {
			messages = append(messages, s.analyzeOperation("post", "paths"+pair.Name, path.Post)...)
		}
		if path.Put != nil {
			messages = append(messages, s.analyzeOperation("put", "paths"+pair.Name, path.Put)...)
		}
		if path.Delete != nil {
			messages = append(messages, s.analyzeOperation("delete", "paths"+pair.Name, path.Delete)...)
		}
	}
	if document.Definitions != nil {
		for _, pair := range document.Definitions.AdditionalProperties {
			definition := pair.Value
			messages = append(messages, s.analyzeDefinition("definitions/"+pair.Name, definition)...)
		}
	}
	return messages
}

func (s *DocumentLinterV2) analyzeOperation(method string, path string, operation *openapi.Operation) []*plugins.Message {
	messages := make([]*plugins.Message, 0)

	fullname := method + " " + path

	if operation.Description == "" {
		messages = append(messages,
			&plugins.Message{
				Level: plugins.Message_WARNING,
				Code:  "NODESCRIPTION",
				Text:  "Operation has no description.",
				Path:  fullname})
	}
	for _, parameter := range operation.Parameters {
		p := parameter.GetParameter()
		if p != nil {
			b := p.GetBodyParameter()
			if b != nil && b.Description == "" {
				messages = append(messages,
					&plugins.Message{
						Level: plugins.Message_WARNING,
						Code:  "NODESCRIPTION",
						Text:  "Parameter has no description.",
						Path:  fullname + "/" + b.Name})
			}
			n := p.GetNonBodyParameter()
			if n != nil {
				hp := n.GetHeaderParameterSubSchema()
				if hp != nil && hp.Description == "" {
					messages = append(messages,
						&plugins.Message{
							Level: plugins.Message_WARNING,
							Code:  "NODESCRIPTION",
							Text:  "Parameter has no description.",
							Path:  fullname + "/" + hp.Name})
				}
				fp := n.GetFormDataParameterSubSchema()
				if fp != nil && fp.Description == "" {
					messages = append(messages,
						&plugins.Message{
							Level: plugins.Message_WARNING,
							Code:  "NODESCRIPTION",
							Text:  "Parameter has no description.",
							Path:  fullname + "/" + fp.Name})
				}
				qp := n.GetQueryParameterSubSchema()
				if qp != nil && qp.Description == "" {
					messages = append(messages,
						&plugins.Message{
							Level: plugins.Message_WARNING,
							Code:  "NODESCRIPTION",
							Text:  "Parameter has no description.",
							Path:  fullname + "/" + qp.Name})
				}
				pp := n.GetPathParameterSubSchema()
				if pp != nil && pp.Description == "" {
					messages = append(messages,
						&plugins.Message{
							Level: plugins.Message_WARNING,
							Code:  "NODESCRIPTION",
							Text:  "Parameter has no description.",
							Path:  fullname + "/" + pp.Name})
				}
			}
		}
	}
	for _, pair := range operation.Responses.ResponseCode {
		value := pair.Value
		response := value.GetResponse()
		if response != nil {
			responseSchema := response.Schema
			responseSchemaSchema := responseSchema.GetSchema()
			if responseSchemaSchema != nil && responseSchemaSchema.Description == "" {
				messages = append(messages,
					&plugins.Message{
						Level: plugins.Message_WARNING,
						Code:  "NODESCRIPTION",
						Text:  "Response has no description.",
						Path:  fullname + "/" + pair.Name})
			}
			responseFileSchema := responseSchema.GetFileSchema()
			if responseFileSchema != nil && responseFileSchema.Description == "" {
				messages = append(messages,
					&plugins.Message{
						Level: plugins.Message_WARNING,
						Code:  "NODESCRIPTION",
						Text:  "Response has no description.",
						Path:  fullname + "/" + pair.Name})
			}
		}
	}
	return messages
}

// Analyze a definition in an OpenAPI description.
// Collect information about the definition type and any subsidiary types,
// such as the types of object fields or array elements.
func (s *DocumentLinterV2) analyzeDefinition(path string, definition *openapi.Schema) []*plugins.Message {
	messages := make([]*plugins.Message, 0)
	if definition.Description == "" {
		messages = append(messages,
			&plugins.Message{
				Level: plugins.Message_WARNING,
				Code:  "NODESCRIPTION",
				Text:  "Definition has no description.",
				Path:  path})
	}

	if definition.Properties != nil {
		for _, pair := range definition.Properties.AdditionalProperties {
			propertySchema := pair.Value
			if propertySchema.Description == "" {
				messages = append(messages,
					&plugins.Message{
						Level: plugins.Message_WARNING,
						Code:  "NODESCRIPTION",
						Text:  "Property has no description.",
						Path:  path + "/" + pair.Name})
			}
		}
	}
	return messages
}
