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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// This is a global map of all known Schemas.
// It is initialized when the first Schema is created and inserted.
var schemas map[string]*Schema

// This struct models a JSON Schema and, because schemas are defined
// hierarchically, contains many references to itself.
// All fields are pointers and are nil if the associated values
// are not specified.
type Schema struct {
	Schema *string // $schema
	Id     *string // id keyword used for $ref resolution scope
	Ref    *string // $ref, i.e. JSON Pointers

	// http://json-schema.org/latest/json-schema-validation.html
	// 5.1.  Validation keywords for numeric instances (number and integer)
	MultipleOf       *SchemaNumber
	Maximum          *SchemaNumber
	ExclusiveMaximum *bool
	Minimum          *SchemaNumber
	ExclusiveMinimum *bool

	// 5.2.  Validation keywords for strings
	MaxLength *int64
	MinLength *int64
	Pattern   *string

	// 5.3.  Validation keywords for arrays
	AdditionalItems *SchemaOrBoolean
	Items           *SchemaOrSchemaArray
	MaxItems        *int64
	MinItems        *int64
	UniqueItems     *bool

	// 5.4.  Validation keywords for objects
	MaxProperties        *int64
	MinProperties        *int64
	Required             *[]string
	AdditionalProperties *SchemaOrBoolean
	Properties           *map[string]*Schema
	PatternProperties    *map[string]*Schema
	Dependencies         *map[string]*SchemaOrStringArray

	// 5.5.  Validation keywords for any instance type
	Enumeration *[]SchemaEnumValue
	Type        *StringOrStringArray
	AllOf       *[]*Schema
	AnyOf       *[]*Schema
	OneOf       *[]*Schema
	Not         *Schema
	Definitions *map[string]*Schema

	// 6.  Metadata keywords
	Title       *string
	Description *string
	Default     *interface{}

	// 7.  Semantic validation with "format"
	Format *string
}

// These helper structs model "combination" types that generally can
// have values of one type or another. All are used to represent parts
// of Schemas.

type SchemaNumber struct {
	Integer *int64
	Float   *float64
}

type SchemaOrBoolean struct {
	Schema  *Schema
	Boolean *bool
}

type StringOrStringArray struct {
	String *string
	Array  *[]string
}

type SchemaOrStringArray struct {
	Schema *Schema
	Array  *[]string
}

type SchemaOrSchemaArray struct {
	Schema *Schema
	Array  *[]*Schema
}

type SchemaEnumValue struct {
	String *string
	Bool   *bool
}

// Reads a schema from a file.
// Currently this assumes that schemas are stored in the source distribution of this project.
func NewSchemaFromFile(filename string) *Schema {
	schemasDir := os.Getenv("GOPATH") + "/src/github.com/googleapis/openapi-compiler/schemas"
	file, e := ioutil.ReadFile(schemasDir + "/" + filename)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}
	var info interface{}
	json.Unmarshal(file, &info)
	return NewSchemaFromObject(info)
}

// Constructs a schema from a parsed JSON object.
// Due to the complexity of the schema representation, this is a
// custom reader and not the standard Go JSON reader (encoding/json).
func NewSchemaFromObject(jsonData interface{}) *Schema {
	switch t := jsonData.(type) {
	default:
		fmt.Printf("schemaValue: unexpected type %T\n", t)
		return nil
	case map[string]interface{}:
		schema := &Schema{}
		for k, v := range t {

			switch k {
			case "$schema":
				schema.Schema = schema.stringValue(v)
			case "id":
				schema.Id = schema.stringValue(v)

			case "multipleOf":
				schema.MultipleOf = schema.numberValue(v)
			case "maximum":
				schema.Maximum = schema.numberValue(v)
			case "exclusiveMaximum":
				schema.ExclusiveMaximum = schema.boolValue(v)
			case "minimum":
				schema.Minimum = schema.numberValue(v)
			case "exclusiveMinimum":
				schema.ExclusiveMinimum = schema.boolValue(v)

			case "maxLength":
				schema.MaxLength = schema.intValue(v)
			case "minLength":
				schema.MinLength = schema.intValue(v)
			case "pattern":
				schema.Pattern = schema.stringValue(v)

			case "additionalItems":
				schema.AdditionalItems = schema.schemaOrBooleanValue(v)
			case "items":
				schema.Items = schema.schemaOrSchemaArrayValue(v)
			case "maxItems":
				schema.MaxItems = schema.intValue(v)
			case "minItems":
				schema.MinItems = schema.intValue(v)
			case "uniqueItems":
				schema.UniqueItems = schema.boolValue(v)

			case "maxProperties":
				schema.MaxProperties = schema.intValue(v)
			case "minProperties":
				schema.MinProperties = schema.intValue(v)
			case "required":
				schema.Required = schema.arrayOfStringsValue(v)
			case "additionalProperties":
				schema.AdditionalProperties = schema.schemaOrBooleanValue(v)
			case "properties":
				schema.Properties = schema.mapOfSchemasValue(v)
			case "patternProperties":
				schema.PatternProperties = schema.mapOfSchemasValue(v)
			case "dependencies":
				schema.Dependencies = schema.mapOfSchemasOrStringArraysValue(v)

			case "enum":
				schema.Enumeration = schema.arrayOfEnumValuesValue(v)

			case "type":
				schema.Type = schema.stringOrStringArrayValue(v)
			case "allOf":
				schema.AllOf = schema.arrayOfSchemasValue(v)
			case "anyOf":
				schema.AnyOf = schema.arrayOfSchemasValue(v)
			case "oneOf":
				schema.OneOf = schema.arrayOfSchemasValue(v)
			case "not":
				schema.Not = NewSchemaFromObject(v)
			case "definitions":
				schema.Definitions = schema.mapOfSchemasValue(v)

			case "title":
				schema.Title = schema.stringValue(v)
			case "description":
				schema.Description = schema.stringValue(v)

			case "default":
				schema.Default = &v

			case "format":
				schema.Format = schema.stringValue(v)
			case "$ref":
				schema.Ref = schema.stringValue(v)
			default:
				fmt.Printf("UNSUPPORTED (%s)\n", k)
			}
		}

		// insert schema in global map
		if schema.Id != nil {
			if schemas == nil {
				schemas = make(map[string]*Schema, 0)
			}
			schemas[*(schema.Id)] = schema
		}
		return schema
	}
	return nil
}

// Returns true if no members of the Schema are specified.
func (schema *Schema) isEmpty() bool {
	return (schema.Schema == nil) &&
		(schema.Id == nil) &&
		(schema.MultipleOf == nil) &&
		(schema.Maximum == nil) &&
		(schema.ExclusiveMaximum == nil) &&
		(schema.Minimum == nil) &&
		(schema.ExclusiveMinimum == nil) &&
		(schema.MaxLength == nil) &&
		(schema.MinLength == nil) &&
		(schema.Pattern == nil) &&
		(schema.AdditionalItems == nil) &&
		(schema.Items == nil) &&
		(schema.MaxItems == nil) &&
		(schema.MinItems == nil) &&
		(schema.UniqueItems == nil) &&
		(schema.MaxProperties == nil) &&
		(schema.MinProperties == nil) &&
		(schema.Required == nil) &&
		(schema.AdditionalProperties == nil) &&
		(schema.Properties == nil) &&
		(schema.PatternProperties == nil) &&
		(schema.Dependencies == nil) &&
		(schema.Enumeration == nil) &&
		(schema.Type == nil) &&
		(schema.AllOf == nil) &&
		(schema.AnyOf == nil) &&
		(schema.OneOf == nil) &&
		(schema.Not == nil) &&
		(schema.Definitions == nil) &&
		(schema.Title == nil) &&
		(schema.Description == nil) &&
		(schema.Default == nil) &&
		(schema.Format == nil) &&
		(schema.Ref == nil)
}

//
// BUILDERS
// The following methods build elements of Schemas from interface{} values.
// Each returns nil if it is unable to build the desired element.
//

// Gets the string value of an interface{} value if possible.
func (schema *Schema) stringValue(v interface{}) *string {
	switch v := v.(type) {
	default:
		fmt.Printf("stringValue: unexpected type %T\n", v)
	case string:
		return &v
	}
	return nil
}

// Gets the numeric value of an interface{} value if possible.
func (schema *Schema) numberValue(v interface{}) *SchemaNumber {
	number := &SchemaNumber{}
	switch v := v.(type) {
	default:
		fmt.Printf("numberValue: unexpected type %T\n", v)
	case float64:
		v2 := float64(v)
		number.Float = &v2
		return number
	case float32:
		v2 := float64(v)
		number.Float = &v2
		return number
	}
	return nil
}

// Gets the integer value of an interface{} value if possible.
func (schema *Schema) intValue(v interface{}) *int64 {
	switch v := v.(type) {
	default:
		fmt.Printf("intValue: unexpected type %T\n", v)
	case float64:
		v2 := int64(v)
		return &v2
	case int64:
		return &v
	}
	return nil
}

// Gets the bool value of an interface{} value if possible.
func (schema *Schema) boolValue(v interface{}) *bool {
	switch v := v.(type) {
	default:
		fmt.Printf("boolValue: unexpected type %T\n", v)
	case bool:
		return &v
	}
	return nil
}

// Gets a map of Schemas from an interface{} value if possible.
func (schema *Schema) mapOfSchemasValue(v interface{}) *map[string]*Schema {
	switch v := v.(type) {
	default:
		fmt.Printf("mapOfSchemasValue: unexpected type %T\n", v)
	case map[string]interface{}:
		m := make(map[string]*Schema)
		for k2, v2 := range v {
			m[k2] = NewSchemaFromObject(v2)
		}
		return &m
	}
	return nil
}

// Gets an array of Schemas from an interface{} value if possible.
func (schema *Schema) arrayOfSchemasValue(v interface{}) *[]*Schema {
	switch v := v.(type) {
	default:
		fmt.Printf("arrayOfSchemasValue: unexpected type %T\n", v)
	case []interface{}:
		m := make([]*Schema, 0)
		for _, v2 := range v {
			switch v2 := v2.(type) {
			default:
				fmt.Printf("arrayOfSchemasValue: unexpected type %T\n", v2)
			case map[string]interface{}:
				s := NewSchemaFromObject(v2)
				m = append(m, s)
			}
		}
		return &m
	case map[string]interface{}:
		m := make([]*Schema, 0)
		s := NewSchemaFromObject(v)
		m = append(m, s)
		return &m
	}
	return nil
}

// Gets a Schema or an array of Schemas from an interface{} value if possible.
func (schema *Schema) schemaOrSchemaArrayValue(v interface{}) *SchemaOrSchemaArray {
	switch v := v.(type) {
	default:
		fmt.Printf("schemaOrSchemaArrayValue: unexpected type %T\n", v)
	case []interface{}:
		m := make([]*Schema, 0)
		for _, v2 := range v {
			switch v2 := v2.(type) {
			default:
				fmt.Printf("schemaOrSchemaArrayValue: unexpected type %T\n", v2)
			case map[string]interface{}:
				s := NewSchemaFromObject(v2)
				m = append(m, s)
			}
		}
		return &SchemaOrSchemaArray{Array: &m}
	case map[string]interface{}:
		s := NewSchemaFromObject(v)
		return &SchemaOrSchemaArray{Schema: s}
	}
	return nil
}

// Gets an array of strings from an interface{} value if possible.
func (schema *Schema) arrayOfStringsValue(v interface{}) *[]string {
	switch v := v.(type) {
	default:
		fmt.Printf("arrayOfStringsValue: unexpected type %T\n", v)
	case []string:
		return &v
	case string:
		a := []string{v}
		return &a
	case []interface{}:
		a := make([]string, 0)
		for _, v2 := range v {
			switch v2 := v2.(type) {
			default:
				fmt.Printf("arrayOfStringsValue: unexpected type %T\n", v2)
			case string:
				a = append(a, v2)
			}
		}
		return &a
	}
	return nil
}

// Gets a string or an array of strings from an interface{} value if possible.
func (schema *Schema) stringOrStringArrayValue(v interface{}) *StringOrStringArray {
	switch v := v.(type) {
	default:
		fmt.Printf("arrayOfStringsValue: unexpected type %T\n", v)
	case []string:
		s := &StringOrStringArray{}
		s.Array = &v
		return s
	case string:
		s := &StringOrStringArray{}
		s.String = &v
		return s
	case []interface{}:
		a := make([]string, 0)
		for _, v2 := range v {
			switch v2 := v2.(type) {
			default:
				fmt.Printf("arrayOfStringsValue: unexpected type %T\n", v2)
			case string:
				a = append(a, v2)
			}
		}
		s := &StringOrStringArray{}
		s.Array = &a
		return s
	}
	return nil
}

// Gets an array of enum values from an interface{} value if possible.
func (schema *Schema) arrayOfEnumValuesValue(v interface{}) *[]SchemaEnumValue {
	a := make([]*SchemaEnumValue, 0)
	switch v := v.(type) {
	default:
		fmt.Printf("arrayOfEnumValuesValue: unexpected type %T\n", v)
	case []interface{}:
		for _, v2 := range v {
			switch v2 := v2.(type) {
			default:
				fmt.Printf("arrayOfEnumValuesValue: unexpected type %T\n", v2)
			case string:
				a = append(a, &SchemaEnumValue{String: &v2})
			case bool:
				a = append(a, &SchemaEnumValue{Bool: &v2})
			}
		}
	}
	return nil
}

// Gets a map of schemas or string arrays from an interface{} value if possible.
func (schema *Schema) mapOfSchemasOrStringArraysValue(v interface{}) *map[string]*SchemaOrStringArray {
	m := make(map[string]*SchemaOrStringArray, 0)
	switch v := v.(type) {
	default:
		fmt.Printf("mapOfSchemasOrStringArraysValue: unexpected type %T %+v\n", v, v)
	case map[string]interface{}:
		for k2, v2 := range v {
			switch v2 := v2.(type) {
			default:
				fmt.Printf("mapOfSchemasOrStringArraysValue: unexpected type %T %+v\n", v2, v2)
			case []interface{}:
				a := make([]string, 0)
				for _, v3 := range v2 {
					switch v3 := v3.(type) {
					default:
						fmt.Printf("mapOfSchemasOrStringArraysValue: unexpected type %T %+v\n", v3, v3)
					case string:
						a = append(a, v3)
					}
				}
				s := &SchemaOrStringArray{}
				s.Array = &a
				m[k2] = s
			}
		}
	}
	return &m
}

// Gets a schema or a boolean value from an interface{} value if possible.
func (schema *Schema) schemaOrBooleanValue(v interface{}) *SchemaOrBoolean {
	schemaOrBoolean := &SchemaOrBoolean{}
	switch v := v.(type) {
	case bool:
		schemaOrBoolean.Boolean = &v
	case map[string]interface{}:
		schemaOrBoolean.Schema = NewSchemaFromObject(v)
	default:
		fmt.Printf("schemaOrBooleanValue: unexpected type %T\n", v)
	case []map[string]interface{}:

	}
	return schemaOrBoolean
}

//
// DISPLAY
// The following methods display Schemas.
//

// Returns a string representation of a Schema.
func (schema *Schema) display() string {
	return schema.displaySchema("")
}

// Helper: Returns a string representation of a Schema indented by a specified string.
func (schema *Schema) displaySchema(indent string) string {
	result := ""
	if schema.Schema != nil {
		result += indent + "$schema: " + *(schema.Schema) + "\n"
	}
	if schema.Id != nil {
		result += indent + "id: " + *(schema.Id) + "\n"
	}
	if schema.MultipleOf != nil {
		result += indent + fmt.Sprintf("multipleOf: %+v\n", *(schema.MultipleOf))
	}
	if schema.Maximum != nil {
		result += indent + fmt.Sprintf("maximum: %+v\n", *(schema.Maximum))
	}
	if schema.ExclusiveMaximum != nil {
		result += indent + fmt.Sprintf("exclusiveMaximum: %+v\n", *(schema.ExclusiveMaximum))
	}
	if schema.Minimum != nil {
		result += indent + fmt.Sprintf("minimum: %+v\n", *(schema.Minimum))
	}
	if schema.ExclusiveMinimum != nil {
		result += indent + fmt.Sprintf("exclusiveMinimum: %+v\n", *(schema.ExclusiveMinimum))
	}
	if schema.MaxLength != nil {
		result += indent + fmt.Sprintf("maxLength: %+v\n", *(schema.MaxLength))
	}
	if schema.MinLength != nil {
		result += indent + fmt.Sprintf("minLength: %+v\n", *(schema.MinLength))
	}
	if schema.Pattern != nil {
		result += indent + fmt.Sprintf("pattern: %+v\n", *(schema.Pattern))
	}
	if schema.AdditionalItems != nil {
		s := schema.AdditionalItems.Schema
		if s != nil {
			result += indent + "additionalItems:\n"
			result += s.displaySchema(indent + "  ")
		} else {
			b := *(schema.AdditionalItems.Boolean)
			result += indent + fmt.Sprintf("additionalItems: %+v\n", b)
		}
	}
	if schema.Items != nil {
		result += indent + "items:\n"
		items := schema.Items
		if items.Array != nil {
			for i, s := range *(items.Array) {
				result += indent + "  " + fmt.Sprintf("%d", i) + ":\n"
				result += s.displaySchema(indent + "  " + "  ")
			}
		} else if items.Schema != nil {
			result += items.Schema.displaySchema(indent + "  " + "  ")
		}
	}
	if schema.MaxItems != nil {
		result += indent + fmt.Sprintf("maxItems: %+v\n", *(schema.MaxItems))
	}
	if schema.MinItems != nil {
		result += indent + fmt.Sprintf("minItems: %+v\n", *(schema.MinItems))
	}
	if schema.UniqueItems != nil {
		result += indent + fmt.Sprintf("uniqueItems: %+v\n", *(schema.UniqueItems))
	}
	if schema.MaxProperties != nil {
		result += indent + fmt.Sprintf("maxProperties: %+v\n", *(schema.MaxProperties))
	}
	if schema.MinProperties != nil {
		result += indent + fmt.Sprintf("minProperties: %+v\n", *(schema.MinProperties))
	}
	if schema.Required != nil {
		result += indent + fmt.Sprintf("required: %+v\n", *(schema.Required))
	}
	if schema.AdditionalProperties != nil {
		s := schema.AdditionalProperties.Schema
		if s != nil {
			result += indent + "additionalProperties:\n"
			result += s.displaySchema(indent + "  ")
		} else {
			b := *(schema.AdditionalProperties.Boolean)
			result += indent + fmt.Sprintf("additionalProperties: %+v\n", b)
		}
	}
	if schema.Properties != nil {
		result += indent + "properties:\n"
		for name, s := range *(schema.Properties) {
			result += indent + "  " + name + ":\n"
			result += s.displaySchema(indent + "  " + "  ")
		}
	}
	if schema.PatternProperties != nil {
		result += indent + "patternProperties:\n"
		for name, s := range *(schema.PatternProperties) {
			result += indent + "  " + name + ":\n"
			result += s.displaySchema(indent + "  " + "  ")
		}
	}
	if schema.Dependencies != nil {
		result += indent + "dependencies:\n"
		for name, schemaOrStringArray := range *(schema.Dependencies) {
			s := schemaOrStringArray.Schema
			if s != nil {
				result += indent + "  " + name + ":\n"
				result += s.displaySchema(indent + "  " + "  ")
			} else {
				a := schemaOrStringArray.Array
				if a != nil {
					result += indent + "  " + name + ":\n"
					for _, s2 := range *a {
						result += indent + "  " + "  " + s2 + "\n"
					}
				}
			}

		}
	}
	if schema.Enumeration != nil {
		result += indent + "enumeration:\n"
		for _, value := range *(schema.Enumeration) {
			if value.String != nil {
				result += indent + "  " + fmt.Sprintf("%+v\n", value.String)
			} else {
				result += indent + "  " + fmt.Sprintf("%+v\n", value.Bool)
			}
		}
	}
	if schema.Type != nil {
		result += indent + fmt.Sprintf("type: %+v\n", *(schema.Type))
	}
	if schema.AllOf != nil {
		result += indent + "allOf:\n"
		for _, s := range *(schema.AllOf) {
			result += s.displaySchema(indent + "  ")
			result += indent + "-\n"
		}
	}
	if schema.AnyOf != nil {
		result += indent + "anyOf:\n"
		for _, s := range *(schema.AnyOf) {
			result += s.displaySchema(indent + "  ")
			result += indent + "-\n"
		}
	}
	if schema.OneOf != nil {
		result += indent + "oneOf:\n"
		for _, s := range *(schema.OneOf) {
			result += s.displaySchema(indent + "  ")
			result += indent + "-\n"
		}
	}
	if schema.Not != nil {
		result += indent + "not:\n"
		result += schema.Not.displaySchema(indent + "  ")
	}
	if schema.Definitions != nil {
		result += indent + "definitions:\n"
		for name, s := range *(schema.Definitions) {
			result += indent + "  " + name + ":\n"
			result += s.displaySchema(indent + "  " + "  ")
		}
	}
	if schema.Title != nil {
		result += indent + "title: " + *(schema.Title) + "\n"
	}
	if schema.Description != nil {
		result += indent + "description: " + *(schema.Description) + "\n"
	}
	if schema.Default != nil {
		result += indent + "default:\n"
		result += indent + fmt.Sprintf("  %+v\n", *(schema.Default))
	}
	if schema.Format != nil {
		result += indent + "format: " + *(schema.Format) + "\n"
	}
	if schema.Ref != nil {
		result += indent + "$ref: " + *(schema.Ref) + "\n"
	}
	return result
}

//
// OPERATIONS
// The following methods perform operations on Schemas.
//

// A type that represents a function that can be applied to a Schema.
type SchemaOperation func(schema *Schema, context string)

// Applies a specified function to a Schema and all of the Schemas that it contains.
func (schema *Schema) applyToSchemas(operation SchemaOperation, context string) {

	if schema.AdditionalItems != nil {
		s := schema.AdditionalItems.Schema
		if s != nil {
			s.applyToSchemas(operation, "AdditionalItems")
		}
	}

	if schema.Items != nil {
		if schema.Items.Array != nil {
			for _, s := range *(schema.Items.Array) {
				s.applyToSchemas(operation, "Items.Array")
			}
		} else if schema.Items.Schema != nil {
			schema.Items.Schema.applyToSchemas(operation, "Items.Schema")
		}
	}

	if schema.AdditionalProperties != nil {
		s := schema.AdditionalProperties.Schema
		if s != nil {
			s.applyToSchemas(operation, "AdditionalProperties")
		}
	}

	if schema.Properties != nil {
		for _, s := range *(schema.Properties) {
			s.applyToSchemas(operation, "Properties")
		}
	}
	if schema.PatternProperties != nil {
		for _, s := range *(schema.PatternProperties) {
			s.applyToSchemas(operation, "PatternProperties")
		}
	}

	if schema.Dependencies != nil {
		for _, schemaOrStringArray := range *(schema.Dependencies) {
			s := schemaOrStringArray.Schema
			if s != nil {
				s.applyToSchemas(operation, "Dependencies")
			}
		}
	}

	if schema.AllOf != nil {
		for _, s := range *(schema.AllOf) {
			s.applyToSchemas(operation, "AllOf")
		}
	}
	if schema.AnyOf != nil {
		for _, s := range *(schema.AnyOf) {
			s.applyToSchemas(operation, "AnyOf")
		}
	}
	if schema.OneOf != nil {
		for _, s := range *(schema.OneOf) {
			s.applyToSchemas(operation, "OneOf")
		}
	}
	if schema.Not != nil {
		schema.Not.applyToSchemas(operation, "Not")
	}

	if schema.Definitions != nil {
		for _, s := range *(schema.Definitions) {
			s.applyToSchemas(operation, "Definitions")
		}
	}

	operation(schema, context)
}

// Copies all non-nil properties from the source Schema to the destination Schema.
func (destination *Schema) copyProperties(source *Schema) {
	if source.Schema != nil {
		destination.Schema = source.Schema
	}
	if source.Id != nil {
		destination.Id = source.Id
	}
	if source.MultipleOf != nil {
		destination.MultipleOf = source.MultipleOf
	}
	if source.Maximum != nil {
		destination.Maximum = source.Maximum
	}
	if source.ExclusiveMaximum != nil {
		destination.ExclusiveMaximum = source.ExclusiveMaximum
	}
	if source.Minimum != nil {
		destination.Minimum = source.Minimum
	}
	if source.ExclusiveMinimum != nil {
		destination.ExclusiveMinimum = source.ExclusiveMinimum
	}
	if source.MaxLength != nil {
		destination.MaxLength = source.MaxLength
	}
	if source.MinLength != nil {
		destination.MinLength = source.MinLength
	}
	if source.Pattern != nil {
		destination.Pattern = source.Pattern
	}
	if source.AdditionalItems != nil {
		destination.AdditionalItems = source.AdditionalItems
	}
	if source.Items != nil {
		destination.Items = source.Items
	}
	if source.MaxItems != nil {
		destination.MaxItems = source.MaxItems
	}
	if source.MinItems != nil {
		destination.MinItems = source.MinItems
	}
	if source.UniqueItems != nil {
		destination.UniqueItems = source.UniqueItems
	}
	if source.MaxProperties != nil {
		destination.MaxProperties = source.MaxProperties
	}
	if source.MinProperties != nil {
		destination.MinProperties = source.MinProperties
	}
	if source.Required != nil {
		destination.Required = source.Required
	}
	if source.AdditionalProperties != nil {
		destination.AdditionalProperties = source.AdditionalProperties
	}
	if source.Properties != nil {
		destination.Properties = source.Properties
	}
	if source.PatternProperties != nil {
		destination.PatternProperties = source.PatternProperties
	}
	if source.Dependencies != nil {
		destination.Dependencies = source.Dependencies
	}
	if source.Enumeration != nil {
		destination.Enumeration = source.Enumeration
	}
	if source.Type != nil {
		destination.Type = source.Type
	}
	if source.AllOf != nil {
		destination.AllOf = source.AllOf
	}
	if source.AnyOf != nil {
		destination.AnyOf = source.AnyOf
	}
	if source.OneOf != nil {
		destination.OneOf = source.OneOf
	}
	if source.Not != nil {
		destination.Not = source.Not
	}
	if source.Definitions != nil {
		destination.Definitions = source.Definitions
	}
	if source.Title != nil {
		destination.Title = source.Title
	}
	if source.Description != nil {
		destination.Description = source.Description
	}
	if source.Default != nil {
		destination.Default = source.Default
	}
	if source.Format != nil {
		destination.Format = source.Format
	}
	if source.Ref != nil {
		destination.Ref = source.Ref
	}
}

// Returns true if the "type" of a Schema includes the specified type
func (schema *Schema) typeIs(typeName string) bool {
	if schema.Type != nil {
		// the type is either a string or an array of strings
		if schema.Type.String != nil {
			return (*(schema.Type.String) == typeName)
		} else if schema.Type.Array != nil {
			for _, n := range *(schema.Type.Array) {
				if n == typeName {
					return true
				}
			}
		}
	}
	return false
}

// Resolves "$ref" elements in a Schema and its children.
// But if a reference refers to an object type or is inside a oneOf,
// the reference is kept and we expect downstream tools to separately model these
// referenced schemas.
func (schema *Schema) resolveRefs() {
	rootSchema := schema
	count := 1
	for count > 0 {
		count = 0
		schema.applyToSchemas(
			func(schema *Schema, context string) {
				if schema.Ref != nil {
					resolvedRef := rootSchema.resolveJSONPointer(*(schema.Ref))
					if resolvedRef.typeIs("object") {
						// don't substitute, we'll model the referenced item with a class
					} else if context == "OneOf" {
						// don't substitute, we'll model the referenced item with a class
					} else {
						schema.Ref = nil
						schema.copyProperties(resolvedRef)
						count += 1
					}
				}
			}, "")
	}
}

// Resolves JSON pointers.
// This current implementation is very crude and custom for OpenAPI 2.0 schemas.
// It panics for any pointer that it is unable to resolve.
func (root *Schema) resolveJSONPointer(ref string) *Schema {
	var result *Schema

	parts := strings.Split(ref, "#")
	if len(parts) == 2 {
		documentName := parts[0] + "#"
		if documentName == "#" {
			documentName = *(root.Id)
		}
		path := parts[1]
		document := schemas[documentName]
		pathParts := strings.Split(path, "/")

		// we currently do a very limited (hard-coded) resolution of certain paths and log errors for missed cases
		if len(pathParts) == 1 {
			return document
		} else if len(pathParts) == 3 {
			switch pathParts[1] {
			case "definitions":
				dictionary := document.Definitions
				result = (*dictionary)[pathParts[2]]
			case "properties":
				dictionary := document.Properties
				result = (*dictionary)[pathParts[2]]
			default:
				break
			}
		}
	}
	if result == nil {
		panic(fmt.Sprintf("UNRESOLVED POINTER: %+v", ref))
	}
	return result
}

// Replaces "allOf" elements by merging their properties into the parent Schema.
func (schema *Schema) resolveAllOfs() {
	schema.applyToSchemas(
		func(schema *Schema, context string) {
			if schema.AllOf != nil {
				for _, allOf := range *(schema.AllOf) {
					schema.copyProperties(allOf)
				}
				schema.AllOf = nil
			}
		}, "")
}

// Replaces all "anyOf" elements by merging their properties into the parent Schema.
func (schema *Schema) resolveAnyOfs() {
	schema.applyToSchemas(
		func(schema *Schema, context string) {
			if schema.AnyOf != nil {
				for _, anyOf := range *(schema.AnyOf) {
					schema.copyProperties(anyOf)
				}
				schema.AnyOf = nil
			}
		}, "")
}

// Flattens any "oneOf" elements that contain other "oneOf" elements.
func (schema *Schema) flattenOneOfs() {
	schema.applyToSchemas(
		func(schema *Schema, context string) {
			if schema.OneOf != nil {
				newOneOfs := make([]*Schema, 0)
				for _, oneOf := range *(schema.OneOf) {
					innerOneOfs := oneOf.OneOf
					if innerOneOfs != nil {
						for _, innerOneOf := range *innerOneOfs {
							newOneOfs = append(newOneOfs, innerOneOf)
						}
					} else {
						newOneOfs = append(newOneOfs, oneOf)
					}
				}
				schema.OneOf = &newOneOfs
			}
		}, "")
}
