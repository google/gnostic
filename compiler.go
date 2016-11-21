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
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"os"
	"sort"

	pb "openapi"
)

func DescribeMap(in interface{}, indent string) string {
	description := ""
	m, ok := in.(map[string]interface{})
	if ok {
		keys := make([]string, 0)
		for k, _ := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := m[k]
			description += fmt.Sprintf("%s%s:\n", indent, k)
			description += DescribeMap(v, indent+"  ")
		}
		return description
	}
	a, ok := in.([]interface{})
	if ok {
		for i, v := range a {
			description += fmt.Sprintf("%s%d:\n", indent, i)
			description += DescribeMap(v, indent+"  ")
		}
		return description
	}
	description += fmt.Sprintf("%s%+v\n", indent, in)
	return description
}

func ReadDocumentFromFile(filename string) *pb.Document {
	examplesDir := os.Getenv("GOPATH") + "/src/github.com/googleapis/openapi-compiler/examples"
	file, e := ioutil.ReadFile(examplesDir + "/" + filename)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}
	var raw interface{}
	json.Unmarshal(file, &raw)

	rawDescription := DescribeMap(raw, "")
	ioutil.WriteFile("petstore.txt", []byte(rawDescription), 0644)

	document := buildDocument(raw)
	return document
}

func main() {
	fmt.Printf("Version: %s\n", version())
	document := ReadDocumentFromFile("petstore.json")
	ioutil.WriteFile("petstore.pb", []byte(proto.MarshalTextString(document)), 0644)
}
