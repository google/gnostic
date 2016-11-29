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
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/googleapis/openapi-compiler/OpenAPIv2"
)

func describeMap(in interface{}, indent string) string {
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
			description += describeMap(v, indent+"  ")
		}
		return description
	}
	a, ok := in.([]interface{})
	if ok {
		for i, v := range a {
			description += fmt.Sprintf("%s%d:\n", indent, i)
			description += describeMap(v, indent+"  ")
		}
		return description
	}
	description += fmt.Sprintf("%s%+v\n", indent, in)
	return description
}

func readFile(filename string) interface{} {
	file, e := ioutil.ReadFile(filename)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}

	var info yaml.MapSlice
	yaml.Unmarshal(file, &info)
	return info
}

func main() {
	var input = flag.String("input", "", "OpenAPI source file to read")
	var rawInput = flag.Bool("raw", false, "Output the raw json input")
	var textProtobuf = flag.Bool("text", false, "Output a text protobuf representation")
	var jsonProtobuf = flag.Bool("json", false, "Output a json protobuf representation")
	var binaryProtobuf = flag.Bool("pb", false, "Output a binary protobuf representation")
	flag.Parse()

	if *input == "" {
		flag.PrintDefaults()
		return
	}

	fmt.Printf("Compiling %s (%s)\n", *input, openapi_v2.Version())

	raw := readFile(*input)
	if *rawInput {
		rawDescription := describeMap(raw, "")
		rawFileName := strings.TrimSuffix(path.Base(*input), path.Ext(*input)) + ".raw"
		ioutil.WriteFile(rawFileName, []byte(rawDescription), 0644)
	}

	document := openapi_v2.NewDocument(raw)

	if *textProtobuf {
		textProtoFileName := strings.TrimSuffix(path.Base(*input), path.Ext(*input)) + ".text"
		ioutil.WriteFile(textProtoFileName, []byte(proto.MarshalTextString(document)), 0644)
	}

	if *jsonProtobuf {
		jsonProtoFileName := strings.TrimSuffix(path.Base(*input), path.Ext(*input)) + ".json"
		jsonBytes, _ := json.Marshal(document)
		ioutil.WriteFile(jsonProtoFileName, jsonBytes, 0644)
	}

	if *binaryProtobuf {
		binaryProtoFileName := strings.TrimSuffix(path.Base(*input), path.Ext(*input)) + ".pb"
		protoBytes, _ := proto.Marshal(document)
		ioutil.WriteFile(binaryProtoFileName, protoBytes, 0644)
	}
}
