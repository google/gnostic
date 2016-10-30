# OpenAPI Compiler

This repository contains a Go program that reads an OpenAPI JSON
description and writes an equivalent Protocol Buffer representation.

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

