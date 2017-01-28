package main

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"
	"strings"
	"text/template"
)

const MAIN_GO = `package main

func templates() map[string]string {
	return map[string]string{ {{range .TemplateStrings}}
        "{{.Name}}": "{{.Encoding}}",{{end}}
    }
}`

type NamedTemplateString struct {
	Name     string
	Encoding string
}

func main() {
	templateFiles, err := ioutil.ReadDir("templates")
	if err != nil {
		panic(err)
	}
	templateStrings := make([]*NamedTemplateString, 0)
	for _, templateFile := range templateFiles {
		name := templateFile.Name()
		data, _ := ioutil.ReadFile("templates/" + name)
		encoding := base64.StdEncoding.EncodeToString(data)
		templateStrings = append(templateStrings,
			&NamedTemplateString{
				Name:     strings.TrimSuffix(name, ".tmpl"),
				Encoding: encoding})
	}
	t, err := template.New("main.go").Parse(MAIN_GO)
	f := new(bytes.Buffer)
	// file header
	err = t.Execute(f, struct {
		TemplateStrings []*NamedTemplateString
	}{
		templateStrings,
	})

	ioutil.WriteFile("templates.go", f.Bytes(), 0644)
}
