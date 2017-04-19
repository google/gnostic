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
package statistics

import (
	"fmt"
	"log"
	"strings"

	openapi "github.com/googleapis/gnostic/OpenAPIv2"
)

// DocumentStatistics contains information collected about an API description.
type DocumentStatistics struct {
	Name                   string         `json:"name"`
	Title                  string         `json:"title"`
	Operations             map[string]int `json:"operations"`
	DefinitionCount        int            `json:"definitions"`
	ParameterTypes         map[string]int `json:"parameterTypes"`
	ResultTypes            map[string]int `json:"resultTypes"`
	DefinitionFieldTypes   map[string]int `json:"definitionFieldTypes"`
	DefinitionArrayTypes   map[string]int `json:"definitionArrayTypes"`
	HasAnonymousOperations bool           `json:"hasAnonymousOperations"`
	HasAnonymousObjects    bool           `json:"hasAnonymousObjects"`
	AnonymousOperations    []string       `json:"anonymousOperations"`
	AnonymousObjects       []string       `json:"anonymousObjects"`
}

func NewDocumentStatistics() *DocumentStatistics {
	s := &DocumentStatistics{}
	s.Operations = make(map[string]int, 0)
	s.ParameterTypes = make(map[string]int, 0)
	s.ResultTypes = make(map[string]int, 0)
	s.DefinitionFieldTypes = make(map[string]int, 0)
	s.DefinitionArrayTypes = make(map[string]int, 0)
	s.HasAnonymousOperations = false
	s.HasAnonymousObjects = false
	s.AnonymousOperations = make([]string, 0)
	s.AnonymousObjects = make([]string, 0)
	return s
}

func (s *DocumentStatistics) addOperation(name string) {
	s.Operations[name] = s.Operations[name] + 1
}

func (s *DocumentStatistics) addParameterType(path string, name string) {
	if strings.Contains(name, "object") {
		s.HasAnonymousObjects = true
		s.AnonymousObjects = append(s.AnonymousObjects, path)
	}
	s.ParameterTypes[name] = s.ParameterTypes[name] + 1
}

func (s *DocumentStatistics) addResultType(path string, name string) {
	if strings.Contains(name, "object") {
		s.HasAnonymousObjects = true
		s.AnonymousObjects = append(s.AnonymousObjects, path)
	}
	s.ResultTypes[name] = s.ResultTypes[name] + 1
}

func (s *DocumentStatistics) addDefinitionFieldType(path string, name string) {
	if strings.Contains(name, "object") {
		s.HasAnonymousObjects = true
		s.AnonymousObjects = append(s.AnonymousObjects, path)
	}
	s.DefinitionFieldTypes[name] = s.DefinitionFieldTypes[name] + 1
}

func (s *DocumentStatistics) addDefinitionArrayType(path string, name string) {
	if strings.Contains(name, "object") {
		s.HasAnonymousObjects = true
		s.AnonymousObjects = append(s.AnonymousObjects, path)
	}
	s.DefinitionArrayTypes[name] = s.DefinitionArrayTypes[name] + 1
}

func (s *DocumentStatistics) analyzeOperation(path string, operation *openapi.Operation) {
	s.addOperation("total")
	if operation.OperationId == "" {
		s.addOperation("anonymous")
		s.HasAnonymousOperations = true
		s.AnonymousOperations = append(s.AnonymousOperations, path)
	}
	for _, parameter := range operation.Parameters {
		p := parameter.GetParameter()
		if p != nil {
			b := p.GetBodyParameter()
			if b != nil {
				typeName := typeForSchema(b.Schema)
				s.addParameterType(path+"/"+b.Name, typeName)
			}
			n := p.GetNonBodyParameter()
			if n != nil {
				hp := n.GetHeaderParameterSubSchema()
				if hp != nil {
					t := hp.Type
					if t == "array" {
						if hp.Items.Type != "" {
							t += "-of-" + hp.Items.Type
						} else {
							t += "-of-? " + fmt.Sprintf("(%+v)", hp)
						}
					}
					s.addParameterType(path+"/"+hp.Name, t)
				}
				fp := n.GetFormDataParameterSubSchema()
				if fp != nil {
					t := fp.Type
					if t == "array" {
						if fp.Items.Type != "" {
							t += "-of-" + fp.Items.Type
						} else {
							t += "-of-" + fmt.Sprintf("(%+v)", fp)
						}
					}
					s.addParameterType(path+"/"+fp.Name, t)
				}
				qp := n.GetQueryParameterSubSchema()
				if qp != nil {
					t := qp.Type
					if t == "array" {
						if qp.Items.Type != "" {
							t += "-of-" + qp.Items.Type
						} else {
							t += "-of-? " + fmt.Sprintf("(%+v)", qp)
						}
					}
					s.addParameterType(path+"/"+qp.Name, t)
				}
				pp := n.GetPathParameterSubSchema()
				if pp != nil {
					t := pp.Type
					if t == "array" {
						if pp.Items.Type != "" {
							t += "-of-" + pp.Items.Type
						} else {
							t += "-of-? " + fmt.Sprintf("(%+v)", pp)
						}
					}
					s.addParameterType(path+"/"+pp.Name, t)
				}
			}
		}
		r := parameter.GetJsonReference()
		if r != nil {
			s.addParameterType(path+"/", "reference")
		}
	}

	for _, pair := range operation.Responses.ResponseCode {
		value := pair.Value
		response := value.GetResponse()
		if response != nil {
			responseSchema := response.Schema
			responseSchemaSchema := responseSchema.GetSchema()
			if responseSchemaSchema != nil {
				s.addResultType(path+"/responses/"+pair.Name, typeForSchema(responseSchemaSchema))
			}
			responseFileSchema := responseSchema.GetFileSchema()
			if responseFileSchema != nil {
				s.addResultType(path+"/responses/"+pair.Name, typeForFileSchema(responseFileSchema))
			}
		}
		ref := value.GetJsonReference()
		if ref != nil {
		}
	}

}

func (s *DocumentStatistics) analyzeDefinition(path string, definition *openapi.Schema) {
	s.DefinitionCount++
	if definition.Type != nil {
		typeName := definition.Type.Value[0]
		switch typeName {
		case "object":
			if definition.Properties != nil {
				for _, pair := range definition.Properties.AdditionalProperties {
					propertySchema := pair.Value
					propertyType := typeForSchema(propertySchema)
					s.addDefinitionFieldType(path+"/"+pair.Name, propertyType)
				}
			}
		case "array":
			s.addDefinitionArrayType(path+"/", typeForSchema(definition))
		case "string":
			// seems ok
		case "boolean":
			// seems ok
		case "integer":
			// seems ok
		case "null":
			// ...a null definition?
		default:
			log.Printf("type %s", typeName)
		}
	} else {
		// treat unspecified types as "object"
		if definition.Properties != nil {
			for _, pair := range definition.Properties.AdditionalProperties {
				propertySchema := pair.Value
				propertyType := typeForSchema(propertySchema)
				s.addDefinitionFieldType(path+"/"+pair.Name, propertyType)
			}
		}
	}
}

func (s *DocumentStatistics) AnalyzeDocument(document *openapi.Document) {
	s.Title = document.Info.Title
	for _, pair := range document.Paths.Path {
		path := pair.Value
		if path.Get != nil {
			s.addOperation("get")
			s.analyzeOperation("paths"+pair.Name+"/get", path.Get)
		}
		if path.Post != nil {
			s.addOperation("post")
			s.analyzeOperation("paths"+pair.Name+"/post", path.Post)
		}
		if path.Put != nil {
			s.addOperation("put")
			s.analyzeOperation("paths"+pair.Name+"/put", path.Put)
		}
		if path.Delete != nil {
			s.addOperation("delete")
			s.analyzeOperation("paths"+pair.Name+"/delete", path.Delete)
		}
	}
	if document.Definitions != nil {
		for _, pair := range document.Definitions.AdditionalProperties {
			definition := pair.Value
			s.analyzeDefinition("definitions/"+pair.Name, definition)
		}
	}
}

// helpers

// Return a type name to use for a schema.
func typeForSchema(schema *openapi.Schema) string {
	if schema.Type != nil {
		value := schema.Type.Value[0]
		if value == "array" {
			if schema.Items != nil {
				items := schema.Items
				itemSchema := items.Schema[0]
				itemType := typeForSchema(itemSchema)
				return "array-of-" + itemType
			} else if schema.XRef != "" {
				return "array-of-reference"
			} else {
				return fmt.Sprintf("array-of-%+v", schema)
			}
		} else if value == "object" {
			// this might be representable with a map
			// or not
			if (schema.Properties != nil) && (len(schema.Properties.AdditionalProperties) > 0) {
				return value
			}
			if schema.AdditionalProperties.GetSchema().Type != nil {
				return "map-of-" + schema.AdditionalProperties.GetSchema().Type.Value[0]
			}
			if schema.AdditionalProperties.GetSchema().XRef != "" {
				return "map-of-reference"
			} else {
				return "map-of-" + fmt.Sprintf("%+v", schema)
			}
		} else {
			return value
		}
	}
	if schema.XRef != "" {
		return "reference"
	}
	if len(schema.Enum) > 0 {
		return "enum"
	}
	return "object?"
}

func typeForFileSchema(schema *openapi.FileSchema) string {
	if schema.Type != "" {
		value := schema.Type
		switch value {
		case "boolean":
			return "fileschema-" + value
		case "string":
			return "fileschema-" + value
		case "file":
			return "fileschema-" + value
		case "number":
			return "fileschema-" + value
		case "integer":
			return "fileschema-" + value
		case "object":
			return "fileschema-" + value
		case "null":
			return "fileschema-" + value
		}
	}
	return fmt.Sprintf("FILE SCHEMA %+v", schema)
}
