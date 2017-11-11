# gnostic-lint

This directory contains a `gnostic` plugin that analyzes an OpenAPI
description for factors that might influence code generation and other
API automation.

The plugin can be invoked like this:

	gnostic bookstore.json --linter_out=.

This will write analysis results to a file in the current directory.

