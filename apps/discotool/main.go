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

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/docopt/docopt-go"
	"github.com/golang/protobuf/proto"
	"github.com/googleapis/gnostic/compiler"
	"github.com/googleapis/gnostic/discovery"
)

func main() {
	usage := `discotool.

Usage: 
	discotool list
	discotool fetch <api> [<version>] [--out] [--openapi2] [--openapi3]
	discotool <file> [--out] [--openapi2] [--openapi3]
	`
	arguments, err := docopt.Parse(usage, nil, true, "Discotool 1.0", false)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	if arguments["list"].(bool) {
		// Read the list of APIs from the apis/list service.
		bytes, err := compiler.FetchFile(discovery.APIsListServiceURL)
		if err != nil {
			log.Fatalf("%+v", err)
		}
		// Unpack the apis/list response.
		listResponse, err := discovery.NewList(bytes)
		// List the APIs.
		for _, api := range listResponse.APIs {
			fmt.Printf("%s %s\n", api.Name, api.Version)
		}
	}

	if arguments["fetch"].(bool) {
		// Read the list of APIs from the apis/list service.
		bytes, err := compiler.FetchFile(discovery.APIsListServiceURL)
		if err != nil {
			log.Fatalf("%+v", err)
		}
		// Unpack the apis/list response.
		listResponse, err := discovery.NewList(bytes)
		// If an API was specified for export, convert it and write the result.
		if arguments["<api>"] != nil && arguments["<version>"] != nil {
			log.Printf("WTF")
			apiName := arguments["<api>"].(string)
			apiVersion := arguments["<version>"].(string)
			api := listResponse.APIWithID(apiName + ":" + apiVersion)
			// Get the description of an API
			if api == nil {
				log.Fatalf("Error: API not found")
			}
			// Fetch the discovery description of the API.
			bytes, err := compiler.FetchFile(api.DiscoveryRestURL)
			if err != nil {
				log.Fatalf("%+v", err)
			}
			if arguments["--out"].(bool) {
				fmt.Printf("%s\n", string(bytes))
			}
			outputname := api.Name + "-" + api.Version + ".pb"
			handleExportArgumentsForBytes(arguments, bytes, outputname)
		}
	}

	if arguments["<file>"] != nil {
		filename := arguments["<file>"].(string)
		bytes, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatalf("%+v", err)
		}
		outputname := strings.Replace(filename, ".json", ".pb", -1)
		handleExportArgumentsForBytes(arguments, bytes, outputname)
	}
}

func handleExportArgumentsForBytes(arguments map[string]interface{}, bytes []byte, outputname string) {
	if arguments["--openapi2"].(bool) && arguments["--openapi3"].(bool) {
		log.Fatalf("Error: Please specify either --openapi2 or --openapi3 but not both.")
	}
	if arguments["--openapi2"].(bool) || arguments["--openapi3"].(bool) {
		// Unpack the discovery response.
		discoveryDocument, err := discovery.NewDocument(bytes)
		if err != nil {
			log.Fatalf("%+v", err)
		}
		// Generate the OpenAPI equivalent.
		var bytes []byte
		switch {
		case arguments["--openapi2"].(bool):
			openAPIDocument, err := discoveryDocument.OpenAPIv2()
			if err != nil {
				panic(err)
			}
			bytes, err = proto.Marshal(openAPIDocument)
		case arguments["--openapi3"].(bool):
			openAPIDocument, err := discoveryDocument.OpenAPIv3()
			if err != nil {
				panic(err)
			}
			bytes, err = proto.Marshal(openAPIDocument)
		default:
		}
		if err != nil {
			panic(err)
		}
		err = ioutil.WriteFile(outputname, bytes, 0644)
		if err != nil {
			panic(err)
		}
	}
}
