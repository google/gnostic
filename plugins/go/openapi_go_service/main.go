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
	"path"
	"runtime"
	"strings"
	"text/template"

	"github.com/golang/protobuf/proto"
	"github.com/googleapis/openapi-compiler/printer"

	openapi "github.com/googleapis/openapi-compiler/OpenAPIv2"
	plugins "github.com/googleapis/openapi-compiler/plugins"
)

type FieldPosition int

const (
	FieldPositionBody     FieldPosition = iota
	FieldPositionHeader                 = iota
	FieldPositionFormData               = iota
	FieldPositionQuery                  = iota
	FieldPositionPath                   = iota
)

type ServiceType struct {
	Name   string
	Fields []*ServiceTypeField
}

type ServiceTypeField struct {
	Name     string
	Type     string
	JSONName string
	Position FieldPosition
}

type ServiceMethod struct {
	Name               string
	ParametersTypeName string
	ResponsesTypeName  string
	ParametersType     *ServiceType
	ResponsesType      *ServiceType

	Path        string
	Method      string
	Description string

	HandlerName   string
	ProcessorName string
}

type ServiceRenderer struct {
	templatedAppYaml         *template.Template
	templatedFileHeader      *template.Template
	templatedStruct          *template.Template
	templatedHandlerMethod   *template.Template
	templatedInitMethod      *template.Template
	templatedServiceHeader   *template.Template
	templatedProcessorMethod *template.Template

	name    string
	types   []*ServiceType
	methods []*ServiceMethod
}

func NewServiceRenderer(document *openapi.Document) (renderer *ServiceRenderer, err error) {
	renderer = &ServiceRenderer{}
	renderer.name = "myapp"
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

func truncate(in string) (out string) {
	limit := 18
	if len(in) <= limit {
		return in
	} else {
		return in[0:limit]
	}
}

func (renderer *ServiceRenderer) loadTemplate(name string) (err error) {

	_, filename, _, _ := runtime.Caller(1)
	TEMPLATES := path.Join(path.Dir(filename), "templates")

	renderer.templatedAppYaml, err = template.ParseFiles(TEMPLATES + "/app_yaml.tmpl")
	renderer.templatedAppYaml.Funcs(funcMap)
	if err != nil {
		return err
	}
}

// instantiate templates
func (renderer *ServiceRenderer) loadTemplates() (err error) {

	funcMap := template.FuncMap{
		"truncate": truncate,
	}

	_, filename, _, _ := runtime.Caller(1)
	TEMPLATES := path.Join(path.Dir(filename), "templates")

	renderer.templatedAppYaml, err = template.ParseFiles(TEMPLATES + "/app_yaml.tmpl")
	renderer.templatedAppYaml.Funcs(funcMap)
	if err != nil {
		return err
	}
	renderer.templatedFileHeader, err = template.ParseFiles(TEMPLATES + "/file_header.tmpl")
	if err != nil {
		return err
	}
	renderer.templatedStruct, err = template.ParseFiles(TEMPLATES + "/struct.tmpl")
	if err != nil {
		return err
	}
	renderer.templatedHandlerMethod, err = template.ParseFiles(TEMPLATES + "/handler_method.tmpl")
	if err != nil {
		return err
	}
	renderer.templatedInitMethod, err = template.ParseFiles(TEMPLATES + "/init_method.tmpl")
	if err != nil {
		return err
	}
	renderer.templatedServiceHeader, err = template.ParseFiles(TEMPLATES + "/service_header.tmpl")
	if err != nil {
		return err
	}
	renderer.templatedProcessorMethod, err = template.ParseFiles(TEMPLATES + "/processor_method.tmpl")
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
					f.Position = FieldPositionBody
				}
			}
			nonBodyParameter := parameter.GetNonBodyParameter()
			if nonBodyParameter != nil {
				headerParameter := nonBodyParameter.GetHeaderParameterSubSchema()
				if headerParameter != nil {
					f.Name = headerParameter.Name
					f.Position = FieldPositionHeader
				}
				formDataParameter := nonBodyParameter.GetFormDataParameterSubSchema()
				if formDataParameter != nil {
					f.Name = formDataParameter.Name
					f.Position = FieldPositionFormData
				}
				queryParameter := nonBodyParameter.GetQueryParameterSubSchema()
				if queryParameter != nil {
					f.Name = queryParameter.Name
					f.Position = FieldPositionQuery
				}
				pathParameter := nonBodyParameter.GetPathParameterSubSchema()
				if pathParameter != nil {
					f.Name = pathParameter.Name
					f.Position = FieldPositionPath
					f.Type = typeForName(pathParameter.Type, pathParameter.Format)
				}
			}
			f.JSONName = f.Name
			f.Name = strings.Title(f.Name)
			t.Fields = append(t.Fields, &f)
		}
	}
	t.Name = name
	renderer.types = append(renderer.types, t)
	return t, err
}

func propertyNameForResponseCode(code string) string {
	if code == "200" {
		return "OK"
	}
	return code
}

func (renderer *ServiceRenderer) loadServiceTypeFromResponses(name string, responses *openapi.Responses) (t *ServiceType, err error) {
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
		}
	}

	t.Name = name
	renderer.types = append(renderer.types, t)
	return t, err
}

func (renderer *ServiceRenderer) loadOperation(op *openapi.Operation, method string, path string) (err error) {
	var m ServiceMethod
	m.Name = strings.Title(op.OperationId)
	m.Path = path
	m.Method = method
	m.HandlerName = "handle" + m.Name
	m.ProcessorName = "process" + m.Name
	m.ParametersTypeName = m.Name + "Parameters"
	m.ResponsesTypeName = m.Name + "Responses"
	m.ParametersType, err = renderer.loadServiceTypeFromParameters(m.ParametersTypeName, op.Parameters)
	m.ResponsesType, err = renderer.loadServiceTypeFromResponses(m.ResponsesTypeName, op.Responses)
	renderer.methods = append(renderer.methods, &m)
	return err
}

// preprocess the types and methods of the API
func (renderer *ServiceRenderer) loadService(document *openapi.Document) (err error) {

	// collect service type descriptions
	renderer.types = make([]*ServiceType, 0)

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
		renderer.types = append(renderer.types, &t)
	}

	// collect service method descriptions
	renderer.methods = make([]*ServiceMethod, 0)
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
	renderer.writeAppYaml(response, "app.yaml")
	renderer.writeAppGo(response, "app.go")
	renderer.writeServiceGo(response, "/service.go-finishme")
	return err
}

func (renderer *ServiceRenderer) writeAppYaml(response *plugins.PluginResponse, filename string) (err error) {
	file := &plugins.File{}
	file.Name = filename
	f := new(bytes.Buffer)
	// file header
	err = renderer.templatedAppYaml.Execute(f, struct {
		Name string
	}{
		renderer.name,
	})
	if err != nil {
		response.Text = append(response.Text, fmt.Sprintf("ERROR %v", err))
	}
	file.Data = f.Bytes()
	response.File = append(response.File, file)
	return err
}

// write the non-editable app.go
func (renderer *ServiceRenderer) writeAppGo(response *plugins.PluginResponse, filename string) (err error) {
	file := &plugins.File{}
	file.Name = filename
	f := new(bytes.Buffer)

	// file header
	err = renderer.templatedFileHeader.Execute(f, struct {
		Name string
	}{
		renderer.name,
	})
	if err != nil {
		response.Text = append(response.Text, fmt.Sprintf("ERROR %v", err))
	}

	// type declarations
	for _, tt := range renderer.types {
		err = renderer.templatedStruct.Execute(f, struct {
			Name   string
			Fields []*ServiceTypeField
		}{
			filteredTypeName(tt.Name),
			tt.Fields,
		})
		if err != nil {
			response.Text = append(response.Text, fmt.Sprintf("ERROR %v", err))
		}
	}

	// request handlers
	for _, method := range renderer.methods {
		err = renderer.templatedHandlerMethod.Execute(f, struct {
			Name   string
			Method *ServiceMethod
		}{
			renderer.name,
			method,
		})
		if err != nil {
			response.Text = append(response.Text, fmt.Sprintf("ERROR %v", err))
		}
	}

	// init/main
	err = renderer.templatedInitMethod.Execute(f, struct {
		Name    string
		Methods []*ServiceMethod
	}{
		renderer.name,
		renderer.methods,
	})
	if err != nil {
		response.Text = append(response.Text, fmt.Sprintf("ERROR %v", err))
	}

	//err = exec.Command("/usr/local/go/bin/gofmt", "-w", filename).Run()
	file.Data = f.Bytes()
	response.File = append(response.File, file)
	return err
}

// write the user-editable service.go
func (renderer *ServiceRenderer) writeServiceGo(response *plugins.PluginResponse, filename string) (err error) {
	file := &plugins.File{}
	file.Name = filename
	f := new(bytes.Buffer)

	err = renderer.templatedServiceHeader.Execute(f, struct {
		Name string
	}{
		renderer.name,
	})
	if err != nil {
		response.Text = append(response.Text, fmt.Sprintf("ERROR %v", err))
	}

	for _, method := range renderer.methods {
		err = renderer.templatedProcessorMethod.Execute(f, struct {
			Name   string
			Method *ServiceMethod
		}{
			renderer.name,
			method,
		})
		if err != nil {
			response.Text = append(response.Text, fmt.Sprintf("ERROR %v", err))
		}
	}

	//err = exec.Command("/usr/local/go/bin/gofmt", "-w", filename).Run()
	file.Data = f.Bytes()
	response.File = append(response.File, file)
	return err
}

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
	log.Printf("OK")
	response := &plugins.PluginResponse{}
	response.Text = []string{}

	foreachDocumentFromPluginInput(
		func(name string, version string, document *openapi.Document) {

			code := &printer.Code{}
			code.Print("READING %s (%s)", name, version)
			printDocument(code, document)
			response.Text = append(response.Text, code.String())

			renderer, err := NewServiceRenderer(document)
			if err != nil {
				response.Text = append(response.Text, fmt.Sprintf("ERROR %v", err))
				return
			}
			err = renderer.GenerateApp(response)
			if err != nil {
				response.Text = append(response.Text, fmt.Sprintf("ERROR %v", err))
				return
			}
			response.Text = append(response.Text, fmt.Sprintf("DONE"))
		})

	responseBytes, _ := proto.Marshal(response)
	os.Stdout.Write(responseBytes)
}
