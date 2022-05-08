// Copyright 2020 Google LLC. All Rights Reserved.
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
//

package generator

import (
	"log"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"

	wk "github.com/google/gnostic/cmd/protoc-gen-openapi/generator/wellknown"
	v3 "github.com/google/gnostic/openapiv3"
)

const (
	protobufValueName = "GoogleProtobufValue"
	protobufAnyName   = "GoogleProtobufAny"
)

type OpenAPIv3Reflector struct {
	conf Configuration

	requiredSchemas []string // Names of schemas which are used through references.
}

// NewOpenAPIv3Reflector creates a new reflector.
func NewOpenAPIv3Reflector(conf Configuration) *OpenAPIv3Reflector {
	return &OpenAPIv3Reflector{
		conf: conf,

		requiredSchemas: make([]string, 0),
	}
}

func (r *OpenAPIv3Reflector) getMessageName(message protoreflect.MessageDescriptor) string {
	prefix := ""
	parent := message.Parent()

	if _, ok := parent.(protoreflect.MessageDescriptor); ok {
		prefix = string(parent.Name()) + "_" + prefix
	}

	return prefix + string(message.Name())
}

func (r *OpenAPIv3Reflector) formatMessageName(message protoreflect.MessageDescriptor) string {
	typeName := r.fullMessageTypeName(message)

	name := r.getMessageName(message)
	if !*r.conf.FQSchemaNaming {
		if typeName == ".google.protobuf.Value" {
			name = protobufValueName
		} else if typeName == ".google.protobuf.Any" {
			name = protobufAnyName
		}
	}

	if *r.conf.Naming == "json" {
		if len(name) > 1 {
			name = strings.ToUpper(name[0:1]) + name[1:]
		}

		if len(name) == 1 {
			name = strings.ToLower(name)
		}
	}

	if *r.conf.FQSchemaNaming {
		package_name := string(message.ParentFile().Package())
		name = package_name + "." + name
	}

	return name
}

func (r *OpenAPIv3Reflector) formatFieldName(field protoreflect.FieldDescriptor) string {
	if *r.conf.Naming == "proto" {
		return string(field.Name())
	}

	return field.JSONName()
}

// fullMessageTypeName builds the full type name of a message.
func (r *OpenAPIv3Reflector) fullMessageTypeName(message protoreflect.MessageDescriptor) string {
	name := r.getMessageName(message)
	return "." + string(message.ParentFile().Package()) + "." + name
}

func (r *OpenAPIv3Reflector) responseContentForMessage(message protoreflect.MessageDescriptor) (string, *v3.MediaTypes) {
	typeName := r.fullMessageTypeName(message)

	if typeName == ".google.protobuf.Empty" {
		return "200", &v3.MediaTypes{}
	}

	if typeName == ".google.api.HttpBody" {
		return "200", wk.NewGoogleApiHttpBodyMediaType()
	}

	return "200", wk.NewApplicationJsonMediaType(r.schemaOrReferenceForMessage(message))
}

func (r *OpenAPIv3Reflector) schemaReferenceForMessage(message protoreflect.MessageDescriptor) string {
	schemaName := r.formatMessageName(message)
	if !contains(r.requiredSchemas, schemaName) {
		r.requiredSchemas = append(r.requiredSchemas, schemaName)
	}
	return "#/components/schemas/" + schemaName
}

// Returns a full schema for simple types, and a schema reference for complex types that reference
// the definition in `#/components/schemas/`
func (r *OpenAPIv3Reflector) schemaOrReferenceForMessage(message protoreflect.MessageDescriptor) *v3.SchemaOrReference {
	typeName := r.fullMessageTypeName(message)
	switch typeName {

	case ".google.api.HttpBody":
		return wk.NewGoogleApiHttpBodySchema()

	case ".google.protobuf.Timestamp":
		return wk.NewGoogleProtobufTimestampSchema()

	case ".google.type.Date":
		return wk.NewGoogleTypeDateSchema()

	case ".google.type.DateTime":
		return wk.NewGoogleTypeDateTimeSchema()

	case ".google.protobuf.FieldMask":
		return wk.NewGoogleProtobufFieldMaskSchema()

	case ".google.protobuf.Struct":
		return wk.NewGoogleProtobufStructSchema()

	case ".google.protobuf.Empty":
		// Empty is closer to JSON undefined than null, so ignore this field
		return nil //&v3.SchemaOrReference{Oneof: &v3.SchemaOrReference_Schema{Schema: &v3.Schema{Type: "null"}}}

	default:
		ref := r.schemaReferenceForMessage(message)
		return &v3.SchemaOrReference{
			Oneof: &v3.SchemaOrReference_Reference{
				Reference: &v3.Reference{XRef: ref}}}
	}
}

func (r *OpenAPIv3Reflector) schemaOrReferenceForField(field protoreflect.FieldDescriptor) *v3.SchemaOrReference {
	var kindSchema *v3.SchemaOrReference

	kind := field.Kind()

	switch kind {

	case protoreflect.MessageKind:
		if field.IsMap() {
			// This means the field is a map, for example:
			//   map<string, value_type> map_field = 1;
			//
			// The map ends up getting converted into something like this:
			//   message MapFieldEntry {
			//     string key = 1;
			//     value_type value = 2;
			//   }
			//
			//   repeated MapFieldEntry map_field = N;
			//
			// So we need to find the `value` field in the `MapFieldEntry` message and
			// then return a MapFieldEntry schema using the schema for the `value` field
			return wk.NewGoogleProtobufMapFieldEntrySchema(r.schemaOrReferenceForField(field.MapValue()))
		} else {
			kindSchema = r.schemaOrReferenceForMessage(field.Message())
		}

	case protoreflect.StringKind:
		kindSchema = wk.NewStringSchema()

	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Uint32Kind,
		protoreflect.Sfixed32Kind, protoreflect.Fixed32Kind:
		kindSchema = wk.NewIntegerSchema(kind.String())

	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Uint64Kind,
		protoreflect.Sfixed64Kind, protoreflect.Fixed64Kind:
		kindSchema = wk.NewStringSchema()

	case protoreflect.EnumKind:
		kindSchema = wk.NewEnumSchema(*&r.conf.EnumType, field)

	case protoreflect.BoolKind:
		kindSchema = wk.NewBooleanSchema()

	case protoreflect.FloatKind, protoreflect.DoubleKind:
		kindSchema = wk.NewNumberSchema(kind.String())

	case protoreflect.BytesKind:
		kindSchema = wk.NewBytesSchema()

	default:
		log.Printf("(TODO) Unsupported field type: %+v", r.fullMessageTypeName(field.Message()))
	}

	if field.IsList() {
		kindSchema = wk.NewListSchema(kindSchema)
	}

	return kindSchema
}
