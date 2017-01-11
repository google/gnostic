# Service Generator

This repository contains a Go program that can be used to generate
a Go-language app for Google App Engine from an OpenAPI model.
The scaffolding that it generates allows services to be implemented
with minimal knowledge of Google App Engine and HTTP processing.

In addition to Google App Engine, generated services can be easily 
run on any hosting platform that supports Go.

## Requirements

Service Generator can be run in any environment that supports Go
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

2. Run the OpenAPI Compiler and use the plugin to generate the app.

	`openapi-compiler bookstore.json --go_service_out=app`

3. Switch to the app directory.

	`cd app`

4. Get dependencies.

	`goapp get`

5. Run the app locally.

	`goapp serve`

Note that `service.go` is a manually-edited version of generated `service.go-finishme`.
Each new service that is generated will require customization of this file.
