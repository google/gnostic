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

package jsonschema

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
	Properties           *[]*NamedSchema
	PatternProperties    *[]*NamedSchema
	Dependencies         *[]*NamedSchemaOrStringArray

	// 5.5.  Validation keywords for any instance type
	Enumeration *[]SchemaEnumValue
	Type        *StringOrStringArray
	AllOf       *[]*Schema
	AnyOf       *[]*Schema
	OneOf       *[]*Schema
	Not         *Schema
	Definitions *[]*NamedSchema

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
	String      *string
	StringArray *[]string
}

type SchemaOrStringArray struct {
	Schema      *Schema
	StringArray *[]string
}

type SchemaOrSchemaArray struct {
	Schema      *Schema
	SchemaArray *[]*Schema
}

type SchemaEnumValue struct {
	String *string
	Bool   *bool
}

// These structs provide key-value pairs that are kept in slices.
// They are used to emulate maps with ordered keys.
type NamedSchema struct {
	Name  string
	Value *Schema
}

type NamedSchemaOrStringArray struct {
	Name  string
	Value *SchemaOrStringArray
}
