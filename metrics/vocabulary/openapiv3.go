package vocabulary

import (
	"sort"

	metrics "github.com/googleapis/gnostic/metrics"
	openapi_v3 "github.com/googleapis/gnostic/openapiv3"
)

func fillProtoStructures(m map[string]int) []*metrics.WordCount {
	keyNames := make([]string, 0, len(m))
	for key := range m {
		keyNames = append(keyNames, key)
	}
	sort.Strings(keyNames)

	counts := make([]*metrics.WordCount, 0)
	for _, k := range keyNames {
		temp := &metrics.WordCount{
			Word:  k,
			Count: int32(m[k]),
		}
		counts = append(counts, temp)
	}
	return counts
}

func processOperationV3(operation *openapi_v3.Operation, operationID, parameters map[string]int) {
	if operation.OperationId != "" {
		operationID[operation.OperationId]++
	}
	for _, item := range operation.Parameters {
		switch t := item.Oneof.(type) {
		case *openapi_v3.ParameterOrReference_Parameter:
			parameters[t.Parameter.Name]++
		}
	}
}

func processComponentsV3(components *openapi_v3.Components, schemas, parameters, properties map[string]int) {
	processParametersV3(components, schemas, parameters)
	processSchemasV3(components, schemas, properties)
	processResponsesV3(components, schemas)
}

func processParametersV3(components *openapi_v3.Components, schemas, parameters map[string]int) {
	if components.Parameters == nil {
		return
	}
	for _, pair := range components.Parameters.AdditionalProperties {
		switch t := pair.Value.Oneof.(type) {
		case *openapi_v3.ParameterOrReference_Parameter:
			parameters[t.Parameter.Name]++
		}
	}
}

func processSchemasV3(components *openapi_v3.Components, schemas, properties map[string]int) {
	if components.Schemas == nil {
		return
	}
	for _, pair := range components.Schemas.AdditionalProperties {
		schemas[pair.Name]++
		processSchemaV3(pair.Value, properties)
	}
}

func processSchemaV3(schema *openapi_v3.SchemaOrReference, properties map[string]int) {
	if schema == nil {
		return
	}
	switch t := schema.Oneof.(type) {
	case *openapi_v3.SchemaOrReference_Reference:
		return
	case *openapi_v3.SchemaOrReference_Schema:
		if t.Schema.Properties != nil {
			for _, pair := range t.Schema.Properties.AdditionalProperties {
				properties[pair.Name]++
			}
		}
	}
}

func processResponsesV3(components *openapi_v3.Components, schemas map[string]int) {
	if components.Responses == nil {
		return
	}
	for _, pair := range components.Responses.AdditionalProperties {
		schemas[pair.Name]++
	}
}

func NewVocabularyFromOpenAPIv3(document *openapi_v3.Document) *metrics.Vocabulary {
	schemas := make(map[string]int)
	operationID := make(map[string]int)
	parameters := make(map[string]int)
	properties := make(map[string]int)

	if document.Components != nil {
		processComponentsV3(document.Components, schemas, parameters, properties)

	}
	for _, pair := range document.Paths.Path {
		v := pair.Value
		if v.Get != nil {
			processOperationV3(v.Get, operationID, parameters)
		}
		if v.Post != nil {
			processOperationV3(v.Post, operationID, parameters)
		}
		if v.Put != nil {
			processOperationV3(v.Put, operationID, parameters)
		}
		if v.Patch != nil {
			processOperationV3(v.Patch, operationID, parameters)
		}
		if v.Delete != nil {
			processOperationV3(v.Delete, operationID, parameters)
		}
	}

	vocab := &metrics.Vocabulary{
		Properties: fillProtoStructures(properties),
		Schemas:    fillProtoStructures(schemas),
		Operations: fillProtoStructures(operationID),
		Parameters: fillProtoStructures(parameters),
	}
	return vocab
}
