
build:	
	go install github.com/googleapis/openapi-compiler/plugins/go/template_compiler
	go install github.com/googleapis/openapi-compiler/openapic
	go generate github.com/googleapis/openapi-compiler/plugins/go/openapi_go_service
	go install github.com/googleapis/openapi-compiler/plugins/go/openapi_go_service
	go generate github.com/googleapis/openapi-compiler/plugins/go/openapi_go_client
	go install github.com/googleapis/openapi-compiler/plugins/go/openapi_go_client

test: 	build
	openapic swagger.json --go_service_out=- --go_client_out=-

service:build
	openapic swagger.json --go_service_out=package=main:service 

client:	build
	openapic swagger.json --go_client_out=package=bookstore:bookstore

all:	build
	openapic swagger.json --go_service_out=package=main:service --go_client_out=package=bookstore:bookstore

clean:  
	rm -rf bookstore service/app.go service/app.yaml service/service service/service.go service/types.go swagger.text
