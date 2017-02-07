
all:
	swift build
	.build/debug/TemplateEncoder > Sources/openapi_swift_generator/Templates.swift
	swift build
	cp .build/debug/openapi_swift_generator openapi_swift_generator
	rm -f openapi_swift_client openapi_swift_server
	ln -s openapi_swift_generator openapi_swift_client
	ln -s openapi_swift_generator openapi_swift_server

clean:
	rm -rf .build Packages
	rm -rf openapi_swift_client openapi_swift_service
