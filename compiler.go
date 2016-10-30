//go:generate ./COMPILE-PROTOS.sh

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"jsonschema"
	"log"
	"os"
	"os/user"
	"strings"
)

func stringValue(v interface{}) string {
	switch v := v.(type) {
	default:
		fmt.Printf("stringValue: unexpected type %T\n", v)
	case string:
		return v
	}
	return ""
}

func numberValue(v interface{}) *jsonschema.Number {
	number := &jsonschema.Number{}
	switch v := v.(type) {
	default:
		fmt.Printf("numberValue: unexpected type %T\n", v)
	case float64:
		number.Value = &jsonschema.Number_Float{float32(v)}
		return number
	case float32:
		number.Value = &jsonschema.Number_Float{float32(v)}
		return number
	}
	return nil
}

func intValue(v interface{}) int64 {
	switch v := v.(type) {
	default:
		fmt.Printf("intValue: unexpected type %T\n", v)
	case float64:
		return int64(v)
	case int64:
		return v
	}
	return 0
}

func boolValue(v interface{}) bool {
	switch v := v.(type) {
	default:
		fmt.Printf("boolValue: unexpected type %T\n", v)
	case bool:
		return v
	}
	return false
}

func dictionaryOfSchemasValue(v interface{}) map[string]*jsonschema.Schema {
	switch v := v.(type) {
	default:
		fmt.Printf("dictionaryOfSchemasValue: unexpected type %T\n", v)
	case map[string]interface{}:
		m := make(map[string]*jsonschema.Schema)
		for k2, v2 := range v {
			m[k2] = schemaValue(v2)
		}
		return m
	}
	return nil
}

func arrayOfSchemasValue(v interface{}) []*jsonschema.Schema {
	switch v := v.(type) {
	default:
		fmt.Printf("arrayOfSchemasValue: unexpected type %T\n", v)
	case []interface{}:
		m := make([]*jsonschema.Schema, 0)
		for _, v2 := range v {
			switch v2 := v2.(type) {
			default:
				fmt.Printf("arrayOfSchemasValue: unexpected type %T\n", v2)
			case map[string]interface{}:
				s := schemaValue(v2)
				m = append(m, s)
			}
		}
		return m
	case map[string]interface{}:
		m := make([]*jsonschema.Schema, 0)
		s := schemaValue(v)
		m = append(m, s)
		return m
	}
	return nil
}

func arrayOfStringsValue(v interface{}) []string {
	switch v := v.(type) {
	default:
		fmt.Printf("arrayOfStringsValue: unexpected type %T\n", v)
	case []string:
		return v
	case string:
		return []string{v}
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
		return a
	}
	return []string{}
}

func arrayOfValuesValue(v interface{}) []*jsonschema.Value {
	a := make([]*jsonschema.Value, 0)
	switch v := v.(type) {
	default:
		fmt.Printf("arrayOfValuesValue: unexpected type %T\n", v)
	case []interface{}:
		for _, v2 := range v {
			switch v2 := v2.(type) {
			default:
				fmt.Printf("arrayOfValuesValue: unexpected type %T\n", v2)
			case string:
				vv := &jsonschema.Value{String_: v2}
				a = append(a, vv)
			case bool:
				vv := &jsonschema.Value{Bool: v2}
				a = append(a, vv)
			}
		}
	}
	return a
}

func dictionaryOfSchemasOrStringArraysValue(v interface{}) map[string]*jsonschema.SchemaOrStringArray {
	m := make(map[string]*jsonschema.SchemaOrStringArray, 0)
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
				s := &jsonschema.SchemaOrStringArray{}
				s.Value = &jsonschema.SchemaOrStringArray_Array{Array: &jsonschema.StringArray{String_: a}}
				m[k2] = s
			}
		}
	}
	return m
}

func schemaOrBooleanValue(v interface{}) *jsonschema.SchemaOrBoolean {
	schemaOrBoolean := &jsonschema.SchemaOrBoolean{}
	switch v := v.(type) {
	case bool:
		schemaOrBoolean.Value = &jsonschema.SchemaOrBoolean_Boolean{Boolean: v}
	case map[string]interface{}:
		schemaOrBoolean.Value = &jsonschema.SchemaOrBoolean_Schema{Schema: schemaValue(v)}
	default:
		fmt.Printf("schemaOrBooleanValue: unexpected type %T\n", v)
	case []map[string]interface{}:

	}
	return schemaOrBoolean
}

func schemaValue(jsonData interface{}) *jsonschema.Schema {
	switch t := jsonData.(type) {
	default:
		fmt.Printf("schemaValue: unexpected type %T\n", t)
	case map[string]interface{}:
		schema := &jsonschema.Schema{}
		for k, v := range t {

			switch k {
			case "$schema":
				schema.Schema = stringValue(v)
			case "id":
				schema.Id = stringValue(v)

			case "multipleOf":
				schema.MultipleOf = numberValue(v)
			case "maximum":
				schema.Maximum = numberValue(v)
			case "exclusiveMaximum":
				schema.ExclusiveMaximum = boolValue(v)
			case "minimum":
				schema.Minimum = numberValue(v)
			case "exclusiveMinimum":
				schema.ExclusiveMinimum = boolValue(v)

			case "maxLength":
				schema.MaxLength = intValue(v)
			case "minLength":
				schema.MinLength = intValue(v)
			case "pattern":
				schema.Pattern = stringValue(v)

			case "additionalItems":
				schema.AdditionalItems = schemaOrBooleanValue(v)
			case "items":
				schema.Items = arrayOfSchemasValue(v)
			case "maxItems":
				schema.MaxItems = intValue(v)
			case "minItems":
				schema.MinItems = intValue(v)
			case "uniqueItems":
				schema.UniqueItems = boolValue(v)

			case "maxProperties":
				schema.MaxProperties = intValue(v)
			case "minProperties":
				schema.MinProperties = intValue(v)
			case "required":
				schema.Required = arrayOfStringsValue(v)
			case "additionalProperties":
				schema.AdditionalProperties = schemaOrBooleanValue(v)
			case "properties":
				schema.Properties = dictionaryOfSchemasValue(v)
			case "patternProperties":
				schema.PatternProperties = dictionaryOfSchemasValue(v)
			case "dependencies":
				schema.Dependencies = dictionaryOfSchemasOrStringArraysValue(v)

			case "enum":
				schema.Enumeration = arrayOfValuesValue(v)

			case "type":
				schema.Type = arrayOfStringsValue(v)
			case "allOf":
				schema.AllOf = arrayOfSchemasValue(v)
			case "anyOf":
				schema.AnyOf = arrayOfSchemasValue(v)
			case "oneOf":
				schema.OneOf = arrayOfSchemasValue(v)
			case "not":
				schema.Not = schemaValue(v)
			case "definitions":
				schema.Definitions = dictionaryOfSchemasValue(v)

			case "title":
				schema.Title = stringValue(v)
			case "description":
				schema.Description = stringValue(v)

			case "default":
				//schema.DefaultValue = v

			case "format":
				schema.Format = stringValue(v)
			case "$ref":
				schema.Ref = stringValue(v)
			default:
				fmt.Printf("UNSUPPORTED (%s)\n", k)
			}
		}
		return schema

	}
	return nil
}

func loadSchema(filename string) *jsonschema.Schema {
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
	return schemaValue(info)
}

func displaySchema(schema *jsonschema.Schema, indent string) string {
	result := ""
	if schema.Schema != "" {
		result += indent + "$schema: " + schema.Schema + "\n"
	}
	if schema.Id != "" {
		result += indent + "id: " + schema.Id + "\n"
	}
	if floatForNumber(schema.MultipleOf) != 0.0 {
		result += indent + fmt.Sprintf("multipleOf: %+v\n", schema.MultipleOf)
	}
	if floatForNumber(schema.Maximum) != 0.0 {
		result += indent + fmt.Sprintf("maximum: %+v\n", schema.Maximum)
	}
	if schema.ExclusiveMaximum {
		result += indent + fmt.Sprintf("exclusiveMaximum: %+v\n", schema.ExclusiveMaximum)
	}
	if floatForNumber(schema.Minimum) != 0.0 {
		result += indent + fmt.Sprintf("minimum: %+v\n", schema.Minimum)
	}
	if schema.ExclusiveMinimum {
		result += indent + fmt.Sprintf("exclusiveMinimum: %+v\n", schema.ExclusiveMinimum)
	}
	if schema.MaxLength != 0 {
		result += indent + fmt.Sprintf("maxLength: %+v\n", schema.MaxLength)
	}
	if schema.MinLength != 0 {
		result += indent + fmt.Sprintf("minLength: %+v\n", schema.MinLength)
	}
	if schema.Pattern != "" {
		result += indent + fmt.Sprintf("pattern: %+v\n", schema.Pattern)
	}
	if schema.AdditionalItems != nil {
		s := schema.AdditionalItems.GetSchema()
		if s != nil {
			result += indent + "additionalItems:\n"
			result += displaySchema(s, indent+"  ")
		} else {
			b := schema.AdditionalItems.GetBoolean()
			result += indent + fmt.Sprintf("additionalItems: %+v\n", b)
		}
	}
	if len(schema.Items) > 0 {
		result += indent + "items:\n"
		for i, schema := range schema.Items {
			result += indent + "  " + fmt.Sprintf("%d", i) + ":\n"
			result += displaySchema(schema, indent+"  "+"  ")
		}
	}
	if schema.MaxItems != 0 {
		result += indent + fmt.Sprintf("maxItems: %+v\n", schema.MaxItems)
	}
	if schema.MinItems != 0 {
		result += indent + fmt.Sprintf("minItems: %+v\n", schema.MinItems)
	}
	if schema.UniqueItems {
		result += indent + fmt.Sprintf("uniqueItems: %+v\n", schema.UniqueItems)
	}
	if schema.MaxProperties != 0 {
		result += indent + fmt.Sprintf("maxProperties: %+v\n", schema.MaxProperties)
	}
	if schema.MinProperties != 0 {
		result += indent + fmt.Sprintf("minProperties: %+v\n", schema.MinProperties)
	}
	if len(schema.Required) > 0 {
		result += indent + fmt.Sprintf("required: %+v\n", schema.Required)
	}
	if schema.AdditionalProperties != nil {
		s := schema.AdditionalProperties.GetSchema()
		if s != nil {
			result += indent + "additionalProperties:\n"
			result += displaySchema(s, indent+"  ")
		} else {
			b := schema.AdditionalProperties.GetBoolean()
			result += indent + fmt.Sprintf("additionalProperties: %+v\n", b)
		}
	}
	if len(schema.Properties) > 0 {
		result += indent + "properties:\n"
		for name, schema := range schema.Properties {
			result += indent + "  " + name + ":\n"
			result += displaySchema(schema, indent+"  "+"  ")
		}
	}
	if len(schema.PatternProperties) > 0 {
		result += indent + "patternProperties:\n"
		for name, schema := range schema.PatternProperties {
			result += indent + "  " + name + ":\n"
			result += displaySchema(schema, indent+"  "+"  ")
		}
	}
	if len(schema.Dependencies) > 0 {
		result += indent + "dependencies:\n"
		for name, schemaOrStringArray := range schema.Dependencies {
			s := schemaOrStringArray.GetSchema()
			if s != nil {
				result += indent + "  " + name + ":\n"
				result += displaySchema(s, indent+"  "+"  ")
			} else {
				a := schemaOrStringArray.GetArray()
				if a != nil {
					result += indent + "  " + name + ":\n"
					for _, s2 := range a.String_ {
						result += indent + "  " + "  " + s2 + "\n"
					}
				}
			}

		}
	}
	if len(schema.Enumeration) > 0 {
		result += indent + "enumeration:\n"
		for _, value := range schema.Enumeration {
			if value.String_ != "" {
				result += indent + "  " + fmt.Sprintf("%+v\n", value.String_)
			} else {
				result += indent + "  " + fmt.Sprintf("%+v\n", value.Bool)
			}
		}
	}
	if len(schema.Type) > 0 {
		result += indent + fmt.Sprintf("type: %+v\n", schema.Type)
	}
	if len(schema.AllOf) > 0 {
		result += indent + "allOf:\n"
		for _, schema := range schema.AllOf {
			result += displaySchema(schema, indent+"  ")
			result += indent + "-\n"
		}
	}
	if len(schema.AnyOf) > 0 {
		result += indent + "anyOf:\n"
		for _, schema := range schema.AnyOf {
			result += displaySchema(schema, indent+"  ")
			result += indent + "-\n"
		}
	}
	if len(schema.OneOf) > 0 {
		result += indent + "oneOf:\n"
		for _, schema := range schema.OneOf {
			result += displaySchema(schema, indent+"  ")
			result += indent + "-\n"
		}
	}
	if schema.Not != nil {
		result += indent + "not:\n"
		result += displaySchema(schema.Not, indent+"  ")
	}
	if len(schema.Definitions) > 0 {
		result += indent + "definitions:\n"
		for name, schema := range schema.Definitions {
			result += indent + "  " + name + ":\n"
			result += displaySchema(schema, indent+"  "+"  ")
		}
	}
	if schema.Title != "" {
		result += indent + "title: " + schema.Title + "\n"
	}
	if schema.Description != "" {
		result += indent + "description: " + schema.Description + "\n"
	}
	/*
	   if let defaultValue = defaultValue {
	     result += indent + "default:\n"
	     result += indent + "  \(defaultValue)\n"
	   }
	*/
	if schema.Format != "" {
		result += indent + "format: " + schema.Format + "\n"
	}
	if schema.Ref != "" {
		result += indent + "$ref: " + schema.Ref + "\n"
	}
	return result
}

func floatForNumber(n *jsonschema.Number) float64 {
	return 0
}

type operation func(schema *jsonschema.Schema)

func applyToSchemas(schema *jsonschema.Schema, operation operation) {

	if schema.AdditionalItems != nil {
		s := schema.AdditionalItems.GetSchema()
		if s != nil {
			operation(s)
		}
	}

	if len(schema.Items) > 0 {
		for _, schema := range schema.Items {
			operation(schema)
		}
	}

	if schema.AdditionalProperties != nil {
		s := schema.AdditionalProperties.GetSchema()
		if s != nil {
			operation(s)
		}
	}

	if len(schema.Properties) > 0 {
		for _, s := range schema.Properties {
			operation(s)
		}
	}
	if len(schema.PatternProperties) > 0 {
		for _, s := range schema.PatternProperties {
			operation(s)
		}
	}

	if len(schema.Dependencies) > 0 {
		for _, schemaOrStringArray := range schema.Dependencies {
			s := schemaOrStringArray.GetSchema()
			if s != nil {
				operation(s)
			}
		}
	}

	if len(schema.AllOf) > 0 {
		for _, schema := range schema.AllOf {
			operation(schema)
		}
	}
	if len(schema.AnyOf) > 0 {
		for _, schema := range schema.AnyOf {
			operation(schema)
		}
	}
	if len(schema.OneOf) > 0 {
		for _, schema := range schema.OneOf {
			operation(schema)
		}
	}
	if schema.Not != nil {
		operation(schema.Not)
	}

	if len(schema.Definitions) > 0 {
		for _, s := range schema.Definitions {
			operation(s)
		}
	}

	operation(schema)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func copyProperties(destination *jsonschema.Schema, source *jsonschema.Schema) {
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
	destination.DefaultValue = source.DefaultValue
	destination.Format = source.Format
	destination.Ref = source.Ref
}

func resolveRefs(schema *jsonschema.Schema, rootSchema *jsonschema.Schema, classNames []string) {
	applyToSchemas(schema,
		func(schema *jsonschema.Schema) {
			if schema.Ref != "" {
				log.Printf("REF %+v\n", schema.Ref)
				resolvedRef := resolveJSONPointer(schema.Ref, rootSchema)
				if (len(resolvedRef.Type) > 0) && (resolvedRef.Type[0] == "object") {
					// don't substitute, we'll model the referenced item with a class
				} else if contains(classNames, schema.Ref) {
					// don't substitute, we'll model the referenced item with a class
				} else {
					schema.Ref = ""
					copyProperties(schema, resolvedRef)
				}
			}
		})
}

func resolveJSONPointer(ref string, root *jsonschema.Schema) *jsonschema.Schema {
	var result *jsonschema.Schema
	result = nil

	parts := strings.Split(ref, "#")
	if len(parts) == 2 {
		documentName := parts[0] + "#"
		if documentName == "#" {
			documentName = root.Id
		}
		path := parts[1]
		document := schemas[documentName]
		pathParts := strings.Split(path, "/")

		if len(pathParts) == 1 {
			return document
		} else if len(pathParts) == 3 {
			switch pathParts[1] {
			case "definitions":
				dictionary := document.Definitions
				result = dictionary[pathParts[2]]

			case "properties":
				dictionary := document.Properties
				result = dictionary[pathParts[2]]

			default:
				break
			}
		}
	}
	if result == nil {
		print("UNRESOLVED REF: " + ref)
	}
	return result
}

func resolveAllOfs(schema *jsonschema.Schema, root *jsonschema.Schema) {
	applyToSchemas(schema,
		func(schema *jsonschema.Schema) {

			for _, allOf := range schema.AllOf {
				copyProperties(schema, allOf)
			}
			schema.AllOf = []*jsonschema.Schema{}
		})
}

func reduceOneOfs(schema *jsonschema.Schema, root *jsonschema.Schema) {
	applyToSchemas(schema,
		func(schema *jsonschema.Schema) {
			if len(schema.OneOf) > 0 {
				newOneOfs := make([]*jsonschema.Schema, 0)
				for _, oneOf := range schema.OneOf {
					innerOneOfs := oneOf.OneOf
					if len(innerOneOfs) > 0 {
						for _, innerOneOf := range innerOneOfs {
							newOneOfs = append(newOneOfs, innerOneOf)
						}
					} else {
						newOneOfs = append(newOneOfs, oneOf)
					}
				}
				schema.OneOf = newOneOfs
			}
		})
}

var schemas map[string]*jsonschema.Schema

func main() {
	schemas = make(map[string]*jsonschema.Schema, 0)

	var s *jsonschema.Schema
	s = loadSchema("schema.json")
	schemas[s.Id] = s

	s = loadSchema("openapi-2.0.json")
	schemas[s.Id] = s

	fmt.Printf("%s\n", displaySchema(s, ""))

	classNames := []string{
		"#/definitions/headerParameterSubSchema",
		"#/definitions/formDataParameterSubSchema",
		"#/definitions/queryParameterSubSchema",
		"#/definitions/pathParameterSubSchema"}
	resolveRefs(s, s, classNames)
	resolveRefs(s, s, classNames)
	resolveAllOfs(s, s)
	reduceOneOfs(s, s)
}
