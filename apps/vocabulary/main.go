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
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	metrics "github.com/googleapis/gnostic/metrics"
	openapi_v2 "github.com/googleapis/gnostic/openapiv2"
	openapi_v3 "github.com/googleapis/gnostic/openapiv3"
	"google.golang.org/protobuf/proto"
)

func fillProtoStructures(m map[string]int) []*metrics.WordCount {
	counts := make([]*metrics.WordCount, 0)
	for k, v := range m {
		temp := &metrics.WordCount{
			Word:  k,
			Count: int32(v),
		}
		counts = append(counts, temp)
	}
	return counts
}
func runV2(document *openapi_v2.Document) {

}
func runV3(document *openapi_v3.Document) {

}

func main() {
	flag.Parse()
	args := flag.Args()

	var schemas map[string]int
	schemas = make(map[string]int)

	var operationId map[string]int
	operationId = make(map[string]int)

	var names map[string]int
	names = make(map[string]int)

	var properties map[string]int
	properties = make(map[string]int)

	version_flag := strings.Contains(args[0], "swagger")
	switch version_flag {
	case true:
		document, err := readDocumentFromFileWithName(args[0])
		if err != nil {
			log.Printf("Error reading %s.", args[0])
			os.Exit(1)
		}
		processDocument(document, schemas, operationId, names, properties)
	default:
		document, err := readDocumentFromFileWithNameV3(args[0])
		if err != nil {
			log.Printf("Error reading %s.", args[0])
			os.Exit(1)
		}
		processDocumentV3(document, schemas, operationId, names, properties)

	}

	vocab := &metrics.Vocabulary{
		Schemas:    fillProtoStructures(schemas),
		Operations: fillProtoStructures(operationId),
		Paramaters: fillProtoStructures(names),
		Properties: fillProtoStructures(properties),
	}

	bytes, err := proto.Marshal(vocab)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile("vocabulary.pb", bytes, 0644)
	if err != nil {
		panic(err)
	}

	m := jsonpb.Marshaler{Indent: " "}
	s, err := m.MarshalToString(vocab)
	jsonOutput := []byte(s)
	err = ioutil.WriteFile("vocabulary.json", jsonOutput, os.ModePerm)
}
