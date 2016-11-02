# OpenAPI Compiler

This repository contains an experimental project whose goal is to 
read OpenAPI JSON descriptions and write equivalent Protocol Buffer
representations. These Protocol Buffer representations would be 
preprocessed, checked for errors, and available for use in any
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

	`go get github.com/googleapis/openapi-compiler`

2. Generate protocol buffer support code.

	`go generate github.com/googleapis/openapi-compiler`

3. Build and install the OpenAPI compiler.

	`go install github.com/googleapis/openapi-compiler`

## Copyright

Copyright 2016, Google Inc.

## License

Released under the Apache 2.0 license.

