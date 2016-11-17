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
	"log"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
)

const LICENSE = "" +
	"// Copyright 2016 Google Inc. All Rights Reserved.\n" +
	"//\n" +
	"// Licensed under the Apache License, Version 2.0 (the \"License\");\n" +
	"// you may not use this file except in compliance with the License.\n" +
	"// You may obtain a copy of the License at\n" +
	"//\n" +
	"//    http://www.apache.org/licenses/LICENSE-2.0\n" +
	"//\n" +
	"// Unless required by applicable law or agreed to in writing, software\n" +
	"// distributed under the License is distributed on an \"AS IS\" BASIS,\n" +
	"// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.\n" +
	"// See the License for the specific language governing permissions and\n" +
	"// limitations under the License.\n"

// global map of all known Schemas.
// initialized when the first Schema is created and inserted.
var schemas map[string]*Schema

type Schema struct {
	Schema      *string // $schema
	Id          *string // id keyword used for $ref resolution scope
	Ref         *string // $ref, i.e. JSON Pointers
	ResolvedRef *Schema // the resolved pointer reference

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

// Helpers

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
				schema.Properties = schema.dictionaryOfSchemasValue(v)
			case "patternProperties":
				schema.PatternProperties = schema.dictionaryOfSchemasValue(v)
			case "dependencies":
				schema.Dependencies = schema.dictionaryOfSchemasOrStringArraysValue(v)

			case "enum":
				schema.Enumeration = schema.arrayOfValuesValue(v)

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
				schema.Definitions = schema.dictionaryOfSchemasValue(v)

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

func (schema *Schema) stringValue(v interface{}) *string {
	switch v := v.(type) {
	default:
		fmt.Printf("stringValue: unexpected type %T\n", v)
	case string:
		return &v
	}
	return nil
}

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

func (schema *Schema) boolValue(v interface{}) *bool {
	switch v := v.(type) {
	default:
		fmt.Printf("boolValue: unexpected type %T\n", v)
	case bool:
		return &v
	}
	return nil
}

func (schema *Schema) dictionaryOfSchemasValue(v interface{}) *map[string]*Schema {
	switch v := v.(type) {
	default:
		fmt.Printf("dictionaryOfSchemasValue: unexpected type %T\n", v)
	case map[string]interface{}:
		m := make(map[string]*Schema)
		for k2, v2 := range v {
			m[k2] = NewSchemaFromObject(v2)
		}
		return &m
	}
	return nil
}

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

func (schema *Schema) arrayOfValuesValue(v interface{}) *[]SchemaEnumValue {
	a := make([]*SchemaEnumValue, 0)
	switch v := v.(type) {
	default:
		fmt.Printf("arrayOfValuesValue: unexpected type %T\n", v)
	case []interface{}:
		for _, v2 := range v {
			switch v2 := v2.(type) {
			default:
				fmt.Printf("arrayOfValuesValue: unexpected type %T\n", v2)
			case string:
				a = append(a, &SchemaEnumValue{String: &v2})
			case bool:
				a = append(a, &SchemaEnumValue{Bool: &v2})
			}
		}
	}
	return nil
}

func (schema *Schema) dictionaryOfSchemasOrStringArraysValue(v interface{}) *map[string]*SchemaOrStringArray {
	m := make(map[string]*SchemaOrStringArray, 0)
	switch v := v.(type) {
	default:
		fmt.Printf("dictionaryOfSchemasOrStringArraysValue: unexpected type %T %+v\n", v, v)
	case map[string]interface{}:
		for k2, v2 := range v {
			switch v2 := v2.(type) {
			default:
				fmt.Printf("dictionaryOfSchemasOrStringArraysValue: unexpected type %T %+v\n", v2, v2)
			case []interface{}:
				a := make([]string, 0)
				for _, v3 := range v2 {
					switch v3 := v3.(type) {
					default:
						fmt.Printf("dictionaryOfSchemasOrStringArraysValue: unexpected type %T %+v\n", v3, v3)
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

func (schema *Schema) display() string {
	return schema.displaySchema("")
}

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

type operation func(schema *Schema)

func (schema *Schema) applyToSchemas(operation operation) {

	if schema.AdditionalItems != nil {
		s := schema.AdditionalItems.Schema
		if s != nil {
			s.applyToSchemas(operation)
		}
	}

	if schema.Items != nil {
		if schema.Items.Array != nil {
			for _, s := range *(schema.Items.Array) {
				s.applyToSchemas(operation)
			}
		} else if schema.Items.Schema != nil {
			schema.Items.Schema.applyToSchemas(operation)
		}
	}

	if schema.AdditionalProperties != nil {
		s := schema.AdditionalProperties.Schema
		if s != nil {
			s.applyToSchemas(operation)
		}
	}

	if schema.Properties != nil {
		for _, s := range *(schema.Properties) {
			s.applyToSchemas(operation)
		}
	}
	if schema.PatternProperties != nil {
		for _, s := range *(schema.PatternProperties) {
			s.applyToSchemas(operation)
		}
	}

	if schema.Dependencies != nil {
		for _, schemaOrStringArray := range *(schema.Dependencies) {
			s := schemaOrStringArray.Schema
			if s != nil {
				s.applyToSchemas(operation)
			}
		}
	}

	if schema.AllOf != nil {
		for _, s := range *(schema.AllOf) {
			s.applyToSchemas(operation)
		}
	}
	if schema.AnyOf != nil {
		for _, s := range *(schema.AnyOf) {
			s.applyToSchemas(operation)
		}
	}
	if schema.OneOf != nil {
		for _, s := range *(schema.OneOf) {
			s.applyToSchemas(operation)
		}
	}
	if schema.Not != nil {
		schema.Not.applyToSchemas(operation)
	}

	if schema.Definitions != nil {
		for _, s := range *(schema.Definitions) {
			s.applyToSchemas(operation)
		}
	}

	operation(schema)
}

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

func (schema *Schema) typeIs(typeName string) bool {
	if schema.Type != nil {
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

func (schema *Schema) resolveRefs(classNames []string) {
	rootSchema := schema
	contains := func(stringArray []string, element string) bool {
		for _, item := range stringArray {
			if item == element {
				return true
			}
		}
		return false
	}
	count := 1
	for count > 0 {
		count = 0
		schema.applyToSchemas(
			func(schema *Schema) {
				if schema.Ref != nil {
					resolvedRef := rootSchema.resolveJSONPointer(*(schema.Ref))
					if resolvedRef.typeIs("object") {
						// don't substitute, we'll model the referenced item with a class
					} else if contains(classNames, *(schema.Ref)) {
						// don't substitute, we'll model the referenced item with a class
					} else {
						schema.Ref = nil
						schema.copyProperties(resolvedRef)
						count += 1
					}
				}
			})
	}
}

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

func (schema *Schema) resolveAllOfs() {
	schema.applyToSchemas(
		func(schema *Schema) {
			if schema.AllOf != nil {
				for _, allOf := range *(schema.AllOf) {
					schema.copyProperties(allOf)
				}
				schema.AllOf = nil
			}
		})
}

func (schema *Schema) reduceOneOfs() {
	schema.applyToSchemas(
		func(schema *Schema) {
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
		})
}

/// Class Modeling

// models classes that we encounter during traversal that have no named schema
type ClassRequest struct {
	Path   string
	Name   string
	Schema *Schema
}

func NewClassRequest(path string, name string, schema *Schema) *ClassRequest {
	return &ClassRequest{Path: path, Name: name, Schema: schema}
}

// models class properties, eg. fields
type ClassProperty struct {
	Name     string
	Type     string
	Repeated bool
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
	itemTypeName := "google.protobuf.Any"
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
	return itemTypeName
}

func (classes *ClassCollection) buildClassProperties(classModel *ClassModel, schema *Schema, path string) {
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
						classes.ObjectClassRequests[className] = NewClassRequest(path, className, value)
						classModel.Properties[key] = NewClassPropertyWithNameAndType(key, className)
					} else if value.typeIs("array") {
						className := classes.arrayTypeForSchema(value)
						p := NewClassPropertyWithNameAndType(key, className)
						p.Repeated = true
						classModel.Properties[key] = p
					} else {
						log.Printf("%+v:%+v has unsupported property type %+v", path, key, value.Type)
					}
				} else {
					if value.isEmpty() {
						// write accessor for generic object
						className := "google.protobuf.Any"
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

func (classes *ClassCollection) buildClassRequirements(classModel *ClassModel, schema *Schema, path string) {
	if schema.Required != nil {
		classModel.Required = (*schema.Required)
	}
}

func (classes *ClassCollection) buildPatternPropertyAccessors(classModel *ClassModel, schema *Schema, path string) {
	if schema.PatternProperties != nil {
		for key, propertySchema := range *(schema.PatternProperties) {
			className := "google.protobuf.Any"
			propertyName := classes.PatternNames[key]
			if propertySchema.Ref != nil {
				className = classes.classNameForReference(*propertySchema.Ref)
			}
			typeName := fmt.Sprintf("map<string, %s>", className)
			classModel.Properties[propertyName] = NewClassPropertyWithNameAndType(propertyName, typeName)
		}
	}
}

func (classes *ClassCollection) buildAdditionalPropertyAccessors(classModel *ClassModel, schema *Schema, path string) {

	if schema.AdditionalProperties != nil {
		if schema.AdditionalProperties.Boolean != nil {
			if *schema.AdditionalProperties.Boolean == true {
				propertyName := "additionalProperties"
				className := "map<string, google.protobuf.Any>"
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

func (classes *ClassCollection) buildOneOfAccessors(classModel *ClassModel, schema *Schema, path string) {
	if schema.OneOf != nil {
		classes.buildOneOfAccessorsHelper(classModel, schema.OneOf)
	}
}

func (classes *ClassCollection) buildOneOfAccessorsHelper(classModel *ClassModel, oneOfs *[]*Schema) {
	for _, oneOf := range *oneOfs {
		if oneOf.Ref != nil {
			ref := *oneOf.Ref
			className := classes.classNameForReference(ref)
			propertyName := classes.propertyNameForReference(ref)
			if propertyName != nil {
				classModel.Properties[*propertyName] = NewClassPropertyWithNameAndType(*propertyName, className)
			}
		}
	}
}

func (classes *ClassCollection) buildDefaultAccessors(classModel *ClassModel, schema *Schema, path string) {
	key := "additionalProperties"
	className := "map<string, google.protobuf.Any>"
	classModel.Properties[key] = NewClassPropertyWithNameAndType(key, className)
}

func (classes *ClassCollection) buildClassForDefinition(className string, schema *Schema) *ClassModel {
	if schema.Type == nil {
		return classes.buildClassForDefinitionObject(className, schema)
	}
	typeString := *schema.Type.String
	if typeString == "object" {
		return classes.buildClassForDefinitionObject(className, schema)
	} else {
		return nil
	}
}

func (classes *ClassCollection) buildClassForDefinitionObject(className string, schema *Schema) *ClassModel {
	classModel := NewClassModel()
	classModel.Name = className
	if schema.isEmpty() {
		classes.buildDefaultAccessors(classModel, schema, "")
	}
	classes.buildClassProperties(classModel, schema, "")
	classes.buildClassRequirements(classModel, schema, "")
	classes.buildPatternPropertyAccessors(classModel, schema, "")
	classes.buildAdditionalPropertyAccessors(classModel, schema, "")
	classes.buildOneOfAccessors(classModel, schema, "")
	return classModel
}

func (classes *ClassCollection) build() {
	// create a class for the top-level schema
	className := classes.Prefix + "Document"
	classModel := NewClassModel()
	classModel.Name = className
	classes.buildClassProperties(classModel, classes.Schema, "")
	classes.buildClassRequirements(classModel, classes.Schema, "")

	classes.ClassModels[className] = classModel

	// create a class for each object defined in the schema
	for key, value := range *(classes.Schema.Definitions) {
		className := classes.classNameForStub(key)
		model := classes.buildClassForDefinition(className, value)
		if model != nil {
			classes.ClassModels[className] = model
		}
	}

	// iterate over anonymous object classes to be instantiated and generate a class for each
	for className, classRequest := range classes.ObjectClassRequests {
		classes.ClassModels[classRequest.Name] =
			classes.buildClassForDefinitionObject(className, classRequest.Schema)
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

func (classes *ClassCollection) generateProto(packageName string) string {
	code := CodeBuilder{}
	code.AddLine(LICENSE)
	code.AddLine("// THIS FILE IS AUTOMATICALLY GENERATED.")
	code.AddLine()

	code.AddLine("syntax = \"proto3\";")
	code.AddLine()
	code.AddLine("package " + packageName + ";")
	code.AddLine()
	code.AddLine("import \"google/protobuf/any.proto\";")
	code.AddLine()

	classNames := classes.sortedClassNames()
	for _, className := range classNames {
		code.AddLine("message %s {", className)
		classModel := classes.ClassModels[className]
		propertyNames := classModel.sortedPropertyNames()
		var fieldNumber = 0
		for _, propertyName := range propertyNames {
			propertyModel := classModel.Properties[propertyName]
			fieldNumber += 1
			propertyType := propertyModel.Type
			if propertyType == "int" {
				propertyType = "int64"
			}
			var displayName = propertyName
			if displayName == "$ref" {
				displayName = "_ref"
			}
			if displayName == "$schema" {
				displayName = "_schema"
			}
			displayName = camelCaseToSnakeCase(displayName)

			var line = fmt.Sprintf("%s %s = %d;", propertyType, displayName, fieldNumber)
			if propertyModel.Repeated {
				line = "repeated " + line
			}
			code.AddLine("  " + line)
		}
		code.AddLine("}")
		code.AddLine()
	}
	return code.Text()
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

func (classes *ClassCollection) generateCompiler(packageName string) string {
	code := CodeBuilder{}
	code.AddLine(LICENSE)
	code.AddLine("// THIS FILE IS AUTOMATICALLY GENERATED.")
	code.AddLine()
	code.AddLine("package main")
	code.AddLine()
	code.AddLine("import (")
	code.AddLine("\"fmt\"")
	code.AddLine("\"log\"")
	code.AddLine("pb \"openapi\"")
	code.AddLine(")")
	code.AddLine()
	code.AddLine("func version() string {")
	code.AddLine("  return \"OpenAPIv2\"")
	code.AddLine("}")
	code.AddLine()

	classNames := classes.sortedClassNames()
	for _, className := range classNames {
		code.AddLine("func build%sForMap(in interface{}) *pb.%s {", className, className)
		code.AddLine("m, keys, ok := unpackMap(in)")
		code.AddLine("if (!ok) {")
		code.AddLine("return nil")
		code.AddLine("}")
		code.AddLine("fmt.Printf(\"%%d\\n\", len(m))")
		code.AddLine("fmt.Printf(\"%%+v\\n\", keys)")
		code.AddLine("  x := &pb.%s{}", className)

		classModel := classes.ClassModels[className]
		propertyNames := classModel.sortedPropertyNames()
		var fieldNumber = 0
		for _, propertyName := range propertyNames {
			propertyModel := classModel.Properties[propertyName]
			fieldNumber += 1
			propertyType := propertyModel.Type
			if propertyType == "int" {
				propertyType = "int64"
			}
			var displayName = propertyName
			if displayName == "$ref" {
				displayName = "_ref"
			}
			if displayName == "$schema" {
				displayName = "_schema"
			}
			displayName = camelCaseToSnakeCase(displayName)

			var line = fmt.Sprintf("%s %s = %d;", propertyType, displayName, fieldNumber)
			if propertyModel.Repeated {
				line = "repeated " + line
			}
			code.AddLine("//  " + line)

			fieldName := strings.Title(propertyName)
			if propertyName == "$ref" {
				fieldName = "XRef"
			}

			code.AddLine("if mapHasKey(m, \"%s\") {", propertyName)

			if propertyType == "string" {
				if propertyModel.Repeated {
					code.AddLine("v, ok := m[\"%v\"].([]interface{})", propertyName)
					code.AddLine("if ok {")
					code.AddLine("x.%s = convertInterfaceArrayToStringArray(v)", fieldName)
					code.AddLine("} else {")
					code.AddLine(" log.Printf(\"unexpected: %%+v\", m[\"%v\"])", propertyName)
					code.AddLine("}")
				} else {
					code.AddLine("x.%s = m[\"%v\"].(string)", fieldName, propertyName)
				}
			}

			code.AddLine("}")
		}

		code.AddLine("  return x")
		code.AddLine("}\n")
	}

	//document.Swagger = "2.0"
	//document.BasePath = "example.com"

	//info := &pb.Info{}
	//info.Title = "Sample API"
	//info.Description = "My great new API"
	//info.Version = "v1.0"
	//document.Info = info

	return code.Text()
}

/// main program

func main() {
	base_schema := NewSchemaFromFile("schema.json")
	base_schema.resolveRefs(nil)
	base_schema.resolveAllOfs()

	openapi_schema := NewSchemaFromFile("openapi-2.0.json")
	// these non-object definitions are marked for handling as if they were objects
	// in the future, these could be automatically identified by their presence in a oneOf
	classNames := []string{
		"#/definitions/headerParameterSubSchema",
		"#/definitions/formDataParameterSubSchema",
		"#/definitions/queryParameterSubSchema",
		"#/definitions/pathParameterSubSchema"}
	openapi_schema.resolveRefs(classNames)
	openapi_schema.resolveAllOfs()
	openapi_schema.reduceOneOfs()

	// build a simplified model of the classes described by the schema
	cc := NewClassCollection(openapi_schema)
	// these pattern names are a bit of a hack until we find a more automated way to obtain them
	cc.PatternNames = map[string]string{
		"^x-": "vendorExtension",
		"^/":  "path",
		"^([0-9]{3})$|^(default)$": "responseCode",
	}
	cc.build()

	var err error

	// generate the protocol buffer description
	proto := cc.generateProto("OpenAPIv2")
	proto_filename := "openapi-v2.proto"
	err = ioutil.WriteFile(proto_filename, []byte(proto), 0644)
	if err != nil {
		panic(err)
	}

	// generate the compiler
	compiler := cc.generateCompiler("OpenAPIv2")
	go_filename := "openapi-v2.go"
	err = ioutil.WriteFile(go_filename, []byte(compiler), 0644)
	if err != nil {
		panic(err)
	}
	// autoformat the compiler
	err = exec.Command(runtime.GOROOT()+"/bin/gofmt", "-w", go_filename).Run()
}
