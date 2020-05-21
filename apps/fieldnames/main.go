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
func addToMap(word string, names map[string]int) {
	_, ok := names[word]
	if ok {
		names[word] += 1
	} else {
		names[word] = 1
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

func printDocument(document *pb.Document, names map[string]int) {
	//Start
	if document.Definitions != nil && document.Definitions.AdditionalProperties != nil {
		for _, pair := range document.Definitions.AdditionalProperties {
			printSchema(pair.Value, names)
		}
	}
	for _, pair := range document.Paths.Path {
		v := pair.Value
		if v.Get != nil {
			printOperation(v.Get, names)
		}
		if v.Post != nil {
			printOperation(v.Post, names)
		}
	}
}

//^^^ Get rid of print post/get/indent

func printOperation(operation *pb.Operation, names map[string]int) {
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

func printSchema(schema *pb.Schema, names map[string]int) {
	if schema.Properties != nil {
		for _, pair := range schema.Properties.AdditionalProperties {
			addToMap(pair.Name, names)
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

	var names map[string]int
	names = make(map[string]int)

	printDocument(document, names)

	createCSV(names)

}
