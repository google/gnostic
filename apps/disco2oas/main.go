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
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/googleapis/gnostic/compiler"
	"github.com/googleapis/gnostic/discovery"
)

func main() {
	apiName := flag.String("api", "", "Specify the name of an API to export.")
	apiVersion := flag.String("version", "", "Specify the version of an API to export.")
	listAPIs := flag.Bool("list", false, "List all APIs available in the Google Discovery Service.")
	allAPIs := flag.Bool("all", false, "Export all APIs from the Google Discovery Service.")
	v2 := flag.Bool("v2", false, "Export APIs in the OpenAPI v2 format.")
	v3 := flag.Bool("v3", false, "Export APIs in the OpenAPI v3 format.")
	verbose := flag.Bool("verbose", false, "Export APIs verbosely.")

	flag.Parse()

	toolName := os.Args[0]
	if (*apiName == "" || *apiVersion == "") && !*listAPIs && !*allAPIs {
		fmt.Printf("Usage: %s --name=<api name> --version=<api version> [--v2]|[--v3] [--verbose]\n", toolName)
		fmt.Printf("       %s --list\n", toolName)
		fmt.Printf("       %s --all [--v2]|[--v3] [--verbose]\n", toolName)
		os.Exit(0)
	}
	if *v2 && *v3 {
		fmt.Printf("Please specify either --v2 or --v3, but not both.\n")
		os.Exit(0)
	}
	// Read the list of APIs from the apis/list service.
	bytes, err := compiler.FetchFile(discovery.APIsListServiceURL)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	// Unpack the apis/list response.
	listResponse, err := discovery.NewList(bytes)
	if *listAPIs {
		// List the APIs.
		for _, api := range listResponse.APIs {
			fmt.Printf("%s %s\n", api.Name, api.Version)
		}
	}
	// If an API was specified for export, convert it and write the result.
	if *apiName != "" && *apiVersion != "" {
		api := listResponse.APIWithID(*apiName + ":" + *apiVersion)
		if *v2 {
			fetchAndConvertAPI(api, *verbose, "v2")
		} else if *v3 {
			fetchAndConvertAPI(api, *verbose, "v3")
		} else {
			fetchAndConvertAPI(api, *verbose, "v3") // default
		}
	}
	// if all APIs were specified for export, convert and write all of them.
	if *allAPIs {
		for _, api := range listResponse.APIs {
			fmt.Printf("Fetching and converting %s %s\n", api.Name, api.Version)
			if *v2 {
				fetchAndConvertAPI(api, *verbose, "v2")
			} else if *v3 {
				fetchAndConvertAPI(api, *verbose, "v3")
			} else {
				fetchAndConvertAPI(api, *verbose, "v2") // default
			}
		}
	}
}

func fetchAndConvertAPI(api *discovery.API, verbose bool, version string) {
	// Get the description of an API
	if api == nil {
		log.Fatalf("Error: API not found")
	}
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
	// Generate the OpenAPI equivalent.
	switch version {
	case "v2":
		openAPIDocument, err := discoveryDocument.OpenAPIv2()
		if err != nil {
			panic(err)
		}
		bytes, err = proto.Marshal(openAPIDocument)
	case "v3":
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
	err = ioutil.WriteFile(api.Name+"-"+api.Version+".pb", bytes, 0644)
	if err != nil {
		panic(err)
	}
}
