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
	"io/ioutil"
	"os"

	"github.com/golang/protobuf/jsonpb"
	metrics "github.com/googleapis/gnostic/metrics"
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

func serializeProto(vocab *metrics.Vocabulary) {
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

func main() {
	// flag.Parse()
	// args := flag.Args()

	// var schemas map[string]int
	// schemas = make(map[string]int)

	// var operationId map[string]int
	// operationId = make(map[string]int)

	// var names map[string]int
	// names = make(map[string]int)

	// var properties map[string]int
	// properties = make(map[string]int)

	// //Temporary, for now using filename to check file type
	// version_flag := strings.Contains(args[0], "swagger")
	// switch version_flag {
	// case true:
	// 	document, err := readDocumentFromFileWithNameV2(args[0])
	// 	if err != nil {
	// 		log.Printf("Error reading %s.", args[0])
	// 		os.Exit(1)
	// 	}
	// 	vocab := processDocumentV2(document, schemas, operationId, names, properties)
	// 	serializeProto(vocab)
	// default:
	// 	document, err := readDocumentFromFileWithNameV3(args[0])
	// 	if err != nil {
	// 		log.Printf("Error reading %s.", args[0])
	// 		os.Exit(1)
	// 	}
	// 	vocab := processDocumentV3(document, schemas, operationId, names, properties)
	// 	serializeProto(vocab)

	// }

}
