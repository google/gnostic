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

// gnostic_go_generator is a sample Gnostic plugin that generates Go
// code that supports an API.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/golang/protobuf/proto"

	openapiv2 "github.com/googleapis/gnostic/OpenAPIv2"
	openapiv3 "github.com/googleapis/gnostic/OpenAPIv3"
	plugins "github.com/googleapis/gnostic/plugins"
)

var outputPath string // if nonempty, the plugin is run standalone

// respondAndExitIfError checks an error and if it is non-nil, records it and serializes and returns the response and then exits.
func respondAndExitIfError(err error, response *plugins.Response) {
	if err != nil {
		response.Errors = append(response.Errors, err.Error())
		respondAndExit(response)
	}
}

// respondAndExit serializes and returns the plugin response and then exits.
func respondAndExit(response *plugins.Response) {
	if outputPath != "" {
		err := plugins.HandleResponse(response, outputPath)
		if err != nil {
			log.Printf("%s", err.Error())
		}
	} else {
		responseBytes, _ := proto.Marshal(response)
		os.Stdout.Write(responseBytes)
	}
	os.Exit(0)
}

// This is the main function for the code generation plugin.
func main() {
	invocation := os.Args[0]

	// Use the name used to run the plugin to decide which files to generate.
	var files []string
	switch {
	case strings.Contains(invocation, "gnostic-go-client"):
		files = []string{"client.go", "types.go", "constants.go"}
	case strings.Contains(invocation, "gnostic-go-server"):
		files = []string{"server.go", "provider.go", "types.go", "constants.go"}
	default:
		files = []string{"client.go", "server.go", "provider.go", "types.go", "constants.go"}
	}

	// Initialize the plugin response.
	response := &plugins.Response{}
	var packageName string
	var documentv2 *openapiv2.Document
	var documentv3 *openapiv3.Document

	var version string
	var data []byte
	var err error

	if len(os.Args) == 1 {
		// Read the plugin input.
		data, err := ioutil.ReadAll(os.Stdin)
		respondAndExitIfError(err, response)
		if len(data) == 0 {
			respondAndExitIfError(fmt.Errorf("no input data"), response)
		}

		// Deserialize the input.
		request := &plugins.Request{}
		err = proto.Unmarshal(data, request)
		respondAndExitIfError(err, response)

		// Collect parameters passed to the plugin.
		parameters := request.Parameters
		packageName = request.OutputPath // the default package name is the output directory
		for _, parameter := range parameters {
			invocation += " " + parameter.Name + "=" + parameter.Value
			if parameter.Name == "package" {
				packageName = parameter.Value
			}
		}

		// Log the invocation.
		log.Printf("Running %s(input:%s)", invocation, request.Wrapper.Version)

		// Read the document sent by the plugin.
		version = request.Wrapper.Version
		data = request.Wrapper.Value
	} else {
		input := flag.String("input", "", "OpenAPI input (pb)")
		output := flag.String("output", "-", "output path")
		version2 := flag.Bool("v2", false, "OpenAPI version 2")
		version3 := flag.Bool("v3", false, "OpenAPI version 3")

		flag.Parse()
		outputPath = *output
		packageName = outputPath

		switch {
		case *version2:
			version = "v2"
		case *version3:
			version = "v3"
		default:
			version = "v2"
		}

		// Read the input document.
		data, err = ioutil.ReadFile(*input)
		if len(data) == 0 {
			respondAndExitIfError(fmt.Errorf("no input data"), response)
		}
	}

	switch version {
	case "v2":
		documentv2 = &openapiv2.Document{}
		err = proto.Unmarshal(data, documentv2)
		respondAndExitIfError(err, response)
	case "v3":
		documentv3 = &openapiv3.Document{}
		err = proto.Unmarshal(data, documentv3)
		respondAndExitIfError(err, response)
	default:
		err = fmt.Errorf("Unsupported OpenAPI version %s", version)
		respondAndExitIfError(err, response)
	}

	// Create the model.
	var model *ServiceModel
	if documentv2 != nil {
		model, err = NewServiceModelV2(documentv2, packageName)
	} else {
		model, err = NewServiceModelV3(documentv3, packageName)
	}
	respondAndExitIfError(err, response)

	// Create the renderer.
	renderer, err := NewServiceRenderer(model)
	respondAndExitIfError(err, response)

	// Run the renderer to generate files and add them to the response object.
	err = renderer.Generate(response, files)
	respondAndExitIfError(err, response)

	// Return with success.
	respondAndExit(response)
}
