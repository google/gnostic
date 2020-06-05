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
	"sort"

	"github.com/golang/protobuf/proto"

	metrics "github.com/googleapis/gnostic/metrics"
)

/*
These variables were made globally because multiple
functions will be accessing and mutating them.
*/
var schemas map[string]int
var operationId map[string]int
var parameters map[string]int
var properties map[string]int

func writeCSV(v *metrics.Vocabulary) {
	f4, ferror := os.Create("summarize-vocabulary.csv")
	if ferror != nil {
		fmt.Println(ferror)
		f4.Close()
		return
	}
	for _, s := range v.Schemas {
		temp := fmt.Sprintf("%s,%s,%d\n", "schemas", s.Word, int(s.Count))
		f4.WriteString(temp)
	}
	for _, s := range v.Properties {
		temp := fmt.Sprintf("%s,%s,%d\n", "properties", s.Word, int(s.Count))
		f4.WriteString(temp)
	}
	for _, s := range v.Operations {
		temp := fmt.Sprintf("%s,%s,%d\n", "operations", s.Word, int(s.Count))
		f4.WriteString(temp)
	}
	for _, s := range v.Parameters {
		temp := fmt.Sprintf("%s,%s,%d\n", "parameters", s.Word, int(s.Count))
		f4.WriteString(temp)
	}
	f4.Close()
}

func writePb(v *metrics.Vocabulary) {
	bytes, err := proto.Marshal(v)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile("summarize-vocabulary.pb", bytes, 0644)
	if err != nil {
		panic(err)
	}
}

func combineVocabularies() *metrics.Vocabulary {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		readVocabularyFromFileWithName(scanner.Text())
	}

	v := &metrics.Vocabulary{
		Properties: fillProtoStructures(properties),
		Schemas:    fillProtoStructures(schemas),
		Operations: fillProtoStructures(operationId),
		Parameters: fillProtoStructures(parameters),
	}
	return v

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

	v := &metrics.Vocabulary{}
	err = proto.Unmarshal(data, v)
	if err != nil {
		panic(err)
	}
	unpackageVocabulary(v)
}

func unpackageVocabulary(v *metrics.Vocabulary) {
	for _, s := range v.Schemas {
		schemas[s.Word] += int(s.Count)
	}
	for _, op := range v.Operations {
		operationId[op.Word] += int(op.Count)
	}
	for _, param := range v.Parameters {
		parameters[param.Word] += int(param.Count)
	}
	for _, prop := range v.Properties {
		properties[prop.Word] += int(prop.Count)
	}
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
	operationId = make(map[string]int)
	parameters = make(map[string]int)
	properties = make(map[string]int)

	combinedVocab := combineVocabularies()

	if *pbPtr {
		writePb(combinedVocab)
	}
	if *csvPtr {
		writeCSV(combinedVocab)
	}
}
