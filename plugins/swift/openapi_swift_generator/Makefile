
build:
	swift build
	.build/debug/TemplateEncoder > Sources/openapi_swift_generator/Templates.swift
	swift build
	cp .build/debug/openapi_swift_generator openapi_swift_generator
	openapic examples/bookstore/bookstore.json \
		--swift_generator_out=examples/bookstore/Sources/Bookstore 


clean:
	rm -rf .build Packages
	rm -rf openapi_swift_client openapi_swift_service
