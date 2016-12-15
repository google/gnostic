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

	"github.com/googleapis/openapi-compiler/jsonschema"
)

/// Class Modeling

// models classes that we encounter during traversal that have no named schema
type ClassRequest struct {
	Name         string             // name of class to be created
	PropertyName string             // name of a property that refers to this class
	Schema       *jsonschema.Schema // schema for class
	OneOfWrapper bool               // true if the class wraps "oneOfs"
}

func NewClassRequest(name string, propertyName string, schema *jsonschema.Schema) *ClassRequest {
	return &ClassRequest{Name: name, PropertyName: propertyName, Schema: schema}
}

// models class properties, eg. fields
type ClassProperty struct {
	Name        string // name of property
	Type        string // type for property (scalar or message type)
	MapType     string // if this property is for a map, the name of the mapped type
	Repeated    bool   // true if this property is repeated (an array)
	Pattern     string // if the property is a pattern property, names must match this pattern.
	Implicit    bool   // true if this property is implied by a pattern or "additional properties" property
	Description string // if present, the "description" field in the schema
}

func (classProperty *ClassProperty) description() string {
	result := ""
	if classProperty.Description != "" {
		result += fmt.Sprintf("\t// %+s\n", classProperty.Description)
	}
	if classProperty.Repeated {
		result += fmt.Sprintf("\t%s %s repeated %s\n", classProperty.Name, classProperty.Type, classProperty.Pattern)
	} else {
		result += fmt.Sprintf("\t%s %s %s \n", classProperty.Name, classProperty.Type, classProperty.Pattern)
	}
	return result
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
	Name           string           // class name
	Properties     []*ClassProperty // slice of properties
	Required       []string         // required property names
	OneOfWrapper   bool             // true if this class wraps "oneof" properties
	Open           bool             // open classes can have keys outside the specified set
	OpenPatterns   []string         // patterns for properties that we allow
	IsStringArray  bool             // ugly override
	IsItemArray    bool             // ugly override
	IsBlob         bool             // ugly override
	IsPair         bool             // class is a name-value pair used to support ordered maps
	PairValueClass string           // class for pair values (valid if IsPair == true)
	Description    string           // if present, the "description" field in the schema
}

func (classModel *ClassModel) AddProperty(property *ClassProperty) {
	if classModel.Properties == nil {
		classModel.Properties = make([]*ClassProperty, 0)
	}
	classModel.Properties = append(classModel.Properties, property)
}

func (classModel *ClassModel) description() string {
	result := ""
	if classModel.Description != "" {
		result += fmt.Sprintf("// %+s\n", classModel.Description)
	}
	var wrapperinfo string
	if classModel.OneOfWrapper {
		wrapperinfo = " oneof wrapper"
	}
	result += fmt.Sprintf("%+s%s\n", classModel.Name, wrapperinfo)
	for _, property := range classModel.Properties {
		result += property.description()
	}
	return result
}

func NewClassModel() *ClassModel {
	classModel := &ClassModel{}
	classModel.Properties = make([]*ClassProperty, 0)
	return classModel
}

// models a collection of classes that is defined by a schema
type ClassCollection struct {
	ClassModels         map[string]*ClassModel   // models of the classes in the collection
	Prefix              string                   // class prefix to use
	Schema              *jsonschema.Schema       // top-level schema
	PatternNames        map[string]string        // a configured mapping from patterns to property names
	ObjectClassRequests map[string]*ClassRequest // anonymous classes implied by class instantiation
	MapClassRequests    map[string]string        // "NamedObject" classes that will be used to implement ordered maps
}

func NewClassCollection(schema *jsonschema.Schema) *ClassCollection {
	cc := &ClassCollection{}
	cc.ClassModels = make(map[string]*ClassModel, 0)
	cc.PatternNames = make(map[string]string, 0)
	cc.ObjectClassRequests = make(map[string]*ClassRequest, 0)
	cc.MapClassRequests = make(map[string]string, 0)
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
		return "Schema"
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
func (classes *ClassCollection) arrayItemTypeForSchema(propertyName string, schema *jsonschema.Schema) string {
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
	return itemTypeName
}

func (classes *ClassCollection) buildClassProperties(classModel *ClassModel, schema *jsonschema.Schema) {
	if schema.Properties != nil {
		for _, pair := range *(schema.Properties) {
			propertyName := pair.Name
			propertySchema := pair.Value
			if propertySchema.Ref != nil {
				// the property schema is a reference, so we will add a property with the type of the referenced schema
				propertyClassName := classes.classNameForReference(*(propertySchema.Ref))
				classProperty := NewClassProperty()
				classProperty.Name = propertyName
				classProperty.Type = propertyClassName
				classModel.AddProperty(classProperty)
			} else if propertySchema.Type != nil {
				// the property schema specifies a type, so add a property with the specified type
				if propertySchema.TypeIs("string") {
					classProperty := NewClassPropertyWithNameAndType(propertyName, "string")
					if propertySchema.Description != nil {
						classProperty.Description = *propertySchema.Description
					}
					classModel.AddProperty(classProperty)
				} else if propertySchema.TypeIs("boolean") {
					classProperty := NewClassPropertyWithNameAndType(propertyName, "bool")
					if propertySchema.Description != nil {
						classProperty.Description = *propertySchema.Description
					}
					classModel.AddProperty(classProperty)
				} else if propertySchema.TypeIs("number") {
					classProperty := NewClassPropertyWithNameAndType(propertyName, "float")
					if propertySchema.Description != nil {
						classProperty.Description = *propertySchema.Description
					}
					classModel.AddProperty(classProperty)
				} else if propertySchema.TypeIs("integer") {
					classProperty := NewClassPropertyWithNameAndType(propertyName, "int")
					if propertySchema.Description != nil {
						classProperty.Description = *propertySchema.Description
					}
					classModel.AddProperty(classProperty)
				} else if propertySchema.TypeIs("object") {
					// the property has an "anonymous" object schema, so define a new class for it and request its creation
					anonymousObjectClassName := classes.classNameForStub(propertyName)
					classes.ObjectClassRequests[anonymousObjectClassName] =
						NewClassRequest(anonymousObjectClassName, propertyName, propertySchema)
					// add a property with the type of the requested class
					classProperty := NewClassPropertyWithNameAndType(propertyName, anonymousObjectClassName)
					if propertySchema.Description != nil {
						classProperty.Description = *propertySchema.Description
					}
					classModel.AddProperty(classProperty)
				} else if propertySchema.TypeIs("array") {
					// the property has an array type, so define it as a a repeated property of the specified type
					propertyClassName := classes.arrayItemTypeForSchema(propertyName, propertySchema)
					classProperty := NewClassPropertyWithNameAndType(propertyName, propertyClassName)
					classProperty.Repeated = true
					if propertySchema.Description != nil {
						classProperty.Description = *propertySchema.Description
					}
					classModel.AddProperty(classProperty)
				} else {
					log.Printf("ignoring %+v, which has an unsupported property type '%+v'", propertyName, propertySchema.Type)
				}
			} else if propertySchema.IsEmpty() {
				// an empty schema can contain anything, so add an accessor for a generic object
				className := "Any"
				classProperty := NewClassPropertyWithNameAndType(propertyName, className)
				classModel.AddProperty(classProperty)
			} else if propertySchema.OneOf != nil {
				anonymousObjectClassName := classes.classNameForStub(propertyName + "Item")
				classes.ObjectClassRequests[anonymousObjectClassName] =
					NewClassRequest(anonymousObjectClassName, propertyName, propertySchema)
				classProperty := NewClassPropertyWithNameAndType(propertyName, anonymousObjectClassName)
				classModel.AddProperty(classProperty)
			} else if propertySchema.AnyOf != nil {
				anonymousObjectClassName := classes.classNameForStub(propertyName + "Item")
				classes.ObjectClassRequests[anonymousObjectClassName] =
					NewClassRequest(anonymousObjectClassName, propertyName, propertySchema)
				classProperty := NewClassPropertyWithNameAndType(propertyName, anonymousObjectClassName)
				classModel.AddProperty(classProperty)
			} else {
				log.Printf("ignoring %s.%s, which has an unrecognized schema:\n%+v", classModel.Name, propertyName, propertySchema.String())
			}
		}
	}
}

func (classes *ClassCollection) buildClassRequirements(classModel *ClassModel, schema *jsonschema.Schema) {
	if schema.Required != nil {
		classModel.Required = (*schema.Required)
	}
}

func (classes *ClassCollection) buildPatternPropertyAccessors(classModel *ClassModel, schema *jsonschema.Schema) {
	if schema.PatternProperties != nil {
		classModel.OpenPatterns = make([]string, 0)
		for _, pair := range *(schema.PatternProperties) {
			propertyPattern := pair.Name
			propertySchema := pair.Value
			classModel.OpenPatterns = append(classModel.OpenPatterns, propertyPattern)
			className := "Any"
			propertyName := classes.PatternNames[propertyPattern]
			if propertySchema.Ref != nil {
				className = classes.classNameForReference(*propertySchema.Ref)
			}
			propertyTypeName := fmt.Sprintf("Named%s", className)
			property := NewClassPropertyWithNameTypeAndPattern(propertyName, propertyTypeName, propertyPattern)
			property.Implicit = true
			property.MapType = className
			property.Repeated = true
			classes.MapClassRequests[property.MapType] = property.MapType
			classModel.AddProperty(property)
		}
	}
}

func (classes *ClassCollection) buildAdditionalPropertyAccessors(classModel *ClassModel, schema *jsonschema.Schema) {
	if schema.AdditionalProperties != nil {
		if schema.AdditionalProperties.Boolean != nil {
			if *schema.AdditionalProperties.Boolean == true {
				classModel.Open = true
				propertyName := "additionalProperties"
				className := "NamedAny"
				property := NewClassPropertyWithNameAndType(propertyName, className)
				property.Implicit = true
				property.MapType = "Any"
				property.Repeated = true
				classes.MapClassRequests[property.MapType] = property.MapType
				classModel.AddProperty(property)
				return
			}
		} else if schema.AdditionalProperties.Schema != nil {
			classModel.Open = true
			schema := schema.AdditionalProperties.Schema
			if schema.Ref != nil {
				propertyName := "additionalProperties"
				mapType := classes.classNameForReference(*schema.Ref)
				className := fmt.Sprintf("Named%s", mapType)
				property := NewClassPropertyWithNameAndType(propertyName, className)
				property.Implicit = true
				property.MapType = mapType
				property.Repeated = true
				classes.MapClassRequests[property.MapType] = property.MapType
				classModel.AddProperty(property)
				return
			} else if schema.Type != nil {
				typeName := *schema.Type.String
				if typeName == "string" {
					propertyName := "additionalProperties"
					className := "NamedString"
					property := NewClassPropertyWithNameAndType(propertyName, className)
					property.Implicit = true
					property.MapType = "string"
					property.Repeated = true
					classes.MapClassRequests[property.MapType] = property.MapType
					classModel.AddProperty(property)
					return
				} else if typeName == "array" {
					if schema.Items != nil {
						itemType := *schema.Items.Schema.Type.String
						if itemType == "string" {
							propertyName := "additionalProperties"
							className := "NamedStringArray"
							property := NewClassPropertyWithNameAndType(propertyName, className)
							property.Implicit = true
							property.MapType = "StringArray"
							property.Repeated = true
							classes.MapClassRequests[property.MapType] = property.MapType
							classModel.AddProperty(property)
							return
						}
					}
				}
			} else if schema.OneOf != nil {
				propertyClassName := classes.classNameForStub(classModel.Name + "Item")
				propertyName := "additionalProperties"
				className := fmt.Sprintf("Named%s", propertyClassName)
				property := NewClassPropertyWithNameAndType(propertyName, className)
				property.Implicit = true
				property.MapType = propertyClassName
				property.Repeated = true
				classes.MapClassRequests[property.MapType] = property.MapType
				classModel.AddProperty(property)

				classes.ObjectClassRequests[propertyClassName] =
					NewClassRequest(propertyClassName, propertyName, schema)
			}
		}
	}
}

func (classes *ClassCollection) buildOneOfAccessors(classModel *ClassModel, schema *jsonschema.Schema) {
	oneOfs := schema.OneOf
	if oneOfs == nil {
		return
	}
	classModel.Open = true
	classModel.OneOfWrapper = true
	for _, oneOf := range *oneOfs {
		if oneOf.Ref != nil {
			ref := *oneOf.Ref
			className := classes.classNameForReference(ref)
			propertyName := classes.propertyNameForReference(ref)

			if propertyName != nil {
				classProperty := NewClassPropertyWithNameAndType(*propertyName, className)
				classModel.AddProperty(classProperty)
			}
		} else if oneOf.Type != nil && oneOf.Type.String != nil && *oneOf.Type.String == "boolean" {
			classProperty := NewClassPropertyWithNameAndType("boolean", "bool")
			classModel.AddProperty(classProperty)
		} else {
			log.Printf("Unsupported oneOf:\n%+v", oneOf.String())
		}

	}
}

func schemaIsContainedInArray(s1 *jsonschema.Schema, s2 *jsonschema.Schema) bool {
	if s2.TypeIs("array") {
		if s2.Items.Schema != nil {
			if s1.IsEqual(s2.Items.Schema) {
				return true
			} else {
				return false
			}
		} else {
			return false
		}
	} else {
		return false
	}
}

func (classes *ClassCollection) addAnonymousAccessorForSchema(
	classModel *ClassModel,
	schema *jsonschema.Schema,
	repeated bool) {
	ref := schema.Ref
	if ref != nil {
		className := classes.classNameForReference(*ref)
		propertyName := classes.propertyNameForReference(*ref)
		if propertyName != nil {
			property := NewClassPropertyWithNameAndType(*propertyName, className)
			property.Repeated = true
			classModel.AddProperty(property)
			classModel.IsItemArray = true
		}
	} else {
		className := "string"
		propertyName := "value"
		property := NewClassPropertyWithNameAndType(propertyName, className)
		property.Repeated = true
		classModel.AddProperty(property)
		classModel.IsStringArray = true
	}
}

func (classes *ClassCollection) buildAnyOfAccessors(classModel *ClassModel, schema *jsonschema.Schema) {
	anyOfs := schema.AnyOf
	if anyOfs == nil {
		return
	}
	if len(*anyOfs) == 2 {
		if schemaIsContainedInArray((*anyOfs)[0], (*anyOfs)[1]) {
			log.Printf("ARRAY OF %+v", (*anyOfs)[0].String())
			schema := (*anyOfs)[0]
			classes.addAnonymousAccessorForSchema(classModel, schema, true)
		} else if schemaIsContainedInArray((*anyOfs)[1], (*anyOfs)[0]) {
			log.Printf("ARRAY OF %+v", (*anyOfs)[1].String())
			schema := (*anyOfs)[1]
			classes.addAnonymousAccessorForSchema(classModel, schema, true)
		} else {
			for _, anyOf := range *anyOfs {
				ref := anyOf.Ref
				if ref != nil {
					className := classes.classNameForReference(*ref)
					propertyName := classes.propertyNameForReference(*ref)
					if propertyName != nil {
						property := NewClassPropertyWithNameAndType(*propertyName, className)
						classModel.AddProperty(property)
					}
				} else {
					className := "bool"
					propertyName := "boolean"
					property := NewClassPropertyWithNameAndType(propertyName, className)
					classModel.AddProperty(property)
				}
			}
		}
	} else {
		log.Printf("Unhandled anyOfs:\n%s", schema.String())
	}
}

func (classes *ClassCollection) buildDefaultAccessors(classModel *ClassModel, schema *jsonschema.Schema) {
	classModel.Open = true
	propertyName := "additionalProperties"
	className := "NamedAny"
	property := NewClassPropertyWithNameAndType(propertyName, className)
	property.MapType = "Any"
	property.Repeated = true
	classes.MapClassRequests[property.MapType] = property.MapType
	classModel.AddProperty(property)
}

func (classes *ClassCollection) buildClassForDefinition(
	className string,
	propertyName string,
	schema *jsonschema.Schema) *ClassModel {
	if (schema.Type == nil) || (*schema.Type.String == "object") {
		return classes.buildClassForDefinitionObject(className, propertyName, schema)
	} else {
		return nil
	}
}

func (classes *ClassCollection) buildClassForDefinitionObject(
	className string,
	propertyName string,
	schema *jsonschema.Schema) *ClassModel {
	classModel := NewClassModel()
	classModel.Name = className
	if schema.IsEmpty() {
		classes.buildDefaultAccessors(classModel, schema)
	} else {
		if schema.Description != nil {
			classModel.Description = *schema.Description
		}
		classes.buildClassProperties(classModel, schema)
		classes.buildClassRequirements(classModel, schema)
		classes.buildPatternPropertyAccessors(classModel, schema)
		classes.buildAdditionalPropertyAccessors(classModel, schema)
		classes.buildOneOfAccessors(classModel, schema)
		classes.buildAnyOfAccessors(classModel, schema)
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
	for _, pair := range *(classes.Schema.Definitions) {
		definitionName := pair.Name
		definitionSchema := pair.Value
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

	// iterate over map item classes to be instantiated and generate a class for each
	mapClassNames := make([]string, 0)
	for mapClassName, _ := range classes.MapClassRequests {
		mapClassNames = append(mapClassNames, mapClassName)
	}
	sort.Strings(mapClassNames)

	for _, mapClassName := range mapClassNames {
		className := "Named" + strings.Title(mapClassName)
		classModel := NewClassModel()
		classModel.Name = className
		classModel.Description = fmt.Sprintf(
			"Automatically-generated message used to represent maps of %s as ordered (name,value) pairs.",
			mapClassName)
		classModel.IsPair = true
		classModel.PairValueClass = mapClassName

		nameProperty := NewClassProperty()
		nameProperty.Name = "name"
		nameProperty.Type = "string"
		nameProperty.Description = "Map key"
		classModel.AddProperty(nameProperty)

		valueProperty := NewClassProperty()
		valueProperty.Name = "value"
		valueProperty.Type = mapClassName
		valueProperty.Description = "Mapped value"
		classModel.AddProperty(valueProperty)

		classes.ClassModels[className] = classModel
	}

	// add a class for string arrays
	stringArrayClass := NewClassModel()
	stringArrayClass.Name = "StringArray"
	stringProperty := NewClassProperty()
	stringProperty.Name = "value"
	stringProperty.Type = "string"
	stringProperty.Repeated = true
	stringArrayClass.AddProperty(stringProperty)
	classes.ClassModels[stringArrayClass.Name] = stringArrayClass

	// add a class for "Any"
	anyClass := NewClassModel()
	anyClass.Name = "Any"
	anyClass.Open = true
	anyClass.IsBlob = true
	valueProperty := NewClassProperty()
	valueProperty.Name = "value"
	valueProperty.Type = "blob"
	anyClass.AddProperty(valueProperty)
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

func (classes *ClassCollection) description() string {
	classNames := classes.sortedClassNames()
	result := ""
	for _, className := range classNames {
		result += classes.ClassModels[className].description()
	}
	return result
}
