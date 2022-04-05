# protoc-gen-openapi

This directory contains a protoc plugin that generates an
OpenAPI description for a REST API that corresponds to a
Protocol Buffer service.

Installation:

    go install github.com/google/gnostic/cmd/protoc-gen-openapi
  
  
Usage:

	protoc sample.proto -I. --openapi_out=.

This runs the plugin for a file named `sample.proto` which 
refers to additional .proto files in the same directory as
`sample.proto`. Output is written to the current directory.

## options

1. `version`: version number text, e.g. 1.2.3
   - default: `0.0.1`
2. `title`: name of the API
   - default: empty string or service name if there is only one service
3. `description`: description of the API
   - default: empty string or service description if there is only one service
4. `naming`: naming convention. Use "proto" for passing names directly from the proto files
   - default: `json`
   - `json`: will turn field `updated_at` to `updatedAt`
   - `proto`: keep field `updated_at` as it is
5. `fq_schema_naming`: schema naming convention. If "true", generates fully-qualified schema names by prefixing them with the proto message package name
   - default: false
   - `false`: keep message `Book` as it is
   - `true`: turn message `Book` to `google.example.library.v1.Book`, it is useful when there are same named message in different package