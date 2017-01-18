
build:	
	go install github.com/googleapis/openapi-compiler/openapic
	go install github.com/googleapis/openapi-compiler/plugins/go/openapi_go_service

test: 	build
	openapic swagger.json --go_service_out=-
