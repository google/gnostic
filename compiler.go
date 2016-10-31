//go:generate ./COMPILE-PROTOS.sh

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"sort"
	"strings"
)

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
	Items           *[]*Schema
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
	Type        *[]string
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

type SchemaOrStringArray struct {
	Schema *Schema
	Array  *[]string
}

type SchemaEnumValue struct {
	String *string
	Bool   *bool
}

func NewSchemaFromFile(filename string) *Schema {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	schemasDir := usr.HomeDir + "/go/src/github.com/googleapis/openapi-compiler/schemas"
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
				schema.Items = schema.arrayOfSchemasValue(v)
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
				schema.Type = schema.arrayOfStringsValue(v)
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
		for i, s := range *(schema.Items) {
			result += indent + "  " + fmt.Sprintf("%d", i) + ":\n"
			result += s.displaySchema(indent + "  " + "  ")
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
		for _, s := range *(schema.Items) {
			s.applyToSchemas(operation)
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
	destination.Schema = source.Schema
	destination.Id = source.Id
	destination.MultipleOf = source.MultipleOf
	destination.Maximum = source.Maximum
	destination.ExclusiveMaximum = source.ExclusiveMaximum
	destination.Minimum = source.Minimum
	destination.ExclusiveMinimum = source.ExclusiveMinimum
	destination.MaxLength = source.MaxLength
	destination.MinLength = source.MinLength
	destination.Pattern = source.Pattern
	destination.AdditionalItems = source.AdditionalItems
	destination.Items = source.Items
	destination.MaxItems = source.MaxItems
	destination.MinItems = source.MinItems
	destination.UniqueItems = source.UniqueItems
	destination.MaxProperties = source.MaxProperties
	destination.MinProperties = source.MinProperties
	destination.Required = source.Required
	destination.AdditionalProperties = source.AdditionalProperties
	destination.Properties = source.Properties
	destination.PatternProperties = source.PatternProperties
	destination.Dependencies = source.Dependencies
	destination.Enumeration = source.Enumeration
	destination.Type = source.Type
	destination.AllOf = source.AllOf
	destination.AnyOf = source.AnyOf
	destination.OneOf = source.OneOf
	destination.Not = source.Not
	destination.Definitions = source.Definitions
	destination.Title = source.Title
	destination.Description = source.Description
	destination.Default = source.Default
	destination.Format = source.Format
	destination.Ref = source.Ref
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
					if (resolvedRef.Type != nil) && ((*(resolvedRef.Type))[0] == "object") {
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

func (classModel *ClassModel) display() string {
	result := fmt.Sprintf("%+s\n", classModel.Name)

	keys := make([]string, 0)
	for k, _ := range classModel.Properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)

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

func (classes *ClassCollection) arrayTypeForSchema(schema *Schema) string {
	// what is the array type?
	itemTypeName := "google.protobuf.Any"
	ref := (*schema.Items)[0].Ref
	if ref != nil {
		itemTypeName = classes.classNameForReference(*ref)
	} else {
		types := (*schema.Items)[0].Type
		if len(*types) == 1 {
			itemTypeName = (*types)[0]
		} else if len(*types) > 1 {
			itemTypeName = fmt.Sprintf("%+v", types)
		} else {
			itemTypeName = "UNKNOWN"
		}
	}
	return itemTypeName
}

func (classes *ClassCollection) buildClassProperties(classModel *ClassModel, schema *Schema, path string) {
	for key, value := range *(schema.Properties) {
		if value.Ref != nil {
			className := classes.classNameForReference(*(value.Ref))
			cp := NewClassProperty()
			cp.Name = key
			cp.Type = className
			classModel.Properties[key] = cp
		} else {
			if value.Type != nil {
				propertyType := (*value.Type)[0]
				switch propertyType {
				case "string":
					classModel.Properties[key] = NewClassPropertyWithNameAndType(key, "string")
				case "boolean":
					classModel.Properties[key] = NewClassPropertyWithNameAndType(key, "bool")
				case "number":
					classModel.Properties[key] = NewClassPropertyWithNameAndType(key, "float")
				case "integer":
					classModel.Properties[key] = NewClassPropertyWithNameAndType(key, "int")
				case "object":
					className := classes.classNameForStub(key)
					classes.ObjectClassRequests[className] = NewClassRequest(path, className, value)
					classModel.Properties[key] = NewClassPropertyWithNameAndType(key, className)
				case "array":
					className := classes.arrayTypeForSchema(value)
					p := NewClassPropertyWithNameAndType(key, className)
					p.Repeated = true
					classModel.Properties[key] = p
				case "default":
					log.Printf("%+v:%+v has unsupported property type %+v", path, key, propertyType)
				}
			} else {
				/*
				   if value.isEmpty() {
				     // write accessor for generic object
				     let className = "google.protobuf.Any"
				     classModel.properties[key] = ClassProperty(name:key, type:className)
				   } else if value.anyOf != nil {
				     //self.writeAnyOfAccessors(schema: value, path: path, accessorName:accessorName)
				   } else if value.oneOf != nil {
				     //self.writeOneOfAccessors(schema: value, path: path)
				   } else {
				     //print("\(path):\(key) has unspecified property type. Schema is below.\n\(value.description)")
				   }
				*/
			}
		}
	}
}

func (classes *ClassCollection) buildClassRequirements(classModel *ClassModel, schema *Schema, path string) {
	if schema.Required != nil {
		classModel.Required = (*schema.Required)
	}
}

func (classes *ClassCollection) buildClassForDefinition(className string, schema *Schema) *ClassModel {
	classModel := NewClassModel()
	classModel.Name = className
	classes.buildClassProperties(classModel, classes.Schema, "")
	classes.buildClassRequirements(classModel, classes.Schema, "")
	return classModel
}

func (classes *ClassCollection) buildClassForDefinitionObject(className string, schema *Schema) *ClassModel {
	classModel := NewClassModel()
	classModel.Name = className
	classes.buildClassProperties(classModel, classes.Schema, "")
	classes.buildClassRequirements(classModel, classes.Schema, "")
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
		classes.ClassModels[className] = classes.buildClassForDefinition(className, value)
	}

	// iterate over anonymous object classes to be instantiated and generate a class for each
	for className, classRequest := range classes.ObjectClassRequests {
		classes.ClassModels[classRequest.Name] =
			classes.buildClassForDefinitionObject(className, classRequest.Schema)
	}

	// add a class for string arrays
	stringArrayClass := NewClassModel()
	stringArrayClass.Name = "OpenAPIStringArray"
	stringProperty := NewClassProperty()
	stringProperty.Name = "string"
	stringProperty.Type = "string"
	stringProperty.Repeated = true
	stringArrayClass.Properties[stringProperty.Name] = stringProperty
	classes.ClassModels[stringArrayClass.Name] = stringArrayClass
}

func (classes *ClassCollection) display() string {
	keys := make([]string, 0)
	for k, _ := range classes.ClassModels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	result := ""
	for _, k := range keys {
		result += classes.ClassModels[k].display()
	}
	return result
}

/// main program

func main() {
	base_schema := NewSchemaFromFile("schema.json")
	base_schema.resolveRefs(nil)

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

	fmt.Printf("%s\n", openapi_schema.display())

	// build a simplified model of the classes described by the schema
	cc := NewClassCollection(openapi_schema)
	cc.build()
	fmt.Printf("%s\n", cc.display())
}
