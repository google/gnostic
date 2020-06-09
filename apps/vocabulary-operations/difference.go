package main

import (
	"fmt"

	metrics "github.com/googleapis/gnostic/metrics"
)

/*
These variables were made globally because multiple
functions will be accessing and mutating them.
*/

func mapDifference(v *metrics.Vocabulary) {
	for _, s := range v.Schemas {
		_, ok := schemas[s.Word]
		if ok {
			fmt.Printf("HEREEE")
			delete(schemas, s.Word)
		}
	}
	for _, op := range v.Operations {
		_, ok := operationID[op.Word]
		if ok {
			fmt.Printf("HEREEE")
			delete(operationID, op.Word)
		}
	}
	for _, param := range v.Parameters {
		_, ok := parameters[param.Word]
		if ok {
			fmt.Printf("HEREEE")
			delete(parameters, param.Word)
		}
	}
	for _, prop := range v.Properties {
		_, ok := properties[prop.Word]
		if ok {
			fmt.Printf("HEREEE")
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
	fmt.Println(len(schemas))
	fmt.Println(len(operationID))
	fmt.Println(len(parameters))
	fmt.Println(len(properties))
	for i := 1; i < len(vocab); i++ {
		mapDifference(vocab[i])
	}

	v := &metrics.Vocabulary{
		Properties: FillProtoStructures(properties),
		Schemas:    FillProtoStructures(schemas),
		Operations: FillProtoStructures(operationID),
		Parameters: FillProtoStructures(parameters),
	}
	return v
}
