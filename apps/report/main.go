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

func describeSchema(schema *pb.Schema, indent string) string {
	result := ""
	result += fmt.Sprintf(indent + "Properties\n")
	for k, v := range schema.Properties.AdditionalProperties {
		indent2 := indent + "  "
		result += fmt.Sprintf(indent2+"%s %s\n", k, v)
	}
	return result
}

func main() {
	var input = flag.String("input", "", "OpenAPI source file to read")
	flag.Parse()

	if *input == "" {
		flag.PrintDefaults()
		return
	}

	indent := "-"

	document := readFile(*input)

	fmt.Printf(indent+"BasePath %+v\n", document.BasePath)
	fmt.Printf(indent+"Consumes %+v\n", document.Consumes)
	fmt.Printf(indent + "Definitions\n")
	for k, v := range document.Definitions.AdditionalProperties {
		fmt.Printf(indent+"%s\n", k)
		fmt.Printf(indent+"%s\n", describeSchema(v, indent+"  "))
	}
	fmt.Printf(indent+"ExternalDocs %+v\n", document.ExternalDocs)
	fmt.Printf(indent+"Host %+v\n", document.Host)
	fmt.Printf(indent + "Info\n")
	fmt.Printf(indent+"Title %s\n", document.Info.Title)
	fmt.Printf(indent+"Description %s\n", document.Info.Description)
	fmt.Printf(indent+"Version %s\n", document.Info.Version)
	fmt.Printf(indent+"TermsOfService %s\n", document.Info.TermsOfService)
	fmt.Printf(indent+"Contact Email %s\n", document.Info.Contact.Email)
	fmt.Printf(indent+"License Name %s\n", document.Info.License.Name)
	fmt.Printf(indent+"License URL %s\n", document.Info.License.Url)
	fmt.Printf(indent+"Parameters %+v\n", document.Parameters)
	fmt.Printf(indent + "Paths\n")
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
	fmt.Printf("Produces %+v\n", document.Produces)
	fmt.Printf("Responses %+v\n", document.Responses)
	fmt.Printf("Schemes %+v\n", document.Schemes)
	fmt.Printf("Security %+v\n", document.Security)
	fmt.Printf("SecurityDefinitions %+v\n", document.SecurityDefinitions)
	fmt.Printf("Swagger %+v\n", document.Swagger)
	fmt.Printf("Tags %+v\n", document.Tags)

	fmt.Printf("\n")

}
