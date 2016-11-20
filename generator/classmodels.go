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
	"sort"
	"strings"
)

/// Class Modeling

// models classes that we encounter during traversal that have no named schema
type ClassRequest struct {
	Name         string
	PropertyName string // name of a property that refers to this class
	Schema       *Schema
	OneOfWrapper bool
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
		return fmt.Sprintf("\t%s %s repeated %s\n", classProperty.Name, classProperty.Type, classProperty.Pattern)
	} else {
		return fmt.Sprintf("\t%s %s %s \n", classProperty.Name, classProperty.Type, classProperty.Pattern)
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
	Name         string
	Properties   map[string]*ClassProperty
	Required     []string
	OneOfWrapper bool
	Open         bool // open classes can have keys outside the specified set (pattern properties, etc)
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

// Returns a capitalized name to use for a generated class
func (classes *ClassCollection) classNameForStub(stub string) string {
	return classes.Prefix + strings.ToUpper(stub[0:1]) + stub[1:len(stub)]
}

// Returns a capitalized name to use for a generated class based on a JSON reference
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

// Returns a property name to use for a JSON reference
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

// Determines the item type for arrays defined by a schema
func (classes *ClassCollection) arrayItemTypeForSchema(propertyName string, schema *Schema) string {
	// default
	itemTypeName := "Any"

	if schema.Items != nil {

		if schema.Items.SchemaArray != nil {

			if len(*(schema.Items.SchemaArray)) > 0 {
				ref := (*schema.Items.SchemaArray)[0].Ref
				if ref != nil {
					itemTypeName = classes.classNameForReference(*ref)
				} else {
					types := (*schema.Items.SchemaArray)[0].Type
					if types == nil {
						// do nothing
					} else if (types.StringArray != nil) && len(*(types.StringArray)) == 1 {
						itemTypeName = (*types.StringArray)[0]
					} else if (types.StringArray != nil) && len(*(types.StringArray)) > 1 {
						itemTypeName = fmt.Sprintf("%+v", types.StringArray)
					} else if types.String != nil {
						itemTypeName = *(types.String)
					} else {
						itemTypeName = "UNKNOWN"
					}
				}
			}

		} else if schema.Items.Schema != nil {
			types := schema.Items.Schema.Type

			if schema.Items.Schema.Ref != nil {
				itemTypeName = classes.classNameForReference(*schema.Items.Schema.Ref)
			} else if schema.Items.Schema.OneOf != nil {
				// this type is implied by the "oneOf"
				itemTypeName = classes.classNameForStub(propertyName + "Item")
				classes.ObjectClassRequests[itemTypeName] =
					NewClassRequest(itemTypeName, propertyName, schema.Items.Schema)
			} else if types == nil {
				// do nothing
			} else if (types.StringArray != nil) && len(*(types.StringArray)) == 1 {
				itemTypeName = (*types.StringArray)[0]
			} else if (types.StringArray != nil) && len(*(types.StringArray)) > 1 {
				itemTypeName = fmt.Sprintf("%+v", types.StringArray)
			} else if types.String != nil {
				itemTypeName = *(types.String)
			} else {
				itemTypeName = "UNKNOWN"
			}
		}

	}
	log.Printf("is %s\n", itemTypeName)
	return itemTypeName
}

func (classes *ClassCollection) buildClassProperties(classModel *ClassModel, schema *Schema) {
	if schema.Properties != nil {
		for propertyName, propertySchema := range *(schema.Properties) {
			if propertySchema.Ref != nil {
				// the property schema is a reference, so we will add a property with the type of the referenced schema
				propertyClassName := classes.classNameForReference(*(propertySchema.Ref))
				classProperty := NewClassProperty()
				classProperty.Name = propertyName
				classProperty.Type = propertyClassName
				classModel.Properties[propertyName] = classProperty
			} else if propertySchema.Type != nil {
				// the property schema specifies a type, so add a property with the specified type
				if propertySchema.typeIs("string") {
					classModel.Properties[propertyName] = NewClassPropertyWithNameAndType(propertyName, "string")
				} else if propertySchema.typeIs("boolean") {
					classModel.Properties[propertyName] = NewClassPropertyWithNameAndType(propertyName, "bool")
				} else if propertySchema.typeIs("number") {
					classModel.Properties[propertyName] = NewClassPropertyWithNameAndType(propertyName, "float")
				} else if propertySchema.typeIs("integer") {
					classModel.Properties[propertyName] = NewClassPropertyWithNameAndType(propertyName, "int")
				} else if propertySchema.typeIs("object") {
					// the property has an "anonymous" object schema, so define a new class for it and request its creation
					anonymousObjectClassName := classes.classNameForStub(propertyName)
					classes.ObjectClassRequests[anonymousObjectClassName] =
						NewClassRequest(anonymousObjectClassName, propertyName, propertySchema)
					// add a property with the type of the requested class
					classModel.Properties[propertyName] = NewClassPropertyWithNameAndType(propertyName, anonymousObjectClassName)
				} else if propertySchema.typeIs("array") {
					// the property has an array type, so define it as a a repeated property of the specified type
					propertyClassName := classes.arrayItemTypeForSchema(propertyName, propertySchema)
					classProperty := NewClassPropertyWithNameAndType(propertyName, propertyClassName)
					classProperty.Repeated = true
					classModel.Properties[propertyName] = classProperty
				} else {
					log.Printf("ignoring %+v, which has an unsupported property type '%+v'", propertyName, propertySchema.Type)
				}
			} else if propertySchema.isEmpty() {
				// an empty schema can contain anything, so add an accessor for a generic object
				className := "Any"
				classModel.Properties[propertyName] = NewClassPropertyWithNameAndType(propertyName, className)
			} else if propertySchema.OneOf != nil {
				anonymousObjectClassName := classes.classNameForStub(propertyName + "Item")
				classes.ObjectClassRequests[anonymousObjectClassName] =
					NewClassRequest(anonymousObjectClassName, propertyName, propertySchema)
				classModel.Properties[propertyName] = NewClassPropertyWithNameAndType(propertyName, anonymousObjectClassName)
			} else {
				log.Printf("ignoring %s.%s, which has an unrecognized schema:\n%+v", classModel.Name, propertyName, propertySchema.display())
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
		classModel.Open = true
		for propertyPattern, propertySchema := range *(schema.PatternProperties) {
			log.Printf("BUILDING %+v\n%+v", propertyPattern, propertySchema.display())
			className := "Any"
			propertyName := classes.PatternNames[propertyPattern]
			if propertySchema.Ref != nil {
				className = classes.classNameForReference(*propertySchema.Ref)
			}
			propertyTypeName := fmt.Sprintf("map<string, %s>", className)
			classModel.Properties[propertyName] = NewClassPropertyWithNameTypeAndPattern(propertyName, propertyTypeName, propertyPattern)
		}
	}
}

func (classes *ClassCollection) buildAdditionalPropertyAccessors(classModel *ClassModel, schema *Schema) {
	if schema.AdditionalProperties != nil {
		classModel.Open = true
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
				classes.buildOneOfAccessors(classModel, schema)
			}
		}
	}
}

func (classes *ClassCollection) buildOneOfAccessors(classModel *ClassModel, schema *Schema) {
	oneOfs := schema.OneOf
	if oneOfs == nil {
		return
	}
	log.Printf("buildOneOfAccessors(%+v, %+v)", classModel, oneOfs)
	classModel.OneOfWrapper = true
	for _, oneOf := range *oneOfs {
		log.Printf("%+v", oneOf.display())
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
	classModel.Open = true
	propertyName := "additionalProperties"
	className := "map<string, Any>"
	classModel.Properties[propertyName] = NewClassPropertyWithNameAndType(propertyName, className)
}

func (classes *ClassCollection) buildClassForDefinition(
	className string,
	propertyName string,
	schema *Schema) *ClassModel {
	if (schema.Type == nil) || (*schema.Type.String == "object") {
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
	for definitionName, definitionSchema := range *(classes.Schema.Definitions) {
		className := classes.classNameForStub(definitionName)
		classModel := classes.buildClassForDefinition(className, definitionName, definitionSchema)
		if classModel != nil {
			classes.ClassModels[className] = classModel
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
	valueProperty.Type = "string"
	anyClass.Properties[valueProperty.Name] = valueProperty
	classes.ClassModels[anyClass.Name] = anyClass
}

func (classes *ClassCollection) sortedClassNames() []string {
	classNames := make([]string, 0)
	for className, _ := range classes.ClassModels {
		classNames = append(classNames, className)
	}
	sort.Strings(classNames)
	return classNames
}

func (classes *ClassCollection) display() string {
	classNames := classes.sortedClassNames()
	result := ""
	for _, className := range classNames {
		result += classes.ClassModels[className].display()
	}
	return result
}
