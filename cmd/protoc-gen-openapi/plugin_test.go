// Copyright 2020 Google LLC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"os"
	"os/exec"
	"path"
	"testing"
)

var openapiTests = []struct {
	name      string
	path      string
	protofile string
}{
	{name: "Google Library example", path: "examples/google/example/library/v1/", protofile: "library.proto"},
	{name: "Body mapping", path: "examples/tests/bodymapping/", protofile: "message.proto"},
	{name: "Map fields", path: "examples/tests/mapfields/", protofile: "message.proto"},
	{name: "Path params", path: "examples/tests/pathparams/", protofile: "message.proto"},
	{name: "Protobuf types", path: "examples/tests/protobuftypes/", protofile: "message.proto"},
	{name: "JSON options", path: "examples/tests/jsonoptions/", protofile: "message.proto"},
	{name: "Ignore services without annotaions", path: "examples/tests/noannotations/", protofile: "message.proto"},
}

func TestOpenAPIProtobufNaming(t *testing.T) {
	for _, tt := range openapiTests {
		t.Run(tt.name, func(t *testing.T) {
			// Run protoc and the protoc-gen-openapi plugin to generate an OpenAPI spec.
			err := exec.Command("protoc",
				"-I", "../../",
				"-I", "../../third_party",
				"-I", "examples",
				path.Join(tt.path, tt.protofile),
				"--openapi_out=naming=proto:.").Run()
			if err != nil {
				t.Fatalf("protoc failed: %+v", err)
			}
			// Verify that the generated spec matches our expected version.
			err = exec.Command("diff", "openapi.yaml", path.Join(tt.path, "openapi.yaml")).Run()
			if err != nil {
				t.Fatalf("Diff failed: %+v", err)
			}
			// if the test succeeded, clean up
			os.Remove("openapi.yaml")
		})
	}
}

func TestOpenAPIJSONNaming(t *testing.T) {
	for _, tt := range openapiTests {
		t.Run(tt.name, func(t *testing.T) {
			// Run protoc and the protoc-gen-openapi plugin to generate an OpenAPI spec with JSON naming.
			err := exec.Command("protoc",
				"-I", "../../",
				"-I", "../../third_party",
				"-I", "examples",
				path.Join(tt.path, tt.protofile),
				"--openapi_out=version=1.2.3:.").Run()
			if err != nil {
				t.Fatalf("protoc failed: %+v", err)
			}
			// Verify that the generated spec matches our expected version.
			err = exec.Command("diff", "openapi.yaml", path.Join(tt.path, "openapi_json.yaml")).Run()
			if err != nil {
				t.Fatalf("Diff failed: %+v", err)
			}
			// if the test succeeded, clean up
			os.Remove("openapi.yaml")
		})
	}
}
