# protoc-gen-jsonschema

This directory contains a protoc plugin that generates
JSON Schemas for Protocol Buffer messages.

Installation:

        go install github.com/google/gnostic/cmd/protoc-gen-jsonschema
  
  
Usage:

	protoc sample.proto -I. --jsonschema_out=.

This runs the plugin for a file named `sample.proto` which 
refers to additional .proto files in the same directory as
`sample.proto`. Output is written to the current directory.

