# Bookstore Example

This directory contains an OpenAPI description of a simple bookstore API.

Use this example to try the `openapi_go_generator` plugin, which implements
`openapi_go_client` and `openapi_go_service` for generating API client and
service code, respectively.

Run "make all" to build and install `openapic` and the Go plugins.
It will generate both client and servier code. The API client code
will be in the `bookstore` directory and the service code will be
written to the `service` directory alongside a few pre-written files
(`error.go`, `store.go`, `service.go-completed`). 

To run the service, go to the `service` directory and first copy
`service.go-completed` to `service.go` (the original generated 
`service.go` contains a scaffolding to be completed to implement the
bookstore service).

    cd ..
    cp service.go-completed service.go

Then build and run the service.

    go get .
    go build
    ./service &

To test the service with the generated client, go back up to the `service`
directory and run `go test`. The test in `bookstore_test.go` uses the client
generated in `bookstore` and verifies the service.

