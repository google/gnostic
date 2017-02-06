# OpenAPI Swift Generator Plugin

This directory contains an `openapic` plugin that can be used to generate a Swift client library and scaffolding for a Swift server for an API with an OpenAPI description.

The plugin can be invoked like this:

	openapic bookstore.json --swift_generator_out=Bookstore

Where `Bookstore` is the name of a directory where the generated code will be written.

By default, both client and server code will be generated. If the `openapi_swift_generator` binary is also linked from the names `openapi_swift_client` and `openapi_swift_server`, then only client or only server code can be generated as follows:

	openapic bookstore.json --swift_client_out=package=bookstore:bookstore

	openapic bookstore.json --swift_server_out=package=bookstore:bookstore

For example usage, see the [examples/bookstore](examples/bookstore) directory.

## Notes

- When debugging plugins, remember that plugin output is written to stdout, so any print statements that you add must print to stderr to avoid breaking the plugin interface..
