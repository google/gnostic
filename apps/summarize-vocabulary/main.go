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

func combineVocabularies() *metrics.Vocabulary {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		readVocabularyFromFileWithName(scanner.Text())
	}

	v := &metrics.Vocabulary{
		Properties: FillProtoStructures(properties),
		Schemas:    FillProtoStructures(schemas),
		Operations: FillProtoStructures(operationID),
		Parameters: FillProtoStructures(parameters),
	}
	return v

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
	UnpackageVocabulary(v)
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

	combinedVocab := combineVocabularies()

	if *pbPtr {
		WritePb(combinedVocab)
	}
	if *csvPtr {
		WriteCSV(combinedVocab)
	}
}
