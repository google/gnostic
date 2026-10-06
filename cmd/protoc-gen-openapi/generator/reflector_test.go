// Copyright 2026 Google LLC. All Rights Reserved.
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

package generator

import (
	"testing"

	v3 "github.com/google/gnostic/openapiv3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestSchemaOrReferenceForFieldFormats(t *testing.T) {
	tests := []struct {
		name       string
		kind       descriptorpb.FieldDescriptorProto_Type
		typeName   string
		wantType   string
		wantFormat string
	}{
		{name: "int64", kind: descriptorpb.FieldDescriptorProto_TYPE_INT64, wantType: "string", wantFormat: "int64"},
		{name: "uint64", kind: descriptorpb.FieldDescriptorProto_TYPE_UINT64, wantType: "string", wantFormat: "uint64"},
		{name: "sint64", kind: descriptorpb.FieldDescriptorProto_TYPE_SINT64, wantType: "string", wantFormat: "sint64"},
		{name: "fixed64", kind: descriptorpb.FieldDescriptorProto_TYPE_FIXED64, wantType: "string", wantFormat: "fixed64"},
		{name: "sfixed64", kind: descriptorpb.FieldDescriptorProto_TYPE_SFIXED64, wantType: "string", wantFormat: "sfixed64"},
		{name: "string", kind: descriptorpb.FieldDescriptorProto_TYPE_STRING, wantType: "string"},
		{name: "int32", kind: descriptorpb.FieldDescriptorProto_TYPE_INT32, wantType: "integer", wantFormat: "int32"},
		{name: "uint32", kind: descriptorpb.FieldDescriptorProto_TYPE_UINT32, wantType: "integer", wantFormat: "uint32"},
		{name: "Int64Value", kind: descriptorpb.FieldDescriptorProto_TYPE_MESSAGE, typeName: ".google.protobuf.Int64Value", wantType: "string", wantFormat: "int64"},
		{name: "UInt64Value", kind: descriptorpb.FieldDescriptorProto_TYPE_MESSAGE, typeName: ".google.protobuf.UInt64Value", wantType: "string", wantFormat: "uint64"},
		{name: "StringValue", kind: descriptorpb.FieldDescriptorProto_TYPE_MESSAGE, typeName: ".google.protobuf.StringValue", wantType: "string"},
		{name: "Int32Value", kind: descriptorpb.FieldDescriptorProto_TYPE_MESSAGE, typeName: ".google.protobuf.Int32Value", wantType: "integer", wantFormat: "int32"},
		{name: "UInt32Value", kind: descriptorpb.FieldDescriptorProto_TYPE_MESSAGE, typeName: ".google.protobuf.UInt32Value", wantType: "integer", wantFormat: "uint32"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := &descriptorpb.FieldDescriptorProto{
				Name:   proto.String("value"),
				Number: proto.Int32(1),
				Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
				Type:   tt.kind.Enum(),
			}
			if tt.typeName != "" {
				value.TypeName = proto.String(tt.typeName)
			}
			values := proto.Clone(value).(*descriptorpb.FieldDescriptorProto)
			values.Name = proto.String("values")
			values.Number = proto.Int32(2)
			values.Label = descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum()
			mapValue := proto.Clone(value).(*descriptorpb.FieldDescriptorProto)
			mapValue.Number = proto.Int32(2)

			file, err := protodesc.NewFile(&descriptorpb.FileDescriptorProto{
				Name:       proto.String("test.proto"),
				Package:    proto.String("test"),
				Syntax:     proto.String("proto3"),
				Dependency: []string{wrapperspb.File_google_protobuf_wrappers_proto.Path()},
				MessageType: []*descriptorpb.DescriptorProto{{
					Name: proto.String("Message"),
					Field: []*descriptorpb.FieldDescriptorProto{value, values, {
						Name:     proto.String("values_by_key"),
						Number:   proto.Int32(3),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: proto.String(".test.Message.ValuesByKeyEntry"),
					}},
					NestedType: []*descriptorpb.DescriptorProto{{
						Name:    proto.String("ValuesByKeyEntry"),
						Options: &descriptorpb.MessageOptions{MapEntry: proto.Bool(true)},
						Field: []*descriptorpb.FieldDescriptorProto{{
							Name:   proto.String("key"),
							Number: proto.Int32(1),
							Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
							Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						}, mapValue},
					}},
				}},
			}, protoregistry.GlobalFiles)
			if err != nil {
				t.Fatal(err)
			}

			r := NewOpenAPIv3Reflector(Configuration{})
			fields := file.Messages().Get(0).Fields()
			for i := 0; i < fields.Len(); i++ {
				field := fields.Get(i)
				t.Run(string(field.Name()), func(t *testing.T) {
					schema := r.schemaOrReferenceForField(field).GetSchema()
					switch {
					case field.IsMap():
						if schema.GetType() != "object" {
							t.Fatalf("map schema type = %q, want object", schema.GetType())
						}
						schema = schema.GetAdditionalProperties().GetSchemaOrReference().GetSchema()
					case field.IsList():
						items := schema.GetItems().GetSchemaOrReference()
						if schema.GetType() != "array" || len(items) != 1 {
							t.Fatalf("want array schema with one item schema, got %v", schema)
						}
						schema = items[0].GetSchema()
					}
					want := &v3.Schema{Type: tt.wantType, Format: tt.wantFormat}
					if !proto.Equal(schema, want) {
						t.Errorf("schema = %v, want %v", schema, want)
					}
				})
			}
		})
	}
}
