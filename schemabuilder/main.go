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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

// model a section of the OpenAPI specification text document
type Section struct {
	Level    int
	Text     string
	Title    string
	Children []*Section
}

// read a section of the OpenAPI Specification, recursively dividing it into subsections
func ReadSection(text string, level int) (section *Section) {
	titlePattern := regexp.MustCompile("^" + strings.Repeat("#", level) + " .*$")
	subtitlePattern := regexp.MustCompile("^" + strings.Repeat("#", level+1) + " .*$")

	section = &Section{Level: level, Text: text}
	lines := strings.Split(string(text), "\n")
	subsection := ""
	for i, line := range lines {
		if i == 0 && titlePattern.Match([]byte(line)) {
			section.Title = line
		} else if subtitlePattern.Match([]byte(line)) {
			// we've found a subsection title.
			// if there's a subsection that we've already been reading, save it
			if len(subsection) != 0 {
				child := ReadSection(subsection, level+1)
				section.Children = append(section.Children, child)
			}
			// start a new subsection
			subsection = line + "\n"
		} else {
			// add to the subsection we've been reading
			subsection += line + "\n"
		}
	}
	// if this section has subsections, save the last one
	if len(section.Children) > 0 {
		child := ReadSection(subsection, level+1)
		section.Children = append(section.Children, child)
	}
	return
}

// recursively display a section of the specification
func (s *Section) Display(section string) {
	if len(s.Children) == 0 {
		//fmt.Printf("%s\n", s.Text)
	} else {
		for i, child := range s.Children {
			var subsection string
			if section == "" {
				subsection = fmt.Sprintf("%d", i)
			} else {
				subsection = fmt.Sprintf("%s.%d", section, i)
			}
			fmt.Printf("%-12s %s\n", subsection, child.NiceTitle())
			child.Display(subsection)
		}
	}
}

// remove a link from a string, leaving only the text that follows it
// if there is no link, just return the string
func stripLink(input string) (output string) {
	stringPattern := regexp.MustCompile("^(.*)$")
	stringWithLinkPattern := regexp.MustCompile("^<a .*</a>(.*)$")
	if matches := stringWithLinkPattern.FindSubmatch([]byte(input)); matches != nil {
		return string(matches[1])
	} else if matches := stringPattern.FindSubmatch([]byte(input)); matches != nil {
		return string(matches[1])
	} else {
		return input
	}
}

// return a nice-to-display title for a section by removing the opening "###" and any links
func (s *Section) NiceTitle() string {
	titlePattern := regexp.MustCompile("^#+ (.*)$")
	titleWithLinkPattern := regexp.MustCompile("^#+ <a .*</a>(.*)$")
	if matches := titleWithLinkPattern.FindSubmatch([]byte(s.Title)); matches != nil {
		return string(matches[1])
	} else if matches := titlePattern.FindSubmatch([]byte(s.Title)); matches != nil {
		return string(matches[1])
	} else {
		return ""
	}
}

// replace markdown links with their link text (removing the URL part)
func removeMarkdownLinks(input string) (output string) {
	markdownLink := regexp.MustCompile("\\[([^\\]]*)\\]\\(([^\\)]*)\\)") // matches [link title](link url)
	output = string(markdownLink.ReplaceAll([]byte(input), []byte("$1")))
	return
}

// extract the fixed fields from a table in a section
func parseFixedFields(input string, schemaObject *SchemaObject) {
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) > 1 {
			fieldName := strings.Trim(stripLink(parts[0]), " ")
			if fieldName != "Field Name" && fieldName != "---" {
				typeName := parts[1]
				typeName = strings.Trim(typeName, " ")
				typeName = strings.Replace(typeName, "`", "", -1)
				typeName = strings.Replace(typeName, " <span>&#124;</span> ", "|", -1)
				typeName = removeMarkdownLinks(typeName)
				typeName = strings.Replace(typeName, " ", "", -1)
				typeName = strings.Replace(typeName, "Object", "", -1)
				isArray := false
				if typeName[0] == '[' && typeName[len(typeName)-1] == ']' {
					typeName = typeName[1 : len(typeName)-1]
					isArray = true
				}
				isMap := false
				mapPattern := regexp.MustCompile("^Mapstring,\\[(.*)\\]$")
				if matches := mapPattern.FindSubmatch([]byte(typeName)); matches != nil {
					typeName = string(matches[1])
					isMap = true
				}
				description := strings.Trim(parts[len(parts)-1], " ")
				description = removeMarkdownLinks(description)
				requiredLabel := "**Required.** "
				if strings.Contains(description, requiredLabel) {
					schemaObject.RequiredFields = append(schemaObject.RequiredFields, fieldName)
					description = strings.Replace(description, requiredLabel, "", -1)
				}
				schemaField := SchemaObjectField{
					Name:        fieldName,
					Type:        typeName,
					IsArray:     isArray,
					IsMap:       isMap,
					Description: description,
				}
				schemaObject.FixedFields = append(schemaObject.FixedFields, schemaField)
			}
		}
	}
}

// extract the patterned fields from a table in a section
func parsePatternedFields(input string, schemaObject *SchemaObject) {
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) > 1 {
			fieldName := strings.Trim(stripLink(parts[0]), " ")
			fieldName = removeMarkdownLinks(fieldName)
			if fieldName != "Field Pattern" && fieldName != "---" {
				typeName := parts[1]
				typeName = strings.Trim(typeName, " ")
				typeName = strings.Replace(typeName, "`", "", -1)
				typeName = strings.Replace(typeName, " <span>&#124;</span> ", "|", -1)
				typeName = removeMarkdownLinks(typeName)
				typeName = strings.Replace(typeName, " ", "", -1)
				typeName = strings.Replace(typeName, "Object", "", -1)
				isArray := false
				if typeName[0] == '[' && typeName[len(typeName)-1] == ']' {
					typeName = typeName[1 : len(typeName)-1]
					isArray = true
				}
				isMap := false
				mapPattern := regexp.MustCompile("^Mapstring,\\[(.*)\\]$")
				if matches := mapPattern.FindSubmatch([]byte(typeName)); matches != nil {
					typeName = string(matches[1])
					isMap = true
				}
				description := strings.Trim(parts[len(parts)-1], " ")
				description = removeMarkdownLinks(description)
				schemaField := SchemaObjectField{
					Name:        fieldName,
					Type:        typeName,
					IsArray:     isArray,
					IsMap:       isMap,
					Description: description,
				}
				schemaObject.PatternedFields = append(schemaObject.PatternedFields, schemaField)
			}
		}
	}
}

type SchemaObjectField struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	IsArray     bool   `json:"is_array"`
	IsMap       bool   `json:"is_map"`
	Description string `json:"description"`
}

type SchemaObject struct {
	Name            string              `json:"name"`
	Id              string              `json:"id"`
	Description     string              `json:"description"`
	Extendable      bool                `json:"extendable"`
	RequiredFields  []string            `json:"required"`
	FixedFields     []SchemaObjectField `json:"fixed"`
	PatternedFields []SchemaObjectField `json:"patterned"`
}

type SchemaModel struct {
	Objects []SchemaObject
}

func NewSchemaModel(filename string) (schemaModel *SchemaModel, err error) {

	b, err := ioutil.ReadFile("3.0.md")
	if err != nil {
		return nil, err
	}

	// divide the specification into sections
	document := ReadSection(string(b), 1)
	document.Display("")

	// read object names and their details
	specification := document.Children[4] // fragile!
	schema := specification.Children[5]   // fragile!
	anchor := regexp.MustCompile("^#### <a name=\"(.*)Object\"")
	schemaObjects := make([]SchemaObject, 0)
	for _, section := range schema.Children {
		if matches := anchor.FindSubmatch([]byte(section.Title)); matches != nil {

			id := string(matches[1])

			schemaObject := SchemaObject{
				Name:           section.NiceTitle(),
				Id:             id,
				RequiredFields: make([]string, 0),
			}

			if len(section.Children) > 0 {
				details := section.Children[0].Text
				details = removeMarkdownLinks(details)
				details = strings.Trim(details, " \t\n")
				schemaObject.Description = details
			}

			// is the object extendable?
			if strings.Contains(section.Text, "Specification Extensions") {
				schemaObject.Extendable = true
			}

			// look for fixed fields
			for _, child := range section.Children {
				if child.NiceTitle() == "Fixed Fields" {
					parseFixedFields(child.Text, &schemaObject)
				}
			}

			// look for patterned fields
			for _, child := range section.Children {
				if child.NiceTitle() == "Patterned Fields" {
					parsePatternedFields(child.Text, &schemaObject)
				}
			}

			schemaObjects = append(schemaObjects, schemaObject)
		}
	}

	return &SchemaModel{Objects: schemaObjects}, nil
}

func main() {
	// read and parse the text specification into a structure
	model, err := NewSchemaModel("3.0.md")
	if err != nil {
		panic(err)
	}

	modelJSON, _ := json.MarshalIndent(model, "", "  ")
	fmt.Printf("%s\n", string(modelJSON))
	err = ioutil.WriteFile("model.json", modelJSON, 0644)
	if err != nil {
		panic(err)
	}
}
