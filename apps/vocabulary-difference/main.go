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

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/golang/protobuf/proto"

	metrics "github.com/googleapis/gnostic/metrics"
)

/*
These variables were made globally because multiple
functions will be accessing and mutating them.
*/
var schemas map[string]int
var operationID map[string]int
var parameters map[string]int
var properties map[string]int

func openVocabularyFiles() []*metrics.Vocabulary {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanLines)
	var v []*metrics.Vocabulary
	for scanner.Scan() {
		v = append(v, readVocabularyFromFileWithName(scanner.Text()))
	}
	return v
}

func readVocabularyFromFileWithName(filename string) *metrics.Vocabulary {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}

	v := &metrics.Vocabulary{}
	err = proto.Unmarshal(data, v)
	if err != nil {
		panic(err)
	}
	return v
}

func unpackageVocabulary(v *metrics.Vocabulary) {
	for _, s := range v.Schemas {
		schemas[s.Word] += int(s.Count)
	}
	for _, op := range v.Operations {
		operationID[op.Word] += int(op.Count)
	}
	for _, param := range v.Parameters {
		parameters[param.Word] += int(param.Count)
	}
	for _, prop := range v.Properties {
		properties[prop.Word] += int(prop.Count)
	}
}

func mapIntersection(v *metrics.Vocabulary) {
	schemastemp := make(map[string]int)
	operationIDTemp := make(map[string]int)
	parametersTemp := make(map[string]int)
	propertiesTemp := make(map[string]int)
	for _, s := range v.Schemas {
		value, ok := schemas[s.Word]
		if ok {
			schemastemp[s.Word] += (value + int(s.Count))
		}
	}
	for _, op := range v.Operations {
		value, ok := schemas[op.Word]
		if ok {
			operationIDTemp[op.Word] += (value + int(op.Count))
		}
	}
	for _, param := range v.Parameters {
		value, ok := schemas[param.Word]
		if ok {
			parametersTemp[param.Word] += (value + int(param.Count))
		}
	}
	for _, prop := range v.Properties {
		value, ok := schemas[prop.Word]
		if ok {
			propertiesTemp[prop.Word] += (value + int(prop.Count))
		}
	}
	schemas = schemastemp
	operationID = operationIDTemp
	parameters = parametersTemp
	properties = propertiesTemp
}

func main() {
	csvPtr := flag.Bool("csv", false, "generate csv output")
	pbPtr := flag.Bool("pb", false, "generate pb output")

	flag.Parse()

	if !*pbPtr && !*csvPtr {
		flag.PrintDefaults()
		fmt.Printf("Please use one of the above command line arguments.\n")
		os.Exit(-1)
		return
	}
	schemas = make(map[string]int)
	operationID = make(map[string]int)
	parameters = make(map[string]int)
	properties = make(map[string]int)

	vocabSlices := openVocabularyFiles()
	unpackageVocabulary(vocabSlices[0])
	for i := 1; i < len(vocabSlices); i++ {
		mapIntersection(vocabSlices[i])
	}

	v := &metrics.Vocabulary{
		Properties: FillProtoStructures(properties),
		Schemas:    FillProtoStructures(schemas),
		Operations: FillProtoStructures(operationID),
		Parameters: FillProtoStructures(parameters),
	}

	if *pbPtr {
		WritePb(v)
	}
	if *csvPtr {
		WriteCSV(v)
	}
}
