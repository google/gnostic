# Client Generator

This repository contains a Go program that can be used to generate
a Go-language client for an API from its OpenAPI model.

## Requirements

Client Generator can be run in any environment that supports Go
and the [Google Protocol Buffer Compiler](https://github.com/google/protobuf).

## Installation

1. Build and install the OpenAPI Compiler.

	``

2. Generate protocol buffer support code.

	``

3. Build and install this plugin.

	``

## Usage

1. Go to the examples/bookstore directory.

	`cd examples/bookstore`

2. Run the OpenAPI Compiler and use the plugin to generate the client.

	`openapi-compiler bookstore.json --go_client_out=client`

