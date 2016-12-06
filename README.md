[![Build Status](https://travis-ci.org/googleapis/openapi-compiler.svg?branch=master)](https://travis-ci.org/googleapis/openapi-compiler)

# OpenAPI Compiler

This repository contains an experimental project whose goal is to
read OpenAPI specifications in JSON or YAML formats and write 
equivalent Protocol Buffer representations. 
These Protocol Buffer representations are to be
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
buffers. Pre-generated versions of these files are in the OpenAPIv2 directory.

        cd $GOPATH/src/github.com/googleapis/openapi-compiler/generator
        go build generator.go
        cd ..
        ./generator/generator

3. [Optional] Generate protocol buffer support code. 
A pre-generated version of this file is checked into the OpenAPIv2 directory.
This step requires a local installation of protoc, the Protocol Buffer Compiler.
You can get protoc [here](https://github.com/google/protobuf).

        go generate github.com/googleapis/openapi-compiler

4. [Optional] Rebuild openapi-compiler. This is only necessary if you've performed steps
2 or 3 above.

        go install github.com/googleapis/openapi-compiler

5. Run the OpenAPI compiler. This will create a file called "petstore.pb" that contains a binary
Protocol Buffer description of a sample API.

        openapi-compiler -input examples/petstore.json -pb

6. For a sample application, see apps/report.

        go install github.com/googleapis/openapi-compiler/apps/report
		report petstore.pb

## Copyright

Copyright 2016, Google Inc.

## License

Released under the Apache 2.0 license.
