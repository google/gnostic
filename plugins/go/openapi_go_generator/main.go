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

//go:generate encode_templates

package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/golang/protobuf/proto"

	openapi "github.com/googleapis/openapi-compiler/OpenAPIv2"
	plugins "github.com/googleapis/openapi-compiler/plugins"
)

// if error is not nil, record it, serialize and return the response, and exit
func sendAndExitIfError(err error, response *plugins.Response) {
	if err != nil {
		response.Errors = append(response.Errors, err.Error())
		sendAndExit(response)
	}
}

// serialize and return the response
func sendAndExit(response *plugins.Response) {
	responseBytes, _ := proto.Marshal(response)
	os.Stdout.Write(responseBytes)
	os.Exit(0)
}

// This is the main function for the code generation plugin.
func main() {

	// Use the name used to run the plugin to decide which files to generate.
	var files []string
	switch os.Args[0] {
	case "openapi_go_client":
		files = []string{"client.go", "types.go"}
	case "openapi_go_server":
		files = []string{"server.go", "provider.go", "types.go"}
	default:
		files = []string{"client.go", "server.go", "provider.go", "types.go"}
	}

	// Initialize the plugin response.
	response := &plugins.Response{}

	// Read the plugin input.
	data, err := ioutil.ReadAll(os.Stdin)
	sendAndExitIfError(err, response)

	// Deserialize the input
	request := &plugins.Request{}
	err = proto.Unmarshal(data, request)
	sendAndExitIfError(err, response)

	// Read the document sent by the plugin and use it to generate client/server code.
	wrapper := request.Wrapper
	document := &openapi.Document{}
	err = proto.Unmarshal(wrapper.Value, document)
	sendAndExitIfError(err, response)

	// Collect parameters passed to the plugin.
	invocation := os.Args[0]
	parameters := request.Parameters
	packageName := request.OutputPath
	for _, parameter := range parameters {
		invocation += " " + parameter.Name + "=" + parameter.Value
		if parameter.Name == "package" {
			packageName = parameter.Value
		}
	}
	log.Printf("Running %s", invocation)

	// Create the renderer.
	renderer, err := NewServiceRenderer(document, packageName)
	sendAndExitIfError(err, response)

	// Load templates.
	err = renderer.loadTemplates(templates())
	sendAndExitIfError(err, response)

	// Run the renderer to generate files and add them to the response object.
	err = renderer.Generate(response, files)
	sendAndExitIfError(err, response)

	// Return with success.
	sendAndExit(response)
}
