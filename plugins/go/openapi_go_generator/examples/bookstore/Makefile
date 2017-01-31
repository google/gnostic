
build:	
	go install github.com/googleapis/openapi-compiler/openapic
	go install github.com/googleapis/openapi-compiler/plugins/go/openapi_go_generator/encode_templates
	go generate github.com/googleapis/openapi-compiler/plugins/go/openapi_go_generator
	go install github.com/googleapis/openapi-compiler/plugins/go/openapi_go_generator
	rm -f $(GOPATH)/bin/openapi_go_client $(GOPATH)/bin/openapi_go_server
	ln -s $(GOPATH)/bin/openapi_go_generator $(GOPATH)/bin/openapi_go_client
	ln -s $(GOPATH)/bin/openapi_go_generator $(GOPATH)/bin/openapi_go_server

test: 	build
	openapic bookstore.json --go_server_out=- --go_client_out=-

all:	build
	openapic bookstore.json --go_server_out=package=bookstore:bookstore --go_client_out=package=bookstore:bookstore

clean:  
	rm -rf bookstore bookstore.text
