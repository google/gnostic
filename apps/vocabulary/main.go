// Copyright 2017 Google Inc. All Rights Reserved.
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

// report is a demo application that displays information about an
// OpenAPI description.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/golang/protobuf/proto"

	pb "github.com/googleapis/gnostic/openapiv2"
)

func readDocumentFromFileWithName(filename string) (*pb.Document, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	document := &pb.Document{}
	err = proto.Unmarshal(data, document)
	if err != nil {
		return nil, err
	}
	return document, nil
}
func addToMap(word string, mapName map[string]int) {
	_, ok := mapName[word]
	if ok {
		mapName[word] += 1
	} else {
		mapName[word] = 1
	}
}

func createCSV(names map[string]int) {
	f, ferror := os.Create("IndexFreqSort.csv")
	if ferror != nil {
		fmt.Println(ferror)
		f.Close()
		return
	}
	//Sort by value
	type kv struct {
		Key   string
		Value int
	}

	var ss []kv
	for k, v := range names {
		ss = append(ss, kv{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})
	for _, kv := range ss {
		temp := fmt.Sprintf("%s,%d\n", kv.Key, kv.Value)
		f.WriteString(temp)

	}
	f.Close()

	//Sort alphabetically
	key_names := make([]string, 0, len(names))
	key_names_lower := make([]string, 0, len(names))
	for key := range names {
		key_names = append(key_names, key)
		key_names_lower = append(key_names_lower, strings.ToLower(key))
	}
	sort.Strings(key_names)
	sort.Strings(key_names_lower)
	f2, ferror := os.Create("IndexAlphaSort.csv")
	if ferror != nil {
		fmt.Println(ferror)
		f2.Close()
		return
	}
	for _, key := range key_names {
		temp := fmt.Sprintf("%s,%d\n", key, names[key])
		f2.WriteString(temp)

	}
	f2.Close()

	f3, ferror := os.Create("IndexAlphaLowerSort.csv")
	if ferror != nil {
		fmt.Println(ferror)
		f3.Close()
		return
	}
	for _, key := range key_names_lower {
		temp := fmt.Sprintf("%s,%d\n", key, names[key])
		f3.WriteString(temp)

	}
	f3.Close()

	f4, ferror := os.Create("IndexRegular.csv")
	if ferror != nil {
		fmt.Println(ferror)
		f4.Close()
		return
	}
	for k, _ := range names {
		temp := fmt.Sprintf("%s\n", k)
		f4.WriteString(temp)
	}
	f4.Close()
}

func printDocument(document *pb.Document, schemas map[string]int, operationId map[string]int, names map[string]int, properties map[string]int) {
	//Start
	if document.Definitions != nil && document.Definitions.AdditionalProperties != nil {
		for _, pair := range document.Definitions.AdditionalProperties {
			addToMap(pair.Name, schemas)
			printSchema(pair.Value, properties)
		}
	}
	for _, pair := range document.Paths.Path {
		v := pair.Value
		if v.Get != nil {
			printOperation(v.Get, operationId, names)
		}
		if v.Post != nil {
			printOperation(v.Post, operationId, names)
		}
		if v.Put != nil {
			printOperation(v.Put, operationId, names)
		}
		if v.Patch != nil {
			printOperation(v.Patch, operationId, names)
		}
		if v.Delete != nil {
			printOperation(v.Delete, operationId, names)
		}
	}
}

//^^^ Get rid of print post/get/indent

func printOperation(operation *pb.Operation, operationId map[string]int, names map[string]int) {
	if operation.OperationId != "" {
		addToMap(operation.OperationId, operationId)
	}
	for _, item := range operation.Parameters {
		switch t := item.Oneof.(type) {
		case *pb.ParametersItem_Parameter:
			switch t2 := t.Parameter.Oneof.(type) {
			case *pb.Parameter_BodyParameter:
				addToMap(t2.BodyParameter.Name, names)
			case *pb.Parameter_NonBodyParameter:
				switch t3 := t2.NonBodyParameter.Oneof.(type) {
				case *pb.NonBodyParameter_FormDataParameterSubSchema:
					addToMap(t3.FormDataParameterSubSchema.Name, names)
				case *pb.NonBodyParameter_HeaderParameterSubSchema:
					addToMap(t3.HeaderParameterSubSchema.Name, names)
				case *pb.NonBodyParameter_PathParameterSubSchema:
					addToMap(t3.PathParameterSubSchema.Name, names)
				case *pb.NonBodyParameter_QueryParameterSubSchema:
					addToMap(t3.QueryParameterSubSchema.Name, names)
				}
			}
		}
	}
}

func printSchema(schema *pb.Schema, properties map[string]int) {
	if schema.Properties != nil {
		for _, pair := range schema.Properties.AdditionalProperties {
			addToMap(pair.Name, properties)
		}
	}
}

func main() {
	flag.Parse()
	args := flag.Args()

	document, err := readDocumentFromFileWithName(args[0])

	if err != nil {
		log.Printf("Error reading %s. This sample expects OpenAPI v2.", args[0])
		os.Exit(-1)
	}

	var schemas map[string]int
	schemas = make(map[string]int)

	var operationId map[string]int
	operationId = make(map[string]int)

	var names map[string]int
	names = make(map[string]int)

	var properties map[string]int
	properties = make(map[string]int)

	printDocument(document, schemas, operationId, names, properties)

	vocab := &Vocabulary{}
	for k, v := range schemas {
		temp := &WordCount{
			Word:  k,
			Count: int32(v),
		}
		vocab.Schemas = append(vocab.Schemas, temp)
	}

	for k, v := range operationId {
		temp := &WordCount{
			Word:  k,
			Count: int32(v),
		}
		vocab.Operations = append(vocab.Operations, temp)
	}

	for k, v := range names {
		temp := &WordCount{
			Word:  k,
			Count: int32(v),
		}
		vocab.Paramaters = append(vocab.Paramaters, temp)
	}

	for k, v := range properties {
		temp := &WordCount{
			Word:  k,
			Count: int32(v),
		}
		vocab.Properties = append(vocab.Properties, temp)
	}

	bytes, err := proto.Marshal(vocab)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile("vocabulary_results.pb", bytes, 0644)
	if err != nil {
		panic(err)
	}

}
