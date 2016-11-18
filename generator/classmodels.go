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

package main

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
)

/// Class Modeling

// models classes that we encounter during traversal that have no named schema
type ClassRequest struct {
	Name         string
	PropertyName string // name of a property that refers to this class
	Schema       *Schema
}

func NewClassRequest(name string, propertyName string, schema *Schema) *ClassRequest {
	return &ClassRequest{Name: name, PropertyName: propertyName, Schema: schema}
}

// models class properties, eg. fields
type ClassProperty struct {
	Name     string
	Type     string
	Repeated bool
	Pattern  string
}

func (classProperty *ClassProperty) display() string {
	if classProperty.Repeated {
		return fmt.Sprintf("\t%s %s repeated\n", classProperty.Name, classProperty.Type)
	} else {
		return fmt.Sprintf("\t%s %s\n", classProperty.Name, classProperty.Type)
	}
}

func NewClassProperty() *ClassProperty {
	return &ClassProperty{}
}

func NewClassPropertyWithNameAndType(name string, typeName string) *ClassProperty {
	return &ClassProperty{Name: name, Type: typeName}
}

func NewClassPropertyWithNameTypeAndPattern(name string, typeName string, pattern string) *ClassProperty {
	return &ClassProperty{Name: name, Type: typeName, Pattern: pattern}
}

// models classes
type ClassModel struct {
	Name       string
	Properties map[string]*ClassProperty
	Required   []string
}

func (classModel *ClassModel) sortedPropertyNames() []string {
	keys := make([]string, 0)
	for k, _ := range classModel.Properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (classModel *ClassModel) display() string {
	result := fmt.Sprintf("%+s\n", classModel.Name)
	keys := classModel.sortedPropertyNames()
	for _, k := range keys {
		result += classModel.Properties[k].display()
	}
	return result
}

func NewClassModel() *ClassModel {
	cm := &ClassModel{}
	cm.Properties = make(map[string]*ClassProperty, 0)
	return cm
}

// models a collection of classes that is defined by a schema

type ClassCollection struct {
	ClassModels         map[string]*ClassModel
	Prefix              string
	Schema              *Schema
	PatternNames        map[string]string
	ClassNames          []string
	ObjectClassRequests map[string]*ClassRequest
}

func NewClassCollection(schema *Schema) *ClassCollection {
	cc := &ClassCollection{}
	cc.ClassModels = make(map[string]*ClassModel, 0)
	cc.PatternNames = make(map[string]string, 0)
	cc.ClassNames = make([]string, 0)
	cc.ObjectClassRequests = make(map[string]*ClassRequest, 0)
	cc.Schema = schema
	return cc
}

func (classes *ClassCollection) classNameForStub(stub string) string {
	return classes.Prefix + strings.ToUpper(stub[0:1]) + stub[1:len(stub)]
}

func (classes *ClassCollection) classNameForReference(reference string) string {
	parts := strings.Split(reference, "/")
	first := parts[0]
	last := parts[len(parts)-1]
	if first == "#" {
		return classes.classNameForStub(last)
	} else {
		panic("no class name")
		return ""
	}
}

func (classes *ClassCollection) propertyNameForReference(reference string) *string {
	parts := strings.Split(reference, "/")
	first := parts[0]
	last := parts[len(parts)-1]
	if first == "#" {
		return &last
	} else {
		return nil
	}
	return nil
}

func (classes *ClassCollection) arrayTypeForSchema(schema *Schema) string {
	// what is the array type?
	log.Printf("Array type for\n%s", schema.display())

	itemTypeName := "Any"
	if schema.Items != nil {

		if schema.Items.Array != nil {

			if len(*(schema.Items.Array)) > 0 {
				ref := (*schema.Items.Array)[0].Ref
				if ref != nil {
					itemTypeName = classes.classNameForReference(*ref)
				} else {
					types := (*schema.Items.Array)[0].Type
					if types == nil {
						// do nothing
					} else if (types.Array != nil) && len(*(types.Array)) == 1 {
						itemTypeName = (*types.Array)[0]
					} else if (types.Array != nil) && len(*(types.Array)) > 1 {
						itemTypeName = fmt.Sprintf("%+v", types.Array)
					} else if types.String != nil {
						itemTypeName = *(types.String)
					} else {
						itemTypeName = "UNKNOWN"
					}
				}
			}

		} else if schema.Items.Schema != nil {

			var ref *string
			ref = schema.Items.Schema.Ref
			if ref != nil {
				itemTypeName = classes.classNameForReference(*ref)
			} else {
				types := schema.Items.Schema.Type
				if types == nil {
					// do nothing
				} else if (types.Array != nil) && len(*(types.Array)) == 1 {
					itemTypeName = (*types.Array)[0]
				} else if (types.Array != nil) && len(*(types.Array)) > 1 {
					itemTypeName = fmt.Sprintf("%+v", types.Array)
				} else if types.String != nil {
					itemTypeName = *(types.String)
				} else {
					itemTypeName = "UNKNOWN"
				}
			}
		}

	}
	log.Printf("is %s\n", itemTypeName)
	return itemTypeName
}

func (classes *ClassCollection) buildClassProperties(classModel *ClassModel, schema *Schema) {
	if schema.Properties != nil {
		for key, value := range *(schema.Properties) {
			if value.Ref != nil {
				className := classes.classNameForReference(*(value.Ref))
				cp := NewClassProperty()
				cp.Name = key
				cp.Type = className
				classModel.Properties[key] = cp
			} else {
				if value.Type != nil {
					if value.typeIs("string") {
						classModel.Properties[key] = NewClassPropertyWithNameAndType(key, "string")
					} else if value.typeIs("boolean") {
						classModel.Properties[key] = NewClassPropertyWithNameAndType(key, "bool")
					} else if value.typeIs("number") {
						classModel.Properties[key] = NewClassPropertyWithNameAndType(key, "float")
					} else if value.typeIs("integer") {
						classModel.Properties[key] = NewClassPropertyWithNameAndType(key, "int")
					} else if value.typeIs("object") {
						className := classes.classNameForStub(key)
						classes.ObjectClassRequests[className] = NewClassRequest(className, key, value)
						classModel.Properties[key] = NewClassPropertyWithNameAndType(key, className)
					} else if value.typeIs("array") {
						log.Printf("ARRAY PROPETY %s", key)
						className := classes.arrayTypeForSchema(value)
						p := NewClassPropertyWithNameAndType(key, className)
						p.Repeated = true
						classModel.Properties[key] = p
					} else {
						log.Printf("%+v has unsupported property type %+v", key, value.Type)
					}
				} else {
					if value.isEmpty() {
						// write accessor for generic object
						className := "Any"
						classModel.Properties[key] = NewClassPropertyWithNameAndType(key, className)
					} else if value.AnyOf != nil {
						//self.writeAnyOfAccessors(schema: value, path: path, accessorName:accessorName)
					} else if value.OneOf != nil {
						//self.writeOneOfAccessors(schema: value, path: path)
					} else {
						//print("\(path):\(key) has unspecified property type. Schema is below.\n\(value.description)")
					}
				}
			}
		}
	}
}

func (classes *ClassCollection) buildClassRequirements(classModel *ClassModel, schema *Schema) {
	if schema.Required != nil {
		classModel.Required = (*schema.Required)
	}
}

func (classes *ClassCollection) buildPatternPropertyAccessors(classModel *ClassModel, schema *Schema) {
	if schema.PatternProperties != nil {
		for key, propertySchema := range *(schema.PatternProperties) {
			className := "Any"
			propertyName := classes.PatternNames[key]
			if propertySchema.Ref != nil {
				className = classes.classNameForReference(*propertySchema.Ref)
			}
			typeName := fmt.Sprintf("map<string, %s>", className)
			classModel.Properties[propertyName] = NewClassPropertyWithNameTypeAndPattern(propertyName, typeName, key)
		}
	}
}

func (classes *ClassCollection) buildAdditionalPropertyAccessors(classModel *ClassModel, schema *Schema) {

	if schema.AdditionalProperties != nil {
		if schema.AdditionalProperties.Boolean != nil {
			if *schema.AdditionalProperties.Boolean == true {
				propertyName := "additionalProperties"
				className := "map<string, Any>"
				classModel.Properties[propertyName] = NewClassPropertyWithNameAndType(propertyName, className)
				return
			}
		} else if schema.AdditionalProperties.Schema != nil {
			schema := schema.AdditionalProperties.Schema
			if schema.Ref != nil {
				propertyName := "additionalProperties"
				className := fmt.Sprintf("map<string, %s>", classes.classNameForReference(*schema.Ref))
				classModel.Properties[propertyName] = NewClassPropertyWithNameAndType(propertyName, className)
				return
			} else if schema.Type != nil {

				typeName := *schema.Type.String
				if typeName == "string" {
					propertyName := "additionalProperties"
					className := "map<string, string>"
					classModel.Properties[propertyName] = NewClassPropertyWithNameAndType(propertyName, className)
					return
				} else if typeName == "array" {
					if schema.Items != nil {
						itemType := *schema.Items.Schema.Type.String
						if itemType == "string" {
							propertyName := "additionalProperties"
							className := "map<string, StringArray>"
							classModel.Properties[propertyName] = NewClassPropertyWithNameAndType(propertyName, className)
							return
						}
					}
				}
			} else if schema.OneOf != nil {
				classes.buildOneOfAccessorsHelper(classModel, schema.OneOf)
			}
		}
	}
}

func (classes *ClassCollection) buildOneOfAccessors(classModel *ClassModel, schema *Schema) {
	if schema.OneOf != nil {
		classes.buildOneOfAccessorsHelper(classModel, schema.OneOf)
	}
}

func (classes *ClassCollection) buildOneOfAccessorsHelper(classModel *ClassModel, oneOfs *[]*Schema) {
	log.Printf("buildOneOfAccessorsHelper(%+v, %+v)", classModel, oneOfs)
	for _, oneOf := range *oneOfs {
		if oneOf.Ref != nil {
			ref := *oneOf.Ref
			className := classes.classNameForReference(ref)
			propertyName := classes.propertyNameForReference(ref)

			if propertyName != nil {
				log.Printf("property %s class %s", *propertyName, className)
				classModel.Properties[*propertyName] = NewClassPropertyWithNameAndType(*propertyName, className)
			}
		}
	}
}

func (classes *ClassCollection) buildDefaultAccessors(classModel *ClassModel, schema *Schema) {
	key := "additionalProperties"
	className := "map<string, Any>"
	classModel.Properties[key] = NewClassPropertyWithNameAndType(key, className)
}

func (classes *ClassCollection) buildClassForDefinition(
	className string,
	propertyName string,
	schema *Schema) *ClassModel {
	if schema.Type == nil {
		return classes.buildClassForDefinitionObject(className, propertyName, schema)
	}
	typeString := *schema.Type.String
	if typeString == "object" {
		return classes.buildClassForDefinitionObject(className, propertyName, schema)
	} else {
		return nil
	}
}

func (classes *ClassCollection) buildClassForDefinitionObject(
	className string,
	propertyName string,
	schema *Schema) *ClassModel {
	classModel := NewClassModel()
	classModel.Name = className
	if schema.isEmpty() {
		classes.buildDefaultAccessors(classModel, schema)
	} else {
		classes.buildClassProperties(classModel, schema)
		classes.buildClassRequirements(classModel, schema)
		classes.buildPatternPropertyAccessors(classModel, schema)
		classes.buildAdditionalPropertyAccessors(classModel, schema)
		classes.buildOneOfAccessors(classModel, schema)
	}
	return classModel
}

func (classes *ClassCollection) build() {
	// create a class for the top-level schema
	className := classes.Prefix + "Document"
	classModel := NewClassModel()
	classModel.Name = className
	classes.buildClassProperties(classModel, classes.Schema)
	classes.buildClassRequirements(classModel, classes.Schema)

	classes.ClassModels[className] = classModel

	// create a class for each object defined in the schema
	for key, value := range *(classes.Schema.Definitions) {
		className := classes.classNameForStub(key)
		model := classes.buildClassForDefinition(className, key, value)
		if model != nil {
			classes.ClassModels[className] = model
		}
	}

	// iterate over anonymous object classes to be instantiated and generate a class for each
	for className, classRequest := range classes.ObjectClassRequests {
		classes.ClassModels[classRequest.Name] =
			classes.buildClassForDefinitionObject(className, classRequest.PropertyName, classRequest.Schema)
	}

	// add a class for string arrays
	stringArrayClass := NewClassModel()
	stringArrayClass.Name = "StringArray"
	stringProperty := NewClassProperty()
	stringProperty.Name = "value"
	stringProperty.Type = "string"
	stringProperty.Repeated = true
	stringArrayClass.Properties[stringProperty.Name] = stringProperty
	classes.ClassModels[stringArrayClass.Name] = stringArrayClass

	// add a class for "Any"
	anyClass := NewClassModel()
	anyClass.Name = "Any"
	valueProperty := NewClassProperty()
	valueProperty.Name = "value"
	valueProperty.Type = "google.protobuf.Any"
	valueProperty.Repeated = true
	anyClass.Properties[valueProperty.Name] = valueProperty
	classes.ClassModels[anyClass.Name] = anyClass
}

func (classes *ClassCollection) sortedClassNames() []string {
	keys := make([]string, 0)
	for k, _ := range classes.ClassModels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (classes *ClassCollection) display() string {
	keys := classes.sortedClassNames()
	result := ""
	for _, k := range keys {
		result += classes.ClassModels[k].display()
	}
	return result
}

func camelCaseToSnakeCase(input string) string {
	var out = ""

	for index, runeValue := range input {
		//fmt.Printf("%#U starts at byte position %d\n", runeValue, index)
		if runeValue >= 'A' && runeValue <= 'Z' {
			if index > 0 {
				out += "_"
			}
			out += string(runeValue - 'A' + 'a')
		} else {
			out += string(runeValue)
		}

	}
	return out
}

func mapTypeInfo(typeName string) (isMap bool, valueTypeName string) {
	r, err := regexp.Compile("^map<string, (.*)>$")
	if err != nil {
		panic(err)
	}
	match := r.FindStringSubmatch(typeName)
	if len(match) != 2 {
		isMap = false
	} else {
		isMap = true
		valueTypeName = match[1]
	}
	return
}
