package main

import (
	metrics "github.com/googleapis/gnostic/metrics"
)

func mapDifference(v *metrics.Vocabulary) {
	for _, s := range v.Schemas {
		_, ok := schemas[s.Word]
		if ok {
			delete(schemas, s.Word)
		}
	}
	for _, op := range v.Operations {
		_, ok := operationID[op.Word]
		if ok {
			delete(operationID, op.Word)
		}
	}
	for _, param := range v.Parameters {
		_, ok := parameters[param.Word]
		if ok {
			delete(parameters, param.Word)
		}
	}
	for _, prop := range v.Properties {
		_, ok := properties[prop.Word]
		if ok {
			delete(properties, prop.Word)
		}
	}
}

func vocabularyDifference(vocab []*metrics.Vocabulary) *metrics.Vocabulary {
	schemas = make(map[string]int)
	operationID = make(map[string]int)
	parameters = make(map[string]int)
	properties = make(map[string]int)

	unpackageVocabulary(vocab[0])
	for i := 1; i < len(vocab); i++ {
		mapDifference(vocab[i])
	}

	v := &metrics.Vocabulary{
		Properties: fillProtoStructure(properties),
		Schemas:    fillProtoStructure(schemas),
		Operations: fillProtoStructure(operationID),
		Parameters: fillProtoStructure(parameters),
	}
	return v
}
