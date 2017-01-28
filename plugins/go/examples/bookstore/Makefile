
build:	
	go install github.com/googleapis/openapi-compiler/openapic
	go install github.com/googleapis/openapi-compiler/plugins/go/openapi_go_service
	go install github.com/googleapis/openapi-compiler/plugins/go/openapi_go_client

test: 	build
	openapic swagger.json --go_service_out=- --go_client_out=-

service:build
	openapic swagger.json --go_service_out=package=main:service 

client:	build
	openapic swagger.json --go_client_out=package=bookstore:bookstore
