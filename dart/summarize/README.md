This directory contains a demonstration gnostic application written in Dart.

Build instructions:

- install protobuf dependencies
	cd third_party; ./SETUP.sh

- compile protobuf source files
	./COMPILE-PROTOS.sh

- run a Dart program
	dart bin/summarize.dart <filename>

Here <filename> should be a binary-encoded protobuf representation of an OpenAPI description.
