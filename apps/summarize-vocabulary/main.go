// Copyright 2018 Google Inc. All Rights Reserved.
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

// Filter and display messages produced by gnostic invocations.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"

	"github.com/golang/protobuf/proto"

	metrics "github.com/googleapis/gnostic/metrics"
)

var schemas map[string]int
var operationId map[string]int
var parameters map[string]int
var properties map[string]int

func pbToCSV(vocab *metrics.Vocabulary) {
	f4, ferror := os.Create("summarize-vocabulary.csv")
	if ferror != nil {
		fmt.Println(ferror)
		f4.Close()
		return
	}
	for _, s := range vocab.Schemas {
		temp := fmt.Sprintf("%s,%s,%d\n", "Schemas", s.Word, int(s.Count))
		f4.WriteString(temp)
	}
	for _, s := range vocab.Properties {
		temp := fmt.Sprintf("%s,%s,%d\n", "Properties", s.Word, int(s.Count))
		f4.WriteString(temp)
	}
	for _, s := range vocab.Operations {
		temp := fmt.Sprintf("%s,%s,%d\n", "Operations", s.Word, int(s.Count))
		f4.WriteString(temp)
	}
	for _, s := range vocab.Parameters {
		temp := fmt.Sprintf("%s,%s,%d\n", "Parameters", s.Word, int(s.Count))
		f4.WriteString(temp)
	}
	f4.Close()
}

func pbOutput(combinedVocab *metrics.Vocabulary) {
	bytes, err := proto.Marshal(combinedVocab)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile("summarize-vocabulary.pb", bytes, 0644)
	if err != nil {
		panic(err)
	}
}

func openCombinedPbResults(filename string) *metrics.Vocabulary {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		readVocabularyFromFileWithName(scanner.Text())
	}
	file.Close()
	vocab := &metrics.Vocabulary{
		Properties: fillProtoStructures(properties),
		Schemas:    fillProtoStructures(schemas),
		Operations: fillProtoStructures(operationId),
		Parameters: fillProtoStructures(parameters),
	}
	return vocab

}

func fillProtoStructures(m map[string]int) []*metrics.WordCount {
	key_names := make([]string, 0, len(m))
	for key := range m {
		key_names = append(key_names, key)
	}
	sort.Strings(key_names)

	counts := make([]*metrics.WordCount, 0)
	for _, k := range key_names {
		temp := &metrics.WordCount{
			Word:  k,
			Count: int32(m[k]),
		}
		counts = append(counts, temp)
	}
	return counts
}

func readVocabularyFromFileWithName(filename string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}

	vocab := &metrics.Vocabulary{}
	err = proto.Unmarshal(data, vocab)
	if err != nil {
		panic(err)
	}
	unpackageVocabulary(vocab)
}

func unpackageVocabulary(vocab *metrics.Vocabulary) {
	for _, s := range vocab.Schemas {
		schemas[s.Word] += int(s.Count)
	}
	for _, op := range vocab.Operations {
		operationId[op.Word] += int(op.Count)
	}
	for _, param := range vocab.Parameters {
		parameters[param.Word] += int(param.Count)
	}
	for _, prop := range vocab.Properties {
		properties[prop.Word] += int(prop.Count)
	}
}

func main() {
	// outputPtr := flag.String("output", "NIL", "output type")
	flag.Parse()
	args := flag.Args()

	schemas = make(map[string]int)
	operationId = make(map[string]int)
	parameters = make(map[string]int)
	properties = make(map[string]int)

	combinedVocab := openCombinedPbResults(args[0])

	if len(args) == 1 {
		return
	}
	switch args[1] {
	case "--csv":
		pbToCSV(combinedVocab)
		return
	case "--pb":
		pbOutput(combinedVocab)
		return
	default:
		return
	}
}
