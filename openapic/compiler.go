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

//go:generate ./COMPILE-PROTOS.sh

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"

	"github.com/golang/protobuf/proto"
	"github.com/googleapis/openapi-compiler/OpenAPIv2"
	"github.com/googleapis/openapi-compiler/compiler"
	plugins "github.com/googleapis/openapi-compiler/plugins"
)

type PluginCall struct {
	Name   string
	Output string
}

func writeFile(name string, bytes []byte) {
	var writer io.Writer
	if name == "-" {
		writer = os.Stdout
	} else {
		file, _ := os.Create(name)
		defer file.Close()
		writer = file
	}
	writer.Write(bytes)
	if name == "-" {
		writer.Write([]byte("\n"))
	}
}

func main() {
	pb_regex, err := regexp.Compile("--pb_out=(.+)")
	json_regex, err := regexp.Compile("--json_out=(.+)")
	text_regex, err := regexp.Compile("--text_out=(.+)")
	errors_regex, err := regexp.Compile("--errors_out=(.+)")
	plugin_regex, err := regexp.Compile("--(.+)_out=(.+)")

	sourceName := ""
	binaryProtoFileName := ""
	jsonProtoFileName := ""
	textProtoFileName := ""
	errorFileName := ""
	keepReferences := false

	var pluginCalls []*PluginCall

	for i, arg := range os.Args {
		if i == 0 {
			continue
		}
		argbytes := []byte(arg)
		var m [][]byte
		if m = pb_regex.FindSubmatch(argbytes); m != nil {
			binaryProtoFileName = string(m[1])
		} else if m = json_regex.FindSubmatch(argbytes); m != nil {
			jsonProtoFileName = string(m[1])
		} else if m = text_regex.FindSubmatch(argbytes); m != nil {
			textProtoFileName = string(m[1])
		} else if m = errors_regex.FindSubmatch(argbytes); m != nil {
			errorFileName = string(m[1])
		} else if m = plugin_regex.FindSubmatch(argbytes); m != nil {
			pluginCall := &PluginCall{Name: string(m[1]), Output: string(m[2])}
			pluginCalls = append(pluginCalls, pluginCall)
		} else if arg == "--keep_refs" {
			keepReferences = true
		} else {
			sourceName = arg
		}
	}

	usage := `
Usage: openapic OPENAPI_SOURCE [OPTIONS]
  OPENAPI_SOURCE is the filename or URL of the OpenAPI description to read.
Options:
  --pb_out=FILENAME       Write a binary proto to a file with the specified name.
  --text_out=FILENAME     Write a text proto to a file with the specified name.
  --json_out=FILENAME     Write a json proto to a file with the specified name.
  --errors_out=FILENAME   Write compilation errors to a file with the specified name.
  --PLUGIN_out=FILENAME   Run the plugin named openapi_PLUGIN and write results to a file with the specified name.
  --keep_refs             Disable resolution of $ref references.
`

	if textProtoFileName == "" &&
		jsonProtoFileName == "" &&
		binaryProtoFileName == "" &&
		errorFileName == "" &&
		len(pluginCalls) == 0 {
		fmt.Printf("Missing output directives.\n%s\n", usage)
		os.Exit(-1)
	}

	if sourceName == "" {
		fmt.Printf("No input specified.\n%s\n", usage)
		os.Exit(-1)
	}

	if errorFileName == "" {
		errorFileName = "-"
	}

	raw, err := compiler.ReadFile(sourceName)
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
		os.Exit(-1)
	}

	document, err := openapi_v2.NewDocument(raw, compiler.NewContext("$root", nil))
	if err != nil {
		writeFile(errorFileName, []byte(err.Error()))
		os.Exit(-1)
	}

	if !keepReferences {
		_, err = document.ResolveReferences(sourceName)
		if err != nil {
			writeFile(errorFileName, []byte(err.Error()))
			os.Exit(-1)
		}
	}

	if textProtoFileName != "" {
		bytes := []byte(proto.MarshalTextString(document))
		writeFile(textProtoFileName, bytes)
	}

	if jsonProtoFileName != "" {
		jsonBytes, _ := json.Marshal(document)
		writeFile(jsonProtoFileName, jsonBytes)
	}

	if binaryProtoFileName != "" {
		protoBytes, _ := proto.Marshal(document)
		writeFile(binaryProtoFileName, protoBytes)
	}

	for _, pluginCall := range pluginCalls {
		if pluginCall.Name != "" {
			request := &plugins.PluginRequest{}
			request.Parameter = ""

			version := &plugins.Version{}
			version.Major = 0
			version.Minor = 1
			version.Patch = 0
			request.CompilerVersion = version

			wrapper := &plugins.Wrapper{}
			wrapper.Name = sourceName
			wrapper.Version = "v2"
			protoBytes, _ := proto.Marshal(document)
			wrapper.Value = protoBytes
			request.Wrapper = []*plugins.Wrapper{wrapper}
			requestBytes, _ := proto.Marshal(request)

			cmd := exec.Command("openapi_" + pluginCall.Name)
			cmd.Stdin = bytes.NewReader(requestBytes)
			output, err := cmd.Output()
			if err != nil {
				fmt.Printf("Error: %+v\n", err)
			}
			response := &plugins.PluginResponse{}
			err = proto.Unmarshal(output, response)

			var writer io.Writer
			if pluginCall.Output == "-" {
				writer = os.Stdout
			} else {
				file, _ := os.Create(pluginCall.Output)
				defer file.Close()
				writer = file
			}
			for _, text := range response.Text {
				writer.Write([]byte(text))
			}
		}
	}
}
