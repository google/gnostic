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

// Package gnostic_vocabulary provides operation for Vocabulary structs
package vocabulary

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"sort"

	metrics "github.com/googleapis/gnostic/metrics"
	"google.golang.org/protobuf/proto"
)

/*
These variables were made globally because multiple
functions will be accessing and mutating them.
*/
var schemas map[string]int
var operationID map[string]int
var parameters map[string]int
var properties map[string]int

// WriteCSV converts a Vocabulary pb file to a user-friendly readable CSV file.
// The format of the CSV file is as follows: "group","word","frequency"
func WriteCSV(v *metrics.Vocabulary, filename string) error {
	if filename == "" {
		filename = "vocabulary-operation.csv"
	}
	f4, ferror := os.Create(filename)
	defer f4.Close()
	if ferror != nil {
		return ferror
	}
	for _, s := range v.Schemas {
		temp := fmt.Sprintf("%s,\"%s\",%d\n", "schemas", s.Word, int(s.Count))
		f4.WriteString(temp)
	}
	for _, s := range v.Properties {
		temp := fmt.Sprintf("%s,\"%s\",%d\n", "properties", s.Word, int(s.Count))
		f4.WriteString(temp)
	}
	for _, s := range v.Operations {
		temp := fmt.Sprintf("%s,\"%s\",%d\n", "operations", s.Word, int(s.Count))
		f4.WriteString(temp)
	}
	for _, s := range v.Parameters {
		temp := fmt.Sprintf("%s,\"%s\",%d\n", "parameters", s.Word, int(s.Count))
		f4.WriteString(temp)
	}
	return nil
}

// WritePb create a protocol buffer file that contains the wire-format encoding of a Vocabulary struct.
func WritePb(v *metrics.Vocabulary) error {
	bytes, err := proto.Marshal(v)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("vocabulary-operation.pb", bytes, 0644)
	if err != nil {
		return err
	}
	return nil
}

// fillProtoStructure adds data to the Word Count structure.
// The Word Count structure can then be added to the Vocabulary protocol buffer.
func fillProtoStructure(m map[string]int) []*metrics.WordCount {
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

// unpackageVocabulary unravels the Vocabulary struct by converting their
// fields to maps in order to perform operations on the data.
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

// combineVocabularies scans for Vocabulary structures using standard input.
// The structures are then combined into one large Vocabulary.
// This function utilizes the readVocabularyFromFileWithName() function to
// open the Vocabulary protocol buffers.
func combineVocabularies() *metrics.Vocabulary {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		readVocabularyFromFileWithName(scanner.Text())
	}

	v := &metrics.Vocabulary{
		Properties: fillProtoStructure(properties),
		Schemas:    fillProtoStructure(schemas),
		Operations: fillProtoStructure(operationID),
		Parameters: fillProtoStructure(parameters),
	}
	return v

}

// readVocabularyFromFileWithNametakes the filename of a Vocabulary protocol
// buffer file and parses the wire-format message into a Vocabulary struct.
func readVocabularyFromFileWithName(filename string) {
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
	unpackageVocabulary(v)
}

func isEmpty(v *metrics.Vocabulary) bool {
	if len(v.Schemas) == 0 && len(v.Properties) == 0 && len(v.Operations) == 0 && len(v.Parameters) == 0 {
		return true
	}
	return false
}
