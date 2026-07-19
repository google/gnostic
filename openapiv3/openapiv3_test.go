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

package openapi_v3

import (
	"os"
	"testing"
)

func TestParseDocument(t *testing.T) {
	filename := "../examples/v3.0/yaml/petstore.yaml"
	b, err := os.ReadFile(filename)
	if err != nil {
		t.Logf("unable to read file %s", filename)
		t.FailNow()
	}
	d, err := ParseDocument(b)
	if err != nil {
		t.Logf("%s", err.Error())
		t.FailNow()
	}
	title := "OpenAPI Petstore"
	if d.Info.Title != title {
		t.Errorf("unexpected value for Title: %s (expected %s)", d.Info.Title, title)
	}
}

func TestParseDocument_XquikOpenAPI31(t *testing.T) {
	filename := "../examples/v3.1/yaml/xquik.yaml"
	b, err := os.ReadFile(filename)
	if err != nil {
		t.Logf("unable to read file %s", filename)
		t.FailNow()
	}
	d, err := ParseDocument(b)
	if err != nil {
		t.Logf("%s", err.Error())
		t.FailNow()
	}
	if d.Info.Title != "Xquik API" {
		t.Errorf("unexpected value for Title: %s", d.Info.Title)
	}
	if !documentHasPath(d, "/api/v1/x/tweets/search") {
		t.Errorf("missing Xquik search path")
	}
}

func documentHasPath(document *Document, name string) bool {
	for _, path := range document.Paths.Path {
		if path.Name == name {
			return true
		}
	}
	return false
}

func TestParseDocument_Empty(t *testing.T) {
	for _, test := range []struct {
		name string
		data []byte
	}{
		{"nil", nil},
		{"zero_bytes", []byte{}},
		{"whitespace", []byte("   ")},
	} {
		t.Run(test.name, func(t *testing.T) {
			d, err := ParseDocument(test.data)

			t.Log(err)
			if err == nil {
				t.Error("expected error")
			} else if want, got := "document has no content", err.Error(); want != got {
				t.Errorf("unexpected error: %q (expected %q)", got, want)
			}
			if d != nil {
				t.Error("expected document to be nil")
			}
		})
	}
}
