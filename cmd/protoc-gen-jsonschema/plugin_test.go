// Copyright 2021 Google LLC. All Rights Reserved.
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
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/flowstack/go-jsonschema"
)

var (
	testSchemasPath = "testschemas"
)

var jsonschemaTests = []struct {
	name      string
	path      string
	pkg       string
	protofile string
}{
	{name: "Google Library example", path: "examples/google/example/library/v1/", pkg: "google.example.library.v1", protofile: "library.proto"},
	{name: "Map fields", path: "examples/tests/mapfields/", pkg: "tests.mapfields.message.v1", protofile: "message.proto"},
	{name: "JSON options", path: "examples/tests/jsonoptions/", pkg: "", protofile: "message.proto"},
	{name: "Embedded messages", path: "examples/tests/embedded/", pkg: "", protofile: "message.proto"},
	{name: "Protobuf types", path: "examples/tests/protobuftypes/", pkg: "", protofile: "message.proto"},
	{name: "Enum Options", path: "examples/tests/enumoptions/", pkg: "", protofile: "message.proto"},
}

func TestJSONSchemaProtobufNaming(t *testing.T) {
	for _, tt := range jsonschemaTests {
		schemasPath := path.Join(tt.path, "schemas_proto")
		if _, err := os.Stat(schemasPath); errors.Is(err, os.ErrNotExist) {
			continue
		}
		t.Run(tt.name, func(t *testing.T) {
			os.RemoveAll(testSchemasPath)
			os.MkdirAll(testSchemasPath, 0777)
			// Run protoc and the protoc-gen-jsonschema plugin to generate JSON Schema(s) with proto naming.
			err := exec.Command("protoc",
				"-I", "../../",
				"-I", "../../third_party",
				"-I", "examples",
				path.Join(tt.path, tt.protofile),
				"--jsonschema_opt=baseurl=http://example.com/schemas",
				"--jsonschema_opt=version=http://json-schema.org/draft-07/schema#",
				"--jsonschema_out=naming=proto,version=1.2.3:"+testSchemasPath).Run()
			if err != nil {
				t.Fatalf("protoc failed: %+v", err)
			}

			// Verify that the generated spec matches our expected version.
			err = exec.Command("diff", testSchemasPath, schemasPath).Run()
			if err != nil {
				t.Fatalf("Diff failed: %+v", err)
			}

			// if the test succeeded, clean up
			os.RemoveAll(testSchemasPath)
		})
	}
}

func TestJSONSchemaJSONNaming(t *testing.T) {
	for _, tt := range jsonschemaTests {
		schemasPath := path.Join(tt.path, "schemas_json")
		if _, err := os.Stat(schemasPath); errors.Is(err, os.ErrNotExist) {
			continue
		}
		t.Run(tt.name, func(t *testing.T) {
			os.RemoveAll(testSchemasPath)
			os.MkdirAll(testSchemasPath, 0777)
			// Run protoc and the protoc-gen-jsonschema plugin to generate JSON Schema(s) with JSON naming.
			err := exec.Command("protoc",
				"-I", "../../",
				"-I", "../../third_party",
				"-I", "examples",
				path.Join(tt.path, tt.protofile),
				"--jsonschema_opt=baseurl=http://example.com/schemas",
				"--jsonschema_out="+testSchemasPath).Run()
			if err != nil {
				t.Fatalf("protoc failed: %+v", err)
			}

			// Verify that the generated spec matches our expected version.
			err = exec.Command("diff", testSchemasPath, schemasPath).Run()
			if err != nil {
				t.Fatalf("Diff failed: %+v", err)
			}

			// if the test succeeded, clean up
			os.RemoveAll(testSchemasPath)
		})
	}
}

// Meta... Test the tests
func TestJSONSchemaJSONNamingSchemas(t *testing.T) {
	for _, tt := range jsonschemaTests {
		schemasPath := path.Join(tt.path, "schemas_json")
		if _, err := os.Stat(schemasPath); errors.Is(err, os.ErrNotExist) {
			continue
		}
		t.Run(tt.name, func(t *testing.T) {
			schemaFiles, err := os.ReadDir(schemasPath)
			if err != nil {
				t.Fatal(err)
			}

			for _, schemaFile := range schemaFiles {
				// Validate the schema itself against the JSON draft defined
				schema, err := os.ReadFile(path.Join(schemasPath, schemaFile.Name()))
				if err != nil {
					t.Fatalf("%s: %s", schemaFile.Name(), err.Error())
				}

				_, err = jsonschema.Validate(schema)
				if err != nil {
					t.Fatalf("%s: %s", schemaFile.Name(), err.Error())
				}

				// Verify that the validator works with the test json document
				validator, err := jsonschema.New(schema)
				if err != nil {
					t.Fatalf("%s: %s", schemaFile.Name(), err.Error())
				}

				// Add all the other schemas as refs, to be sure that any needed refs are available
				for _, refFile := range schemaFiles {
					if schemaFile.Name() != refFile.Name() {
						ref, err := os.ReadFile(path.Join(schemasPath, refFile.Name()))
						if err != nil {
							t.Fatalf("%s: %s", refFile.Name(), err.Error())
						}
						err = validator.AddSchemaString(string(ref))
						if err != nil {
							t.Fatal(err)
						}
					}
				}

				dataPath := path.Join(tt.path, "testdata_json")
				doc, err := os.ReadFile(path.Join(dataPath, schemaFile.Name()))
				if err != nil {
					t.Fatalf("%s: %s", schemaFile.Name(), err.Error())
				}

				_, err = validator.Validate(doc)
				if err != nil {
					t.Fatalf("%s: %s", schemaFile.Name(), err.Error())
				}
			}
		})
	}
}

func TestJSONSchemaStringEnums(t *testing.T) {
	for _, tt := range jsonschemaTests {
		schemasPath := path.Join(tt.path, "schemas_string_enum")
		if _, err := os.Stat(schemasPath); errors.Is(err, os.ErrNotExist) {
			continue
		}
		t.Run(tt.name, func(t *testing.T) {
			os.RemoveAll(testSchemasPath)
			os.MkdirAll(testSchemasPath, 0777)
			// Run protoc and the protoc-gen-jsonschema plugin to generate JSON Schema(s) with JSON naming.
			err := exec.Command("protoc",
				"-I", "../../",
				"-I", "../../third_party",
				"-I", "examples",
				path.Join(tt.path, tt.protofile),
				"--jsonschema_opt=baseurl=http://example.com/schemas",
				"--jsonschema_opt=enum_type=string",
				"--jsonschema_out="+testSchemasPath).Run()
			if err != nil {
				t.Fatalf("protoc failed: %+v", err)
			}

			// Verify that the generated spec matches our expected version.
			err = exec.Command("diff", testSchemasPath, schemasPath).Run()
			if err != nil {
				t.Fatalf("Diff failed: %+v", err)
			}

			// if the test succeeded, clean up
			os.RemoveAll(testSchemasPath)
		})
	}
}
