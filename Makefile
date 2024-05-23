test:
	# since some tests call separately-built binaries, clear the cache to ensure all get run
	go clean -testcache
	go test ./... -v

vet:
	go vet ./...

pb:
	protobuild vendor

pb_test:
	protobuild -c protobuf_test.yaml vendor
	protobuild -c protobuf_test.yaml gen

install_gnostic:
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@v0.7.0
