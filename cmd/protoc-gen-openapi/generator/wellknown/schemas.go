// Copyright 2020 Google LLC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, softwis
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package wellknown

import (
	v3 "github.com/google/gnostic/openapiv3"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func NewStringSchema() *v3.SchemaOrReference {
	return &v3.SchemaOrReference{
		Oneof: &v3.SchemaOrReference_Schema{
			Schema: &v3.Schema{Type: "string"}}}
}

func NewBooleanSchema() *v3.SchemaOrReference {
	return &v3.SchemaOrReference{
		Oneof: &v3.SchemaOrReference_Schema{
			Schema: &v3.Schema{Type: "boolean"}}}
}

func NewBytesSchema() *v3.SchemaOrReference {
	return &v3.SchemaOrReference{
		Oneof: &v3.SchemaOrReference_Schema{
			Schema: &v3.Schema{Type: "string", Format: "bytes"}}}
}

func NewIntegerSchema(format string) *v3.SchemaOrReference {
	return &v3.SchemaOrReference{
		Oneof: &v3.SchemaOrReference_Schema{
			Schema: &v3.Schema{Type: "integer", Format: format}}}
}

func NewNumberSchema(format string) *v3.SchemaOrReference {
	return &v3.SchemaOrReference{
		Oneof: &v3.SchemaOrReference_Schema{
			Schema: &v3.Schema{Type: "number", Format: format}}}
}

func NewEnumSchema(enum_type *string, field protoreflect.FieldDescriptor) *v3.SchemaOrReference {
	schema := &v3.Schema{Format: "enum"}
	if enum_type != nil && *enum_type == "string" {
		schema.Type = "string"
		schema.Enum = make([]*v3.Any, 0, field.Enum().Values().Len())
		for i := 0; i < field.Enum().Values().Len(); i++ {
			schema.Enum = append(schema.Enum, &v3.Any{
				Yaml: string(field.Enum().Values().Get(i).Name()),
			})
		}
	} else {
		schema.Type = "integer"
	}
	return &v3.SchemaOrReference{
		Oneof: &v3.SchemaOrReference_Schema{
			Schema: schema}}
}

func NewListSchema(item_schema *v3.SchemaOrReference) *v3.SchemaOrReference {
	return &v3.SchemaOrReference{
		Oneof: &v3.SchemaOrReference_Schema{
			Schema: &v3.Schema{
				Type:  "array",
				Items: &v3.ItemsItem{SchemaOrReference: []*v3.SchemaOrReference{item_schema}},
			},
		},
	}
}

// google.api.HttpBody will contain POST body data
// This is based on how Envoy handles google.api.HttpBody
func NewGoogleApiHttpBodySchema() *v3.SchemaOrReference {
	return &v3.SchemaOrReference{
		Oneof: &v3.SchemaOrReference_Schema{
			Schema: &v3.Schema{Type: "string"}}}
}

// google.protobuf.Timestamp is serialized as a string
func NewGoogleProtobufTimestampSchema() *v3.SchemaOrReference {
	return &v3.SchemaOrReference{
		Oneof: &v3.SchemaOrReference_Schema{
			Schema: &v3.Schema{Type: "string", Format: "date-time"}}}
}

// google.type.Date is serialized as a string
func NewGoogleTypeDateSchema() *v3.SchemaOrReference {
	return &v3.SchemaOrReference{
		Oneof: &v3.SchemaOrReference_Schema{
			Schema: &v3.Schema{Type: "string", Format: "date"}}}
}

// google.type.DateTime is serialized as a string
func NewGoogleTypeDateTimeSchema() *v3.SchemaOrReference {
	return &v3.SchemaOrReference{
		Oneof: &v3.SchemaOrReference_Schema{
			Schema: &v3.Schema{Type: "string", Format: "date-time"}}}
}

// google.protobuf.FieldMask masks is serialized as a string
func NewGoogleProtobufFieldMaskSchema() *v3.SchemaOrReference {
	return &v3.SchemaOrReference{
		Oneof: &v3.SchemaOrReference_Schema{
			Schema: &v3.Schema{Type: "string", Format: "field-mask"}}}
}

// google.protobuf.Struct is equivalent to a JSON object
func NewGoogleProtobufStructSchema() *v3.SchemaOrReference {
	return &v3.SchemaOrReference{
		Oneof: &v3.SchemaOrReference_Schema{
			Schema: &v3.Schema{Type: "object"}}}
}

// google.protobuf.Value is handled specially
// See here for the details on the JSON mapping:
//   https://developers.google.com/protocol-buffers/docs/proto3#json
// and here:
//   https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#google.protobuf.Value
func NewGoogleProtobufValueSchema(name string) *v3.NamedSchemaOrReference {
	return &v3.NamedSchemaOrReference{
		Name: name,
		Value: &v3.SchemaOrReference{
			Oneof: &v3.SchemaOrReference_Schema{
				Schema: &v3.Schema{
					Description: "Represents a dynamically typed value which can be either null, a number, a string, a boolean, a recursive struct value, or a list of values.",
				},
			},
		},
	}
}

// google.protobuf.Any is handled specially
// See here for the details on the JSON mapping:
//   https://developers.google.com/protocol-buffers/docs/proto3#json
func NewGoogleProtobufAnySchema(name string) *v3.NamedSchemaOrReference {
	return &v3.NamedSchemaOrReference{
		Name: name,
		Value: &v3.SchemaOrReference{
			Oneof: &v3.SchemaOrReference_Schema{
				Schema: &v3.Schema{
					Type:        "object",
					Description: "Contains an arbitrary serialized message along with a @type that describes the type of the serialized message.",
					Properties: &v3.Properties{
						AdditionalProperties: []*v3.NamedSchemaOrReference{
							{
								Name: "@type",
								Value: &v3.SchemaOrReference{
									Oneof: &v3.SchemaOrReference_Schema{
										Schema: &v3.Schema{
											Type:        "string",
											Description: "The type of the serialized message.",
										},
									},
								},
							},
						},
					},
					AdditionalProperties: &v3.AdditionalPropertiesItem{
						Oneof: &v3.AdditionalPropertiesItem_Boolean{
							Boolean: true,
						},
					},
				},
			},
		},
	}
}

// google.rpc.Status is handled specially
func NewGoogleRpcStatusSchema(name string, any_name string) *v3.NamedSchemaOrReference {
	return &v3.NamedSchemaOrReference{
		Name: name,
		Value: &v3.SchemaOrReference{
			Oneof: &v3.SchemaOrReference_Schema{
				Schema: &v3.Schema{
					Type:        "object",
					Description: "The `Status` type defines a logical error model that is suitable for different programming environments, including REST APIs and RPC APIs. It is used by [gRPC](https://github.com/grpc). Each `Status` message contains three pieces of data: error code, error message, and error details. You can find out more about this error model and how to work with it in the [API Design Guide](https://cloud.google.com/apis/design/errors).",
					Properties: &v3.Properties{
						AdditionalProperties: []*v3.NamedSchemaOrReference{
							{
								Name: "code",
								Value: &v3.SchemaOrReference{
									Oneof: &v3.SchemaOrReference_Schema{
										Schema: &v3.Schema{
											Type:        "integer",
											Format:      "int32",
											Description: "The status code, which should be an enum value of [google.rpc.Code][google.rpc.Code].",
										},
									},
								},
							},
							{
								Name: "message",
								Value: &v3.SchemaOrReference{
									Oneof: &v3.SchemaOrReference_Schema{
										Schema: &v3.Schema{
											Type:        "string",
											Description: "A developer-facing error message, which should be in English. Any user-facing error message should be localized and sent in the [google.rpc.Status.details][google.rpc.Status.details] field, or localized by the client.",
										},
									},
								},
							},
							{
								Name: "details",
								Value: &v3.SchemaOrReference{
									Oneof: &v3.SchemaOrReference_Schema{
										Schema: &v3.Schema{
											Type: "array",
											Items: &v3.ItemsItem{
												SchemaOrReference: []*v3.SchemaOrReference{
													{
														Oneof: &v3.SchemaOrReference_Reference{
															Reference: &v3.Reference{
																XRef: "#/components/schemas/" + any_name,
															},
														},
													},
												},
											},
											Description: "A list of messages that carry the error details.  There is a common set of message types for APIs to use.",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func NewGoogleProtobufMapFieldEntrySchema(value_field_schema *v3.SchemaOrReference) *v3.SchemaOrReference {
	return &v3.SchemaOrReference{
		Oneof: &v3.SchemaOrReference_Schema{
			Schema: &v3.Schema{Type: "object",
				AdditionalProperties: &v3.AdditionalPropertiesItem{
					Oneof: &v3.AdditionalPropertiesItem_SchemaOrReference{
						SchemaOrReference: value_field_schema,
					},
				},
			},
		},
	}
}
