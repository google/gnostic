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
	"fmt"
	"io/ioutil"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/googleapis/openapi-compiler/printer"

	openapi "github.com/googleapis/openapi-compiler/OpenAPIv2"
	plugins "github.com/googleapis/openapi-compiler/plugins"
)

type documentHandler func(name string, version string, document *openapi.Document)

func foreachDocumentFromPluginInput(handler documentHandler) {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}
	request := &plugins.PluginRequest{}
	err = proto.Unmarshal(data, request)
	for _, wrapper := range request.Wrapper {
		document := &openapi.Document{}
		err = proto.Unmarshal(wrapper.Value, document)
		if err != nil {
			panic(err)
		}
		handler(wrapper.Name, wrapper.Version, document)
	}
}

func printDocument(code *printer.Code, document *openapi.Document) {
	code.Print("Swagger: %+v", document.Swagger)
	code.Print("Host: %+v", document.Host)
	code.Print("BasePath: %+v", document.BasePath)
	if document.Info != nil {
		code.Print("Info:")
		code.Indent()
		if document.Info.Title != "" {
			code.Print("Title: %s", document.Info.Title)
		}
		if document.Info.Description != "" {
			code.Print("Description: %s", document.Info.Description)
		}
		if document.Info.Version != "" {
			code.Print("Version: %s", document.Info.Version)
		}
		code.Outdent()
	}
	code.Print("Paths:")
	code.Indent()
	for _, pair := range document.Paths.Path {
		v := pair.Value
		if v.Get != nil {
			code.Print("GET %+v", pair.Name)
		}
		if v.Post != nil {
			code.Print("POST %+v", pair.Name)
		}
	}
	code.Outdent()
}

func main() {
	response := &plugins.PluginResponse{}
	response.Text = []string{}

	foreachDocumentFromPluginInput(
		func(name string, version string, document *openapi.Document) {
			code := &printer.Code{}
			code.Print("READING %s (%s)", name, version)
			printDocument(code, document)
			response.Text = append(response.Text, code.String())
		})

	responseBytes, _ := proto.Marshal(response)
	os.Stdout.Write(responseBytes)
}
