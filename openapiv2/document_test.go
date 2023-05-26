// Copyright 2023 Google LLC. All Rights Reserved.
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

package openapi_v2

import "testing"

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
