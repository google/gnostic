module github.com/pubgo/protoc-gen-openapi

go 1.12

replace github.com/google/gnostic => github.com/google/gnostic v0.7.0

require (
	github.com/google/gnostic v0.0.0-00010101000000-000000000000
	google.golang.org/genproto v0.0.0-20240521202816-d264139d666e // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240521202816-d264139d666e
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240515191416-fc5f0ca64291
	google.golang.org/protobuf v1.34.1
)
