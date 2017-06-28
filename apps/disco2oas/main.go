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
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/googleapis/gnostic/compiler"
	"github.com/googleapis/gnostic/discovery"
)

func main() {
	var apiName string
	var apiVersion string
	var listAPIs bool
	var allAPIs bool
	if len(os.Args) == 3 {
		apiName = os.Args[1]
		apiVersion = os.Args[2]
	} else if len(os.Args) == 2 && (os.Args[1] == "--list") {
		listAPIs = true
	} else if len(os.Args) == 2 && (os.Args[1] == "--all") {
		allAPIs = true
	}
	if apiName == "" && !listAPIs && !allAPIs {
		fmt.Printf("Usage: discovery <api name> <api version>\n")
		fmt.Printf("       discovery --list\n")
		fmt.Printf("       discovery --all\n")
		os.Exit(0)
	}
	// Read the list of APIs from the apis/list service.
	bytes, err := compiler.FetchFile(discovery.APIsListServiceURL)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	// Unpack the apis/list response.
	listResponse, err := discovery.NewList(bytes)
	if listAPIs {
		// List the APIs.
		for _, api := range listResponse.APIs {
			fmt.Printf("%s %s\n", api.Name, api.Version)
		}
	}

	if apiName != "" && apiVersion != "" {
		api := listResponse.APIWithID(apiName + ":" + apiVersion)
		fetchAndConvertAPI(api, true)
	}

	if allAPIs {
		for _, api := range listResponse.APIs {
			fmt.Printf("Fetching and converting %s %s\n", api.Name, api.Version)
			fetchAndConvertAPI(api, false)
		}
	}
}

func fetchAndConvertAPI(api *discovery.API, verbose bool) {
	// Get the description of an API
	if api == nil {
		log.Fatalf("Error: API not found")
	}
	//fmt.Printf("API: %+v\n", api)
	// Fetch the discovery description of the API.
	bytes, err := compiler.FetchFile(api.DiscoveryRestURL)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if verbose {
		fmt.Printf("%s\n", string(bytes))
	}
	// Unpack the discovery response.
	discoveryDocument, err := discovery.NewDocument(bytes)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	//fmt.Printf("DISCOVERY: %+v\n", discoveryDocument)
	// Generate the OpenAPI equivalent
	openAPIDocument, err := discoveryDocument.OpenAPIv2()
	bytes, err = proto.Marshal(openAPIDocument)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(api.Name+"-"+api.Version+".pb", bytes, 0644)
	if err != nil {
		panic(err)
	}
}
