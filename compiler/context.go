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

package compiler

import (
	"bytes"
	"fmt"
	"os/exec"

	"strings"

	"errors"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	ext_plugin "github.com/googleapis/gnostic/extension/extension_data"
	yaml "gopkg.in/yaml.v2"
)

type CustomAnyProtoGenerator struct {
	GeneratorName string
}

func (customAnyProtoGenerator *CustomAnyProtoGenerator) Perform(in interface{}, extensionName string) (*any.Any, error) {
	if customAnyProtoGenerator.GeneratorName != "" {
		binary, _ := yaml.Marshal(in)

		request := &ext_plugin.VendorExtensionHandlerRequest{}
		request.Parameter = ""

		version := &ext_plugin.Version{}
		version.Major = 0
		version.Minor = 1
		version.Patch = 0
		request.CompilerVersion = version

		request.Wrapper = &ext_plugin.Wrapper{}
		request.Wrapper.Name = "TESTETEST"
		request.Wrapper.Version = "v2"

		request.Wrapper.Yaml = string(binary)
		request.Wrapper.ExtensionName = extensionName
		requestBytes, _ := proto.Marshal(request)

		cmd := exec.Command(customAnyProtoGenerator.GeneratorName)
		cmd.Stdin = bytes.NewReader(requestBytes)
		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error: %+v\n", err)
			return nil, err
		}
		response := &ext_plugin.VendorExtensionHandlerResponse{}
		err = proto.Unmarshal(output, response)
		if err != nil {
			fmt.Printf("Error: %+v\n", err)
			fmt.Printf("%s\n", string(output))
			return nil, err
		}
		if !response.Handled {
			return nil, nil
		}
		if len(response.Error) != 0 {
			message := fmt.Sprintf("Errors when parsing: %+v for field %s by vendor extension handler %s. Details %+v", in, extensionName, customAnyProtoGenerator.GeneratorName, strings.Join(response.Error, ","))
			return nil, errors.New(message)
		}
		return response.Value, nil
	}
	return nil, nil
}

type Context struct {
	Parent *Context
	Name   string

	// TODO: Figure out a better way to pass the CustomAnyProtoGenerator to the generated compiler.
	CustomAnyProtoGenerators *[]CustomAnyProtoGenerator
}

func NewContextWithCustomAnyProtoGenerators(name string, parent *Context, customAnyProtoGenerators *[]CustomAnyProtoGenerator) *Context {
	return &Context{Name: name, Parent: parent, CustomAnyProtoGenerators: customAnyProtoGenerators}
}

func NewContext(name string, parent *Context) *Context {
	return &Context{Name: name, Parent: parent, CustomAnyProtoGenerators: parent.CustomAnyProtoGenerators}
}

func (context *Context) Description() string {
	if context.Parent != nil {
		return context.Parent.Description() + "." + context.Name
	} else {
		return context.Name
	}
}
