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
	"github.com/googleapis/openapi-compiler/printer"

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

func printSchema(code *printer.Code, schema *pb.Schema) {
	code.Print("Properties")
	code.Indent()
	for k, v := range schema.Properties.AdditionalProperties {
		code.Print("%s %s", k, v)
	}
	code.Outdent()
}

func main() {
	var input = flag.String("input", "", "OpenAPI source file to read")
	flag.Parse()

	if *input == "" {
		flag.PrintDefaults()
		return
	}

	document := readFile(*input)

	code := &printer.Code{}
	code.Print("API REPORT")
	code.Print("----------")
	code.Print("BasePath %+v", document.BasePath)
	code.Print("Consumes %+v", document.Consumes)
	code.Print("Definitions")
	code.Indent()
	for k, v := range document.Definitions.AdditionalProperties {
		code.Print("%s", k)
		code.Indent()
		printSchema(code, v)
		code.Outdent()
	}
	code.Outdent()
	code.Print("ExternalDocs %+v", document.ExternalDocs)
	code.Print("Host %+v", document.Host)
	code.Print("Info")
	code.Print("Title %s", document.Info.Title)
	code.Print("Description %s", document.Info.Description)
	code.Print("Version %s", document.Info.Version)
	code.Print("TermsOfService %s", document.Info.TermsOfService)
	code.Print("Contact Email %s", document.Info.Contact.Email)
	code.Print("License Name %s", document.Info.License.Name)
	code.Print("License URL %s", document.Info.License.Url)
	code.Print("Parameters %+v", document.Parameters)
	code.Print("Paths")
	for k, v := range document.Paths.Path {
		code.Print("%+v", k)
		if v.Get != nil {
			code.Print("GET %+v", v.Get.OperationId)
			if v.Get.Description != "" {
				code.Print("%s", v.Get.Description)
			}
		}
		if v.Post != nil {
			code.Print("POST %+v", v.Post.OperationId)
			if v.Post.Description != "" {
				code.Print("%s", v.Post.Description)
			}
		}
	}
	code.Print("Produces %+v", document.Produces)
	code.Print("Responses %+v", document.Responses)
	code.Print("Schemes %+v", document.Schemes)
	code.Print("Security %+v", document.Security)
	code.Print("SecurityDefinitions %+v", document.SecurityDefinitions)
	code.Print("Swagger %+v", document.Swagger)
	code.Print("Tags %+v", document.Tags)

	code.Print("")

	fmt.Printf("%s", code.Text())

}
