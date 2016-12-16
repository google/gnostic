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
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/googleapis/openapi-compiler/OpenAPIv2"
	"github.com/googleapis/openapi-compiler/compiler"
)

func main() {
	var input = flag.String("in", "", "OpenAPI source file to read")
	var rawInput = flag.Bool("raw", false, "Output the raw json input")
	var textProtobuf = flag.Bool("text", false, "Output a text protobuf representation")
	var jsonProtobuf = flag.Bool("json", false, "Output a json protobuf representation")
	var binaryProtobuf = flag.Bool("pb", false, "Output a binary protobuf representation")
	var keepReferences = flag.Bool("keep-refs", false, "Disable resolution of $ref references")
	flag.Parse()

	if *input == "" {
		flag.PrintDefaults()
		return
	}

	fmt.Printf("Compiling %s (%s)\n", *input, openapi_v2.Version())

	raw := compiler.ReadFile(*input)
	if *rawInput {
		rawDescription := compiler.DescribeMap(raw, "")
		rawFileName := strings.TrimSuffix(path.Base(*input), path.Ext(*input)) + ".raw"
		ioutil.WriteFile(rawFileName, []byte(rawDescription), 0644)
	}

	document, err := openapi_v2.NewDocument(raw, nil)
	if err != nil {
		fmt.Printf("Error(s):\n%+v\n", err)
		os.Exit(-1)
	}

	if !*keepReferences {
		_, err = document.ResolveReferences(*input)
		if err != nil {
			fmt.Printf("Error %+v\n", err)
		}
	}

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
