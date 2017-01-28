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
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/golang/protobuf/proto"

	openapi "github.com/googleapis/openapi-compiler/OpenAPIv2"
	plugins "github.com/googleapis/openapi-compiler/plugins"
)

type ServiceType struct {
	Name   string
	Fields []*ServiceTypeField
}

func (s *ServiceType) hasFieldNamed(name string) bool {
	for _, f := range s.Fields {
		if f.Name == name {
			return true
		}
	}
	return false
}

type ServiceTypeField struct {
	Name     string
	Type     string
	JSONName string
	Position string // "body", "header", "formdata", "query", or "path"
}

type ServiceMethod struct {
	Name               string
	Path               string
	Method             string
	Description        string
	HandlerName        string
	ProcessorName      string
	ClientName         string
	ResultTypeName     string
	ParametersTypeName string
	ResponsesTypeName  string
	ParametersType     *ServiceType
	ResponsesType      *ServiceType
}

type ServiceFileTemplate struct {
	FileName string
	Template *template.Template
}

type ServiceRenderer struct {
	Templates []*ServiceFileTemplate

	Name    string
	Types   []*ServiceType
	Methods []*ServiceMethod
}

func NewServiceRenderer(document *openapi.Document) (renderer *ServiceRenderer, err error) {
	renderer = &ServiceRenderer{}
	renderer.Name = "myapp"
	err = renderer.loadTemplates()
	if err != nil {
		return nil, err
	}
	err = renderer.loadService(document)
	if err != nil {
		return nil, err
	}
	return renderer, nil
}

// template support functions

func hasOKField(s *ServiceType) bool {
	return s.hasFieldNamed("OK")
}

func parameterList(m *ServiceMethod) string {
	result := ""
	for i, field := range m.ParametersType.Fields {
		if i > 0 {
			result += ", "
		}
		result += field.JSONName + " " + field.Type
	}
	return result
}

// instantiate templates

func (renderer *ServiceRenderer) loadTemplates() (err error) {
	renderer.Templates = make([]*ServiceFileTemplate, 0)

	funcMap := template.FuncMap{
		"HASOK":         hasOKField,
		"parameterList": parameterList,
	}

	_, filename, _, _ := runtime.Caller(1)
	TEMPLATES := path.Join(path.Dir(filename), "templates") + "/"

	files := []string{
		"types.go",
		"client.go",
	}
	for _, filename := range files {
		t, err := template.New(filename + ".tmpl").Funcs(funcMap).ParseFiles(TEMPLATES + filename + ".tmpl")
		if err != nil {
			log.Printf("ERROR: %+v", err)
			return err
		} else {
			renderer.Templates = append(renderer.Templates, &ServiceFileTemplate{FileName: filename, Template: t})
		}
	}
	return err
}

func (renderer *ServiceRenderer) loadServiceTypeFromParameters(name string, parameters []*openapi.ParametersItem) (t *ServiceType, err error) {
	t = &ServiceType{}
	t.Fields = make([]*ServiceTypeField, 0)
	for _, parametersItem := range parameters {
		var f ServiceTypeField
		f.Type = fmt.Sprintf("%+v", parametersItem)
		parameter := parametersItem.GetParameter()
		if parameter != nil {
			bodyParameter := parameter.GetBodyParameter()
			if bodyParameter != nil {
				f.Name = bodyParameter.Name
				if bodyParameter.Schema != nil {
					f.Type = typeForSchema(bodyParameter.Schema)
					f.Position = "body"
				}
			}
			nonBodyParameter := parameter.GetNonBodyParameter()
			if nonBodyParameter != nil {
				headerParameter := nonBodyParameter.GetHeaderParameterSubSchema()
				if headerParameter != nil {
					f.Name = headerParameter.Name
					f.Position = "header"
				}
				formDataParameter := nonBodyParameter.GetFormDataParameterSubSchema()
				if formDataParameter != nil {
					f.Name = formDataParameter.Name
					f.Position = "formdata"
				}
				queryParameter := nonBodyParameter.GetQueryParameterSubSchema()
				if queryParameter != nil {
					f.Name = queryParameter.Name
					f.Position = "query"
				}
				pathParameter := nonBodyParameter.GetPathParameterSubSchema()
				if pathParameter != nil {
					f.Name = pathParameter.Name
					f.Position = "path"
					f.Type = typeForName(pathParameter.Type, pathParameter.Format)
				}
			}
			f.JSONName = f.Name
			f.Name = strings.Title(f.Name)
			t.Fields = append(t.Fields, &f)
		}
	}
	t.Name = name
	renderer.Types = append(renderer.Types, t)
	return t, err
}

func propertyNameForResponseCode(code string) string {
	if code == "200" {
		return "OK"
	}
	return code
}

func (renderer *ServiceRenderer) loadServiceTypeFromResponses(m *ServiceMethod, name string, responses *openapi.Responses) (t *ServiceType, err error) {
	t = &ServiceType{}
	t.Fields = make([]*ServiceTypeField, 0)

	for _, responseCode := range responses.ResponseCode {
		var f ServiceTypeField
		f.Name = propertyNameForResponseCode(responseCode.Name)
		f.JSONName = ""
		response := responseCode.Value.GetResponse()
		if response != nil && response.Schema != nil && response.Schema.GetSchema() != nil {
			f.Type = "*" + typeForSchema(response.Schema.GetSchema())
			t.Fields = append(t.Fields, &f)
			if f.Name == "OK" {
				m.ResultTypeName = typeForSchema(response.Schema.GetSchema())
			}
		}
	}

	t.Name = name
	renderer.Types = append(renderer.Types, t)
	return t, err
}

func (renderer *ServiceRenderer) loadOperation(op *openapi.Operation, method string, path string) (err error) {
	var m ServiceMethod
	m.Name = strings.Title(op.OperationId)
	m.Path = path
	m.Method = method
	m.HandlerName = "handle" + m.Name
	m.ProcessorName = "process" + m.Name
	m.ClientName = "call" + m.Name
	m.ParametersTypeName = m.Name + "Parameters"
	m.ResponsesTypeName = m.Name + "Responses"
	m.ParametersType, err = renderer.loadServiceTypeFromParameters(m.ParametersTypeName, op.Parameters)
	m.ResponsesType, err = renderer.loadServiceTypeFromResponses(&m, m.ResponsesTypeName, op.Responses)
	renderer.Methods = append(renderer.Methods, &m)
	return err
}

// preprocess the types and methods of the API
func (renderer *ServiceRenderer) loadService(document *openapi.Document) (err error) {
	// collect service type descriptions
	renderer.Types = make([]*ServiceType, 0)
	for _, pair := range document.Definitions.AdditionalProperties {
		var t ServiceType
		t.Fields = make([]*ServiceTypeField, 0)
		schema := pair.Value
		for _, pair2 := range schema.Properties.AdditionalProperties {
			var f ServiceTypeField
			f.Name = strings.Title(pair2.Name)
			f.Type = typeForSchema(pair2.Value)
			f.JSONName = pair2.Name
			t.Fields = append(t.Fields, &f)
		}
		t.Name = strings.Title(filteredTypeName(pair.Name))
		renderer.Types = append(renderer.Types, &t)
	}
	// collect service method descriptions
	renderer.Methods = make([]*ServiceMethod, 0)
	for _, pair := range document.Paths.Path {
		v := pair.Value
		if v.Get != nil {
			renderer.loadOperation(v.Get, "GET", pair.Name)
		}
		if v.Post != nil {
			renderer.loadOperation(v.Post, "POST", pair.Name)
		}
		if v.Put != nil {
			renderer.loadOperation(v.Put, "PUT", pair.Name)
		}
		if v.Delete != nil {
			renderer.loadOperation(v.Delete, "DELETE", pair.Name)
		}
	}
	return err
}

func filteredTypeName(typeName string) (name string) {
	// first take the last path segment
	parts := strings.Split(typeName, "/")
	name = parts[len(parts)-1]
	// then take the last part of a dotted name
	parts = strings.Split(name, ".")
	name = parts[len(parts)-1]
	return name
}

func typeForName(name string, format string) (typeName string) {
	switch name {
	case "integer":
		if format == "int32" {
			return "int32"
		} else if format == "int64" {
			return "int64"
		} else {
			return "int32"
		}
	default:
		return name
	}
}

func typeForSchema(schema *openapi.Schema) (typeName string) {
	ref := schema.XRef
	if ref != "" {
		return typeForRef(ref)
	}
	if schema.Type != nil {
		types := schema.Type.Value
		if len(types) == 1 && types[0] == "string" {
			return "string"
		}
		if len(types) == 1 && types[0] == "array" && schema.Items != nil {
			// we have an array.., but of what?
			items := schema.Items.Schema
			if len(items) == 1 && items[0].XRef != "" {
				return "[]" + typeForRef(items[0].XRef)
			}
		}
	}
	// this function is incomplete... so return a string representing anything that we don't handle
	return fmt.Sprintf("%v", schema)
}

func typeForRef(ref string) (typeName string) {
	return strings.Title(path.Base(ref))
}

func (renderer *ServiceRenderer) GenerateApp(response *plugins.PluginResponse) (err error) {
	for _, pair := range renderer.Templates {
		file := &plugins.File{}
		file.Name = pair.FileName
		f := new(bytes.Buffer)
		// file header
		err = pair.Template.Execute(f, struct {
			Renderer *ServiceRenderer
			Package  string
		}{
			renderer,
			"main",
		})
		if err != nil {
			response.Text = append(response.Text, fmt.Sprintf("ERROR %v", err))
		}
		inputBytes := f.Bytes()
		if filepath.Ext(file.Name) == ".go" {
			cmd := exec.Command(runtime.GOROOT() + "/bin/gofmt")
			input, _ := cmd.StdinPipe()
			output, _ := cmd.StdoutPipe()
			cmderr, _ := cmd.StderrPipe()
			err := cmd.Start()
			if err != nil {
				log.Printf("Error: %+v", err)
			}
			input.Write(inputBytes)
			input.Close()
			errors, err := ioutil.ReadAll(cmderr)
			if len(errors) > 0 {
				log.Printf(string(errors))
				file.Data = inputBytes
			} else {
				file.Data, err = ioutil.ReadAll(output)
				if err != nil {
					log.Printf("Error: %+v", err)
					file.Data = inputBytes
				}
			}
		} else {
			file.Data = inputBytes
		}
		response.File = append(response.File, file)
	}
	return
}

// process plugin
type documentHandler func(name string, version string, document *openapi.Document)

func foreachDocumentFromPluginInput(handler documentHandler) {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Printf("File error: %v\n", err)
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

func main() {
	response := &plugins.PluginResponse{}
	response.Text = []string{}

	foreachDocumentFromPluginInput(
		func(name string, version string, document *openapi.Document) {
			log.Printf("Reading %s (%s)", name, version)
			renderer, err := NewServiceRenderer(document)
			if err != nil {
				log.Printf("ERROR %v", err)
				return
			}
			err = renderer.GenerateApp(response)
			if err != nil {
				log.Printf("ERROR %v", err)
				return
			}
		})

	responseBytes, _ := proto.Marshal(response)
	os.Stdout.Write(responseBytes)
}
