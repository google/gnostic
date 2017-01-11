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
	"io/ioutil"
	"log"
	"os"
	"os/exec"
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

// preprocess the types and methods of the API
func (renderer *ServiceRenderer) loadService(document *openapi.Document) (err error) {

	/*
		// collect service type descriptions
		renderer.types = make([]*ServiceType, 0)
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

		// collect service method descriptions
		renderer.methods = make([]*ServiceMethod, 0)
		for _, api := range service.Apis {
			for _, method := range api.Methods {
				var m ServiceMethod
				m.Name = method.Name
				m.RequestTypeName = filteredTypeName(method.RequestTypeUrl)
				m.ResponseTypeName = filteredTypeName(method.ResponseTypeUrl)
				// look up type matching request and response type names
				for _, t := range renderer.types {
					if m.RequestTypeName == t.Name {
						m.RequestType = t
					}
					if m.ResponseTypeName == t.Name {
						m.ResponseType = t
					}
				}
				if m.RequestType == nil {
					log.Printf("ERROR: no type %s for %s", m.RequestTypeName, m.Name)
				}
				if m.ResponseType == nil {
					log.Printf("ERROR: no type %s for %s", m.ResponseTypeName, m.Name)
				}
				renderer.methods = append(renderer.methods, &m)
			}
		}

		// scan http rules to get paths associated with service methods
		for _, rule := range service.Http.Rules {
			selector := rule.Selector
			for _, m := range renderer.methods {
				if m.Name == selector {
					m.HandlerName = "handle" + m.Name
					m.ProcessorName = "process" + m.Name
					var path string
					path = rule.GetGet()
					if path != "" {
						m.Path = path
						m.Method = "GET"
						continue
					}
					path = rule.GetPost()
					if path != "" {
						m.Path = path
						m.Method = "POST"
						continue
					}
					path = rule.GetPut()
					if path != "" {
						m.Path = path
						m.Method = "PUT"
						continue
					}
					path = rule.GetDelete()
					if path != "" {
						m.Path = path
						m.Method = "DELETE"
						continue
					}
				}
			}
		}
	*/
	return
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

func (renderer *ServiceRenderer) GenerateApp(path string) (err error) {
	err = os.MkdirAll(path, 0777)
	renderer.writeAppYaml(path + "/app.yaml")
	renderer.writeAppGo(path + "/app.go")
	renderer.writeServiceGo(path + "/service.go-finishme")
	return err
}

func (renderer *ServiceRenderer) writeAppYaml(filename string) {
	f, err := os.Create(filename)

	// file header
	err = renderer.templatedAppYaml.Execute(f, struct {
		Name string
	}{
		renderer.name,
	})
	if err != nil {
		panic(err)
	}
	f.Close()
}

// write the non-editable app.go
func (renderer *ServiceRenderer) writeAppGo(filename string) {
	f, err := os.Create(filename)

	// file header
	err = renderer.templatedFileHeader.Execute(f, struct {
		Name string
	}{
		renderer.name,
	})
	if err != nil {
		panic(err)
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
			panic(err)
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
			panic(err)
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
		panic(err)
	}

	f.Close()
	err = exec.Command("/usr/local/go/bin/gofmt", "-w", filename).Run()
}

// write the user-editable service.go
func (renderer *ServiceRenderer) writeServiceGo(filename string) {
	f, err := os.Create(filename)

	err = renderer.templatedServiceHeader.Execute(f, struct {
		Name string
	}{
		renderer.name,
	})
	if err != nil {
		panic(err)
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
			panic(err)
		}
	}

	f.Close()
	err = exec.Command("/usr/local/go/bin/gofmt", "-w", filename).Run()
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
				panic(err)
			}
			err = renderer.GenerateApp("app")
			if err != nil {
				panic(err)
			}
			log.Printf("DONE")

		})

	responseBytes, _ := proto.Marshal(response)
	os.Stdout.Write(responseBytes)
}
