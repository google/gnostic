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
	"errors"
	"io"
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
	{name: "RPC types", path: "examples/tests/rpctypes/", protofile: "message.proto"},
	{name: "JSON options", path: "examples/tests/jsonoptions/", protofile: "message.proto"},
	{name: "Ignore services without annotations", path: "examples/tests/noannotations/", protofile: "message.proto"},
	{name: "Enum Options", path: "examples/tests/enumoptions/", protofile: "message.proto"},
	{name: "OpenAPIv3 Annotations", path: "examples/tests/openapiv3annotations/", protofile: "message.proto"},
	{name: "AllOf Wrap Message", path: "examples/tests/allofwrap/", protofile: "message.proto"},
	{name: "Additional Bindings", path: "examples/tests/additional_bindings/", protofile: "message.proto"},
}

// Set this to true to generate/overwrite the fixtures. Make sure you set it back
// to false before you commit it.
const GENERATE_FIXTURES = false

const TEMP_FILE = "openapi.yaml"

func CopyFixture(result, fixture string) error {
	in, err := os.Open(result)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(fixture)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func TestGenerateFixturesIsFalse(t *testing.T) {
	// This is here to ensure the PR builds fail if someone
	// accidentally commits GENERATE_FIXTURES = true
	if GENERATE_FIXTURES {
		t.Fatalf("GENERATE_FIXTURES is true")
	}
}

func TestOpenAPIProtobufNaming(t *testing.T) {
	for _, tt := range openapiTests {
		fixture := path.Join(tt.path, "openapi.yaml")
		if _, err := os.Stat(fixture); errors.Is(err, os.ErrNotExist) {
			if !GENERATE_FIXTURES {
				continue
			}
		}
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
			if GENERATE_FIXTURES {
				err := CopyFixture(TEMP_FILE, fixture)
				if err != nil {
					t.Fatalf("Can't generate fixture: %+v", err)
				}
			} else {
				// Verify that the generated spec matches our expected version.
				err = exec.Command("diff", TEMP_FILE, fixture).Run()
				if err != nil {
					t.Fatalf("Diff failed: %+v", err)
				}
			}
			// if the test succeeded, clean up
			os.Remove(TEMP_FILE)
		})
	}
}

func TestOpenAPIFQSchemaNaming(t *testing.T) {
	for _, tt := range openapiTests {
		fixture := path.Join(tt.path, "openapi_fq_schema_naming.yaml")
		if _, err := os.Stat(fixture); errors.Is(err, os.ErrNotExist) {
			if !GENERATE_FIXTURES {
				continue
			}
		}
		t.Run(tt.name, func(t *testing.T) {
			// Run protoc and the protoc-gen-openapi plugin to generate an OpenAPI spec.
			err := exec.Command("protoc",
				"-I", "../../",
				"-I", "../../third_party",
				"-I", "examples",
				path.Join(tt.path, tt.protofile),
				"--openapi_out=fq_schema_naming=1:.").Run()
			if err != nil {
				t.Fatalf("protoc failed: %+v", err)
			}
			if GENERATE_FIXTURES {
				err := CopyFixture(TEMP_FILE, fixture)
				if err != nil {
					t.Fatalf("Can't generate fixture: %+v", err)
				}
			} else {
				// Verify that the generated spec matches our expected version.
				err = exec.Command("diff", TEMP_FILE, fixture).Run()
				if err != nil {
					t.Fatalf("Diff failed: %+v", err)
				}
			}
			// if the test succeeded, clean up
			os.Remove(TEMP_FILE)
		})
	}
}

func TestOpenAPIJSONNaming(t *testing.T) {
	for _, tt := range openapiTests {
		fixture := path.Join(tt.path, "openapi_json.yaml")
		if _, err := os.Stat(fixture); errors.Is(err, os.ErrNotExist) {
			if !GENERATE_FIXTURES {
				continue
			}
		}
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
			if GENERATE_FIXTURES {
				err := CopyFixture(TEMP_FILE, fixture)
				if err != nil {
					t.Fatalf("Can't generate fixture: %+v", err)
				}
			} else {
				// Verify that the generated spec matches our expected version.
				err = exec.Command("diff", TEMP_FILE, fixture).Run()
				if err != nil {
					t.Fatalf("Diff failed: %+v", err)
				}
			}
			// if the test succeeded, clean up
			os.Remove(TEMP_FILE)
		})
	}
}

func TestOpenAPIStringEnums(t *testing.T) {
	for _, tt := range openapiTests {
		fixture := path.Join(tt.path, "openapi_string_enum.yaml")
		if _, err := os.Stat(fixture); errors.Is(err, os.ErrNotExist) {
			if !GENERATE_FIXTURES {
				continue
			}
		}
		t.Run(tt.name, func(t *testing.T) {
			// Run protoc and the protoc-gen-openapi plugin to generate an OpenAPI spec with string Enums.
			err := exec.Command("protoc",
				"-I", "../../",
				"-I", "../../third_party",
				"-I", "examples",
				path.Join(tt.path, tt.protofile),
				"--openapi_out=enum_type=string:.").Run()
			if err != nil {
				t.Fatalf("protoc failed: %+v", err)
			}
			if GENERATE_FIXTURES {
				err := CopyFixture(TEMP_FILE, fixture)
				if err != nil {
					t.Fatalf("Can't generate fixture: %+v", err)
				}
			} else {
				// Verify that the generated spec matches our expected version.
				err = exec.Command("diff", TEMP_FILE, fixture).Run()
				if err != nil {
					t.Fatalf("diff failed: %+v", err)
				}
			}
			// if the test succeeded, clean up
			os.Remove(TEMP_FILE)
		})
	}
}

func TestOpenAPIDefaultResponse(t *testing.T) {
	for _, tt := range openapiTests {
		fixture := path.Join(tt.path, "openapi_default_response.yaml")
		if _, err := os.Stat(fixture); errors.Is(err, os.ErrNotExist) {
			if !GENERATE_FIXTURES {
				continue
			}
		}
		t.Run(tt.name, func(t *testing.T) {
			// Run protoc and the protoc-gen-openapi plugin to generate an OpenAPI spec with string Enums.
			err := exec.Command("protoc",
				"-I", "../../",
				"-I", "../../third_party",
				"-I", "examples",
				path.Join(tt.path, tt.protofile),
				"--openapi_out=default_response=true:.").Run()
			if err != nil {
				t.Fatalf("protoc failed: %+v", err)
			}
			if GENERATE_FIXTURES {
				err := CopyFixture(TEMP_FILE, fixture)
				if err != nil {
					t.Fatalf("Can't generate fixture: %+v", err)
				}
			} else {
				// Verify that the generated spec matches our expected version.
				err = exec.Command("diff", TEMP_FILE, fixture).Run()
				if err != nil {
					t.Fatalf("diff failed: %+v", err)
				}
			}
			// if the test succeeded, clean up
			os.Remove(TEMP_FILE)
		})
	}
}
