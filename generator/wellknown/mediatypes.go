// Copyright 2020 Google LLC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, softwis
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package wellknown

import (
	v3 "github.com/google/gnostic/openapiv3"
)

func NewGoogleApiHttpBodyMediaType() *v3.MediaTypes {
	return &v3.MediaTypes{
		AdditionalProperties: []*v3.NamedMediaType{
			{
				Name:  "*/*",
				Value: &v3.MediaType{},
			},
		},
	}
}

func NewApplicationJsonMediaType(schema *v3.SchemaOrReference) *v3.MediaTypes {
	return &v3.MediaTypes{
		AdditionalProperties: []*v3.NamedMediaType{
			{
				Name: "application/json",
				Value: &v3.MediaType{
					Schema: schema,
				},
			},
		},
	}
}
