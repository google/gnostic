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
	"regexp"
	"strings"
)

type Section struct {
	Level    int
	Text     string
	Title    string
	Children []*Section
}

// read a section of the OpenAPI Specification, dividing it into subsections
func ReadSection(text string, level int) (section *Section, err error) {
	titlePattern := regexp.MustCompile("^" + strings.Repeat("#", level) + " .*$")
	subtitlePattern := regexp.MustCompile("^" + strings.Repeat("#", level+1) + " .*$")

	section = &Section{Level: level}
	lines := strings.Split(string(text), "\n")
	subsection := ""
	for i, line := range lines {
		if i == 0 && titlePattern.Match([]byte(line)) {
			section.Title = line
		} else if subtitlePattern.Match([]byte(line)) {
			if len(subsection) != 0 {
				child, _ := ReadSection(subsection, level+1)
				section.Children = append(section.Children, child)
			}
			subsection = line + "\n"
		} else {
			subsection += line + "\n"
		}
	}
	if len(section.Children) > 0 {
		child, _ := ReadSection(subsection, level+1)
		section.Children = append(section.Children, child)
	}
	// save the text
	section.Text = text
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

// return a nice-to-display title for a section that removes the opening "###" and any links
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

// replace markdown links with their link text (removing the URLs)
func removeMarkdownLinks(input string) (output string) {
	markdownLink := regexp.MustCompile("\\[([^\\]]*)\\]\\(([^\\)]*)\\)") // matches [link title](link url)
	output = string(markdownLink.ReplaceAll([]byte(input), []byte("$1")))
	return
}

// extract the fixed fields from a table in a section
func parseFixedFields(input string) {
	fmt.Printf("- FIXED FIELDS:\n")
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
				fmt.Printf("%-20s %s\n", fieldName, typeName)
			}
		}
	}
}

// extract the patterned fields from a table in a section
func parsePatternedFields(input string) {
	fmt.Printf("- PATTERNED FIELDS:\n")
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
				fmt.Printf("%-20s %s\n", fieldName, typeName)
			}
		}
	}
}

func main() {
	b, err := ioutil.ReadFile("3.0.md")
	if err != nil {
		panic(err)
	}

	// divide the specification into sections
	document, err := ReadSection(string(b), 1)
	document.Display("")

	// read object names and their details
	specification := document.Children[4]
	schema := specification.Children[5]
	anchor := regexp.MustCompile("^#### <a name=\"(.*)Object\"")
	for _, section := range schema.Children {
		if matches := anchor.FindSubmatch([]byte(section.Title)); matches != nil {
			//name := string(matches[1])
			fmt.Printf("\n%s\n", section.NiceTitle())

			if len(section.Children) > 0 {
				details := section.Children[0].Text
				details = removeMarkdownLinks(details)
				details = strings.Trim(details, " \t\n")
				fmt.Printf("- DESCRIPTION: {%s}\n", details)
			}

			// is the object extendable?
			if strings.Contains(section.Text, "Specification Extensions") {
				fmt.Printf("- EXTENDABLE\n")
			}

			// look for fixed fields
			for _, child := range section.Children {
				if child.NiceTitle() == "Fixed Fields" {
					parseFixedFields(child.Text)
				}
			}

			// look for patterned fields
			for _, child := range section.Children {
				if child.NiceTitle() == "Patterned Fields" {
					parsePatternedFields(child.Text)
				}
			}
		}
	}
}
