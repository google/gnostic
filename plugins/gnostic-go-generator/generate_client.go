package main

func (renderer *ServiceRenderer) GenerateClient() ([]byte, error) {
	f := NewLineWriter()

	f.WriteLine("// GENERATED FILE: DO NOT EDIT!\n")
	f.WriteLine("package " + renderer.Model.Package)
	imports := []string{
		"bytes",
		"encoding/json",
		"errors",
		"fmt",
		"io/ioutil",
		"log",
		"net/http",
		"net/url",
		"strings",
	}
	f.WriteLine(`import (`)
	for _, imp := range imports {
		f.WriteLine(`"` + imp + `"`)
	}
	f.WriteLine(`)`)

	f.WriteLine(`// API client representation.`)
	f.WriteLine(`type Client struct {`)
	f.WriteLine(` service string`)
	f.WriteLine(`	APIKey string`)
	f.WriteLine(`}`)

	f.WriteLine(`// Create an API client.`)
	f.WriteLine(`func NewClient(service string) *Client {`)
	f.WriteLine(`	client := &Client{}`)
	f.WriteLine(`	client.service = service`)
	f.WriteLine(`	return client`)
	f.WriteLine(`}`)

	for _, method := range renderer.Model.Methods {
		f.WriteLine(commentForText(method.Description) + "\n")
		funcline := `func (client *Client) ` + method.ClientName + `(` + parameterList(method) + `) `
		if method.ResponsesType == nil {
			funcline += ` (err error)`
		} else {
			funcline += ` (response *` + method.ResultTypeName + `, err error)`
		}
		funcline += ` {`
		f.WriteLine(funcline)

		f.WriteLine(`path := client.service + "` + method.Path + `"`)

		if hasPathParameters(method) {
			for _, field := range method.ParametersType.Fields {
				if field.Position == "path" {
					f.WriteLine(`path = strings.Replace(path, "{` + field.Name + `}", fmt.Sprintf("%v", ` +
						field.ParameterName + `), 1)`)
				}
			}
		}

		if hasQueryParameters(method) {
			f.WriteLine(`v := url.Values{}`)
			for _, field := range method.ParametersType.Fields {
				if field.Position == "query" {
					f.WriteLine(`v.Set("` + field.Name + `", ` + field.ParameterName + `)`)
				}
			}
			f.WriteLine(`if client.APIKey != "" {`)
			f.WriteLine(`  v.Set("key", client.APIKey)`)
			f.WriteLine(`}`)
			f.WriteLine(`path = path + "?" + v.Encode()`)
		}

		if method.Method == "POST" {
			f.WriteLine(`body := new(bytes.Buffer)`)
			f.WriteLine(`json.NewEncoder(body).Encode(` + bodyParameterName(method) + `)`)
			f.WriteLine(`req, err := http.NewRequest("` + method.Method + `", path, body)`)
		} else {
			f.WriteLine(`req, err := http.NewRequest("` + method.Method + `", path, nil)`)
		}
		f.WriteLine(`if err != nil {return}`)
		f.WriteLine(`resp, err := http.DefaultClient.Do(req)`)
		f.WriteLine(`if err != nil {return}`)
		f.WriteLine(`if false {`)
		f.WriteLine(`log.Printf("%+v", resp)`)
		f.WriteLine(`}`)

		if method.ResponsesType != nil {
			f.WriteLine(`response = &` + method.ResultTypeName + `{}`)

			f.WriteLine(`switch {`)
			// first handle everything that isn't "default"
			for _, responseField := range method.ResponsesType.Fields {
				if responseField.Name != "default" {
					f.WriteLine(`case resp.StatusCode == ` + responseField.Name + `:`)
					f.WriteLine(`  defer resp.Body.Close()`)
					f.WriteLine(`  body, err := ioutil.ReadAll(resp.Body)`)
					f.WriteLine(`  if err != nil {return nil, err}`)
					f.WriteLine(`  result := &` + responseField.ValueType + `{}`)
					f.WriteLine(`  err = json.Unmarshal(body, result)`)
					f.WriteLine(`  if err != nil {return nil, err}`)
					f.WriteLine(`  response.` + responseField.FieldName + ` = result`)
				}
			}

			// then handle "default"
			hasDefault := false
			for _, responseField := range method.ResponsesType.Fields {
				if responseField.Name == "default" {
					hasDefault = true
					f.WriteLine(`default:`)
					f.WriteLine(`  defer resp.Body.Close()`)
					f.WriteLine(`  body, err := ioutil.ReadAll(resp.Body)`)
					f.WriteLine(`  if err != nil {return nil, err}`)
					f.WriteLine(`  result := &` + responseField.ValueType + `{}`)
					f.WriteLine(`  err = json.Unmarshal(body, result)`)
					f.WriteLine(`  if err != nil {return nil, err}`)
					f.WriteLine(`  response.` + responseField.FieldName + ` = result`)
				}
			}
			if !hasDefault {
				f.WriteLine(`default:`)
				f.WriteLine(`  break`)
			}
			f.WriteLine(`}`) // close switch statement
		}
		f.WriteLine("return")
		f.WriteLine("}")
	}
	f.WriteLine(`
// refer to imported packages that may or may not be used in generated code
func forced_package_references() {
	_ = new(bytes.Buffer)
	_ = fmt.Sprintf("")
	_ = strings.Split("","")
	_ = url.Values{}
	_ = errors.New("")
	ioutil.ReadFile("")
}
`)
	return f.Bytes(), nil
}
