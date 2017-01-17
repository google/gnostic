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

type ServiceType struct {
	Name   string
	Fields []*ServiceTypeField
}

type ServiceTypeField struct {
	Name     string
	Type     string
	JSONName string
}

type ServiceMethod struct {
	Name             string
	RequestTypeName  string
	ResponseTypeName string
	RequestType      *ServiceType
	ResponseType     *ServiceType

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

// instantiate templates
func (renderer *ServiceRenderer) loadTemplates() (err error) {

	_, filename, _, _ := runtime.Caller(1)
	TEMPLATES := path.Join(path.Dir(filename), "templates")

	renderer.templatedAppYaml, err = template.ParseFiles(TEMPLATES + "/app_yaml.tmpl")
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

func (renderer *ServiceRenderer) loadServiceType(name string) (t *ServiceType, err error) {
	t = &ServiceType{}
	t.Fields = make([]*ServiceTypeField, 0)
	/*
		for _, field := range serviceType.Fields {
			var f ServiceTypeField
			f.Name = strings.Title(field.Name)
			f.Type = typeForField(field)
			f.JSONName = field.JsonName
			t.Fields = append(t.Fields, &f)
		}
	*/
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
	m.RequestTypeName = m.Name + "Request"
	m.ResponseTypeName = m.Name + "Response"
	m.RequestType, err = renderer.loadServiceType(m.RequestTypeName)
	m.ResponseType, err = renderer.loadServiceType(m.ResponseTypeName)
	renderer.methods = append(renderer.methods, &m)
	return err
}

// preprocess the types and methods of the API
func (renderer *ServiceRenderer) loadService(document *openapi.Document) (err error) {

	// collect service type descriptions
	renderer.types = make([]*ServiceType, 0)

	/*
		for _, serviceType := range service.Types {
			var t ServiceType
			t.Fields = make([]*ServiceTypeField, 0)
			for _, field := range serviceType.Fields {
				var f ServiceTypeField
				f.Name = strings.Title(field.Name)
				f.Type = typeForField(field)
				f.JSONName = field.JsonName
				t.Fields = append(t.Fields, &f)
			}
			t.Name = filteredTypeName(serviceType.Name)
			renderer.types = append(renderer.types, &t)
		}
	*/

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

/*
func typeForField(field *protobuf.Field) (typeName string) {
	var prefix string
	if field.Cardinality == protobuf.Field_CARDINALITY_REPEATED {
		prefix = "[]"
	}

	name := "unknown"
	switch field.Kind {
	case protobuf.Field_TYPE_STRING:
		name = "string"
	case protobuf.Field_TYPE_INT64:
		name = "int64"
	case protobuf.Field_TYPE_INT32:
		name = "int32"
	case protobuf.Field_TYPE_BOOL:
		name = "bool"
	case protobuf.Field_TYPE_DOUBLE:
		name = "float64"
	case protobuf.Field_TYPE_ENUM:
		name = "int"
	case protobuf.Field_TYPE_MESSAGE:
		name = filteredTypeName(field.TypeUrl)
	}
	if name == "unknown" {
		log.Printf(">> UNKNOWN TYPE FOR FIELD %+v", field)
	}
	return prefix + name
}
*/

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
			}
			err = renderer.GenerateApp(response)
			if err != nil {
				response.Text = append(response.Text, fmt.Sprintf("ERROR %v", err))
			}
			response.Text = append(response.Text, fmt.Sprintf("DONE"))
		})

	responseBytes, _ := proto.Marshal(response)
	os.Stdout.Write(responseBytes)
}
