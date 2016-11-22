# OpenAPI Compiler

This repository contains an experimental project whose goal is to
read OpenAPI JSON descriptions and write equivalent Protocol Buffer
representations. These Protocol Buffer representations are to be
preprocessed, checked for errors, and made available for use in any
language that is supported by the Protocol Buffer tools.

## Disclaimer

This is prerelease software and work in progress. Feedback and
contributions are welcome, but we currently make no guarantees of
function or stability.

## Requirements

OpenAPI Compiler can be run in any environment that supports Go
and the [Google Protocol Buffer Compiler](https://github.com/google/protobuf).

## Installation

1. Get this package by downloading it with `go get` or manually cloning it into `go/src`.

        go get github.com/googleapis/openapi-compiler
	
2. [Optional] Build and run the compiler generator. 
This uses the OpenAPI JSON schema to generate a Protocol Buffer language file 
that describes the OpenAPI specification and a Go-language file of code that 
will read a JSON or YAML OpenAPI representation into the generated protocol 
buffers. Precompiled versions of these files are in the OpenAPIv2 directory.

        cd $GOPATH/src/github.com/googleapis/openapi-compiler/generator
        go build generator.go
        cd ..
        ./generator/generator

3. Generate protocol buffer support code. You'll find the generated 
protocol buffer code at `$GOPATH/src/openapi`.

        go generate github.com/googleapis/openapi-compiler

4. Build and install the OpenAPI compiler. 

        go install github.com/googleapis/openapi-compiler

5. Run the OpenAPI compiler.

        openapi-compiler -input examples/petstore.json -pb

This will create a file called "petstore.pb" that contains a binary
Protocol Buffer description of a sample API.

6. For a sample application, see apps/report.

## Copyright

Copyright 2016, Google Inc.

## License

Released under the Apache 2.0 license.
