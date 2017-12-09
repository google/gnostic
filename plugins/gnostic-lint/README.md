# gnostic-lint

This directory contains a `gnostic` plugin that analyzes an OpenAPI
description for factors that might influence code generation and other
API automation.

To build the plugin:

  cd <gnostic-root>/plugins/gnostic-lint
	go install

The plugin can be invoked like this:

	gnostic bookstore.json --lint_out=-

This will write analysis results to standard output.

