
build:
	swift build
	cp .build/debug/openapi_swift_generator openapi_swift_client
	cp .build/debug/openapi_swift_generator openapi_swift_service
	openapic examples/bookstore/bookstore.json \
		--swift_client_out=examples/bookstore/client \
		--swift_service_out=examples/bookstore/service

