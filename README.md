# protoc-gen-openapi

This directory contains a protoc plugin that generates an
OpenAPI description for a REST API that corresponds to a
Protocol Buffer service.

Installation:

    go install github.com/google/gnostic/cmd/protoc-gen-openapi

Usage:

	protoc sample.proto -I=. --openapi_out=.

This runs the plugin for a file named `sample.proto` which 
refers to additional .proto files in the same directory as
`sample.proto`. Output is written to the current directory.

## options

1. `version`: version number text, e.g. 1.2.3
   - **default**: `0.0.1`
2. `title`: name of the API
   - **default**: empty string or service name if there is only one service
3. `description`: description of the API
   - **default**: empty string or service description if there is only one service
4. `naming`: naming convention. Use "proto" for passing names directly from the proto files
   - **default**: `json`
   - `json`: will turn field `updated_at` to `updatedAt`
   - `proto`: keep field `updated_at` as it is
5. `fq_schema_naming`: schema naming convention. If "true", generates fully-qualified schema names by prefixing them with the proto message package name
   - **default**: false
   - `false`: keep message `Book` as it is
   - `true`: turn message `Book` to `google.example.library.v1.Book`, it is useful when there are same named message in different package
6. `enum_type`: type for enum serialization. Use "string" for string-based serialization
   - **default**: `integer`
   - `integer`: setting type to `integer`
      ```yaml
      schema:
        type: integer
        format: enum
      ```
   - `string`: setting type to `string`, and list available values in `enum`
      ```yaml
      schema:
        enum:
          - UNKNOWN_KIND
          - KIND_1
          - KIND_2
        type: string
        format: enum
      ```
7. `depth`: depth of recursion for circular messages
   - **default**: 2, this depth only used in query parameters, usually 2 is enough
8. `default_response`: add default response. If "true", automatically adds a default response to operations which use the google.rpc.Status message.
   Useful if you use envoy or grpc-gateway to transcode as they use this type for their default error responses.
   - **default**: true, this option will add this default response for each method as following:
      ```yaml
      default:
        description: Default error response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/google.rpc.Status'
      ```