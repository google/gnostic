
all:
	# TODO brendon specific path here
	PATH=/Users/brendon/go/bin:$$PATH ./COMPILE-PROTOS.sh
	cd ./cmd/protoc-gen-openapi && go build && cd ../..

