module github.com/google/gnostic/cmd/protoc-gen-openapi

go 1.12

require (
	github.com/google/gnostic v0.0.0
	google.golang.org/genproto v0.0.0-20211223182754-3ac035c7e7cb
	google.golang.org/protobuf v1.27.1
)

replace github.com/google/gnostic v0.0.0 => ../..
