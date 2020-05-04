This directory contains a demonstration gnostic application written in Dart.

Build instructions:

- Be sure that [Dart](dart.dev/get-dart) and [protoc](https://github.com/protocolbuffers/protobuf/releases) are installed in your development environment.

- Install the Dart protobuf plugin.
	pub global activate protoc_plugin
 
- Be sure that your `PATH` includes `$HOME:/.pub-cache/bin`.

- Install protobuf dependencies.
	cd third_party; ./SETUP.sh

- Compile protobuf source files. Note that a variable in this script points to a directory of `.proto` files in your installed `protoc` release.
	./COMPILE-PROTOS.sh

- Fetch Dart dependencies.
	pub get

- Run the Dart program.
	dart bin/summarize.dart <filename>

Here `<filename>` should be a binary-encoded protobuf representation of an OpenAPI description.
