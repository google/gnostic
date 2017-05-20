
build:	
	go get
	go install
	cd generator; go get; go install
	cd apps/report; go get; go install
	cd plugins/go/gnostic_go_sample; go get; go install
	cd plugins/go/gnostic_go_generator/encode_templates; go get; go install
	cd plugins/go/gnostic_go_generator; go get; go install
	rm -f $(GOPATH)/bin/gnostic_go_client $(GOPATH)/bin/gnostic_go_server
	ln -s $(GOPATH)/bin/gnostic_go_generator $(GOPATH)/bin/gnostic_go_client
	ln -s $(GOPATH)/bin/gnostic_go_generator $(GOPATH)/bin/gnostic_go_server
	cd extensions/sample; make

