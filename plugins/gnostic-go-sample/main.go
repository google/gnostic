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

// gnostic_go_sample is a sample Gnostic plugin written in Go.
package main

import (
	"errors"

	"github.com/googleapis/gnostic/printer"

	openapi2 "github.com/googleapis/gnostic/OpenAPIv2"
	plugins "github.com/googleapis/gnostic/plugins"
)

// generate a simple report of an OpenAPI document's contents
func printDocument(code *printer.Code, document *openapi2.Document) {
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
	env, err := plugins.NewEnvironment()
	env.RespondAndExitIfError(err)

	document, ok := env.Document.(*openapi2.Document)
	if !ok {
		err = errors.New("This plugin requires a valid OpenAPI 2 document")
	}
	env.RespondAndExitIfError(err)

	// generate report
	code := &printer.Code{}
	printDocument(code, document)
	file := &plugins.File{
		Name: "report.txt",
		Data: []byte(code.String()),
	}
	env.Response.Files = append(env.Response.Files, file)

	// send with success
	env.RespondAndExit()
}
