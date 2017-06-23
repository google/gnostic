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

	"github.com/golang/protobuf/proto"
	"github.com/googleapis/gnostic/compiler"
	"github.com/googleapis/gnostic/discovery"
)

// Select an API.
const apiName = "people"
const apiVersion = "v1"

func main() {
	// Read the list of APIs from the apis/list service.
	apiListServiceURL := "https://www.googleapis.com/discovery/v1/apis"
	bytes, err := compiler.FetchFile(apiListServiceURL)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	// Unpack the apis/list response.
	listResponse, err := discovery.NewList(bytes)
	// List the APIs.
	for _, api := range listResponse.APIs {
		fmt.Printf("%s\n", api.ID)
	}
	// Get the description of an API
	api := listResponse.APIWithID(apiName + ":" + apiVersion)
	if api == nil {
		log.Fatalf("Error: API not found")
	}
	//fmt.Printf("API: %+v\n", api)
	// Fetch the discovery description of the API.
	bytes, err = compiler.FetchFile(api.DiscoveryRestURL)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	// Unpack the discovery response.
	discoveryDocument, err := discovery.NewDocument(bytes)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	//fmt.Printf("DISCOVERY: %+v\n", discoveryDocument)
	// Generate the OpenAPI equivalent
	openAPIDocument := discovery.BuildOpenAPI2DocumentForDocument(discoveryDocument)
	bytes, err = proto.Marshal(openAPIDocument)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(apiName+"-"+apiVersion+".pb", bytes, 0644)
	if err != nil {
		panic(err)
	}
}
