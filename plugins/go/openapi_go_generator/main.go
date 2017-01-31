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
	response := &plugins.PluginResponse{}
	response.Text = []string{}

	// Read the plugin input.
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Printf("File error: %v\n", err)
		os.Exit(1)
	}
	request := &plugins.PluginRequest{}
	err = proto.Unmarshal(data, request)

	// Loop over OpenAPI documents sent by the plugin and use them to generate client/server code.
	for _, wrapper := range request.Wrapper {
		document := &openapi.Document{}
		err = proto.Unmarshal(wrapper.Value, document)
		if err != nil {
			log.Printf("ERROR %v", err)
			os.Exit(1)
		}

		// Collect parameters passed to the plugin.
		invocation := os.Args[0]
		parameters := request.Parameters
		packageName := "main"
		for _, parameter := range parameters {
			invocation += " " + parameter.Name + "=" + parameter.Value
			if parameter.Name == "package" {
				packageName = parameter.Value
			}
		}
		log.Printf("Running %s", invocation)

		// Create the renderer.
		renderer, err := NewServiceRenderer(document, packageName)
		if err != nil {
			log.Printf("ERROR %v", err)
			os.Exit(1)
		}

		// Load templates.
		err = renderer.loadTemplates(templates())
		if err != nil {
			log.Printf("ERROR %v", err)
			os.Exit(1)
		}

		// Run the renderer to generate files and add them to the response object.
		err = renderer.Generate(response, files)
		if err != nil {
			log.Printf("ERROR %v", err)
			os.Exit(1)
		}
	}

	responseBytes, _ := proto.Marshal(response)
	os.Stdout.Write(responseBytes)
}
