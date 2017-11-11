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
	"log"

	openapi "github.com/googleapis/gnostic/OpenAPIv2"
)

// DocumentLinter contains information collected about an API description.
type DocumentLinterV2 struct {
	document *openapi.Document `json:"-"`
}

func (d *DocumentLinterV2) Run() {
	d.analyzeDocument(d.document)
}

// NewDocumentLinter builds a new DocumentLinter object.
func NewDocumentLinterV2(document *openapi.Document) *DocumentLinterV2 {
	return &DocumentLinterV2{document: document}
}

// Analyze an OpenAPI description.
// Collect information about types used in the API.
// This should be called exactly once per DocumentLinter object.
func (s *DocumentLinterV2) analyzeDocument(document *openapi.Document) {

	for _, pair := range document.Paths.Path {
		path := pair.Value
		if path.Get != nil {
			s.analyzeOperation("get", "paths"+pair.Name, path.Get)
		}
		if path.Post != nil {
			s.analyzeOperation("post", "paths"+pair.Name, path.Post)
		}
		if path.Put != nil {
			s.analyzeOperation("put", "paths"+pair.Name, path.Put)
		}
		if path.Delete != nil {
			s.analyzeOperation("delete", "paths"+pair.Name, path.Delete)
		}
	}
	if document.Definitions != nil {
		for _, pair := range document.Definitions.AdditionalProperties {
			definition := pair.Value
			s.analyzeDefinition("definitions/"+pair.Name, definition)
		}
	}
}

func (s *DocumentLinterV2) analyzeOperation(method string, path string, operation *openapi.Operation) {

	fullname := method + " " + path

	if operation.Description == "" {
		log.Printf("%s has no description", fullname)
	}
	for _, parameter := range operation.Parameters {
		p := parameter.GetParameter()
		if p != nil {
			b := p.GetBodyParameter()
			if b != nil && b.Description == "" {
				log.Printf("%s %s parameter has no description", fullname, b.Name)
			}
			n := p.GetNonBodyParameter()
			if n != nil {
				hp := n.GetHeaderParameterSubSchema()
				if hp != nil && hp.Description == "" {
					log.Printf("%s %s parameter has no description", fullname, hp.Name)
				}
				fp := n.GetFormDataParameterSubSchema()
				if fp != nil && fp.Description == "" {
					log.Printf("%s %s parameter has no description", fullname, fp.Name)
				}
				qp := n.GetQueryParameterSubSchema()
				if qp != nil && qp.Description == "" {
					log.Printf("%s %s parameter has no description", fullname, qp.Name)
				}
				pp := n.GetPathParameterSubSchema()
				if pp != nil && pp.Description == "" {
					log.Printf("%s %s parameter has no description", fullname, pp.Name)
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
				log.Printf("%s %s response has no description", fullname, pair.Name)
			}
			responseFileSchema := responseSchema.GetFileSchema()
			if responseFileSchema != nil && responseFileSchema.Description == "" {
				log.Printf("%s %s response has no description", fullname, pair.Name)
			}
		}
	}
}

// Analyze a definition in an OpenAPI description.
// Collect information about the definition type and any subsidiary types,
// such as the types of object fields or array elements.
func (s *DocumentLinterV2) analyzeDefinition(path string, definition *openapi.Schema) {
	if definition.Description == "" {
		log.Printf("%s has no description", path)
	}

	if definition.Properties != nil {
		for _, pair := range definition.Properties.AdditionalProperties {
			propertySchema := pair.Value
			if propertySchema.Description == "" {
				log.Printf("%s %s property has no description", path, pair.Name)
			}
		}
	}
}
