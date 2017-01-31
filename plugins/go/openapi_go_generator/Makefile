
build:	
	go install github.com/googleapis/openapi-compiler/openapic
	go install github.com/googleapis/openapi-compiler/plugins/go/openapi_go_generator/encode_templates
	go generate github.com/googleapis/openapi-compiler/plugins/go/openapi_go_generator
	go install github.com/googleapis/openapi-compiler/plugins/go/openapi_go_generator
	rm -f $(GOPATH)/bin/openapi_go_client $(GOPATH)/bin/openapi_go_server
	ln -s $(GOPATH)/bin/openapi_go_generator $(GOPATH)/bin/openapi_go_client
	ln -s $(GOPATH)/bin/openapi_go_generator $(GOPATH)/bin/openapi_go_server

