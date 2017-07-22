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
	_ "os"
	"path/filepath"

	plugins "github.com/googleapis/gnostic/plugins"
)

const newline = "\n"

// ServiceRenderer generates code for a ServiceModel.
type ServiceRenderer struct {
	Model     *ServiceModel
}

// NewServiceRenderer creates a renderer.
func NewServiceRenderer(model *ServiceModel) (renderer *ServiceRenderer, err error) {
	renderer = &ServiceRenderer{}
	renderer.Model = model
	return renderer, nil
}

// Generate runs the renderer to generate the named files.
func (renderer *ServiceRenderer) Generate(response *plugins.Response, files []string) (err error) {
	for _, filename := range files {
		file := &plugins.File{Name: filename}
		switch filename {
		case "client.go":
			file.Data, err = renderer.GenerateClient()
		case "types.go":
			file.Data, err = renderer.GenerateTypes()
		case "provider.go":
			file.Data, err = renderer.GenerateProvider()
		case "server.go":
			file.Data, err = renderer.GenerateServer()
		default:
			file.Data = nil
		}
		if err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("ERROR %v", err))
		}
		// run generated Go files through gofmt
		if filepath.Ext(file.Name) == ".go" {
			file.Data, err = gofmt(file.Name, file.Data)
		}
		response.Files = append(response.Files, file)
	}
	return
}
