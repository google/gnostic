
build:	
	go install github.com/googleapis/openapi-compiler/plugins/go/openapi_go_generator/template_compiler
	pushd ../../openapi_go_generator; template_compiler; popd
	go install github.com/googleapis/openapi-compiler/openapic
	go install github.com/googleapis/openapi-compiler/plugins/go/openapi_go_generator
	rm $(GOPATH)/bin/openapi_go_client $(GOPATH)/bin/openapi_go_service
	ln -s $(GOPATH)/bin/openapi_go_generator $(GOPATH)/bin/openapi_go_client
	ln -s $(GOPATH)/bin/openapi_go_generator $(GOPATH)/bin/openapi_go_service

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
