package main

import (
	metrics "github.com/googleapis/gnostic/metrics"
)

func vocabularyUnion(vocab []*metrics.Vocabulary) *metrics.Vocabulary {
	schemas = make(map[string]int)
	operationID = make(map[string]int)
	parameters = make(map[string]int)
	properties = make(map[string]int)

	for _, v := range vocab {
		unpackageVocabulary(v)
	}

	combinedVocab := &metrics.Vocabulary{
		Properties: fillProtoStructure(properties),
		Schemas:    fillProtoStructure(schemas),
		Operations: fillProtoStructure(operationID),
		Parameters: fillProtoStructure(parameters),
	}

	return combinedVocab
}
