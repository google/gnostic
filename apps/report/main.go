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

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/golang/protobuf/proto"

	pb "openapi"
)

func readFile(filename string) *pb.Document {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}
	document := &pb.Document{}
	err = proto.Unmarshal(data, document)
	if err != nil {
		panic(err)
	}
	return document
}

func main() {
	var input = flag.String("input", "", "OpenAPI source file to read")
	flag.Parse()

	if *input == "" {
		flag.PrintDefaults()
		return
	}

	document := readFile(*input)

	fmt.Printf("\n")
	fmt.Printf("Info\n")
	fmt.Printf("  Title %s\n", document.Info.Title)
	fmt.Printf("  Description %s\n", document.Info.Description)
	fmt.Printf("  Version %s\n", document.Info.Version)
	fmt.Printf("  TermsOfService %s\n", document.Info.TermsOfService)
	fmt.Printf("  Contact Email %s\n", document.Info.Contact.Email)
	fmt.Printf("  License Name %s\n", document.Info.License.Name)
	fmt.Printf("  License URL %s\n", document.Info.License.Url)
	fmt.Printf("\n")
	fmt.Printf("BasePath %+v\n", document.BasePath)
	fmt.Printf("\n")

	fmt.Printf("Paths\n")
	for k, v := range document.Paths.Path {
		fmt.Printf("  %+v\n", k)
		if v.Get != nil {
			fmt.Printf("    GET %+v\n", v.Get.OperationId)
			if v.Get.Description != "" {
				fmt.Printf("      %s\n", v.Get.Description)
			}
		}
		if v.Post != nil {
			fmt.Printf("    POST %+v\n", v.Post.OperationId)
			if v.Post.Description != "" {
				fmt.Printf("      %s\n", v.Post.Description)
			}
		}
		fmt.Printf("\n")
	}
}
