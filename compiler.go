// Copyright 2016 Google Inc. All Rights Reserved.
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

//go:generate ./COMPILE-PROTOS.sh

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"

	pb "openapi"
)

func ReadDocumentFromFile(filename string) *pb.Document {
	examplesDir := os.Getenv("GOPATH") + "/src/github.com/googleapis/openapi-compiler/examples"
	file, e := ioutil.ReadFile(examplesDir + "/" + filename)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}
	var raw interface{}
	json.Unmarshal(file, &raw)

	//fmt.Printf("%+v\n", raw)

	document := buildDocumentForMap(raw)
	return document
}

func main() {
	fmt.Printf("Version: %s\n", version())
	document := ReadDocumentFromFile("petstore.json")
	fmt.Printf("doc: %+v\n", document)
}

// helper function for compiler
func unpackMap(in interface{}) (map[string]interface{}, []string, bool) {
	m, ok := in.(map[string]interface{})
	if !ok {
		return nil, nil, ok
	}
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return m, keys, ok
}

func mapHasKey(m map[string]interface{}, key string) bool {
	_, ok := m[key]
	return ok
}

func convertInterfaceArrayToStringArray(interfaceArray []interface{}) []string {
	stringArray := make([]string, 0)
	for _, item := range interfaceArray {
		v, ok := item.(string)
		if ok {
			stringArray = append(stringArray, v)
		}
	}
	return stringArray
}
