// Copyright 2017 Google Inc. All Rights Reserved.
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
package discovery

import (
	"encoding/json"
)

const APIsListServiceURL = "https://www.googleapis.com/discovery/v1/apis"

// A List represents the results of a call to the apis/list API.
// https://developers.google.com/discovery/v1/reference/apis/list
type List struct {
	Kind             string `json:"kind"`
	DiscoveryVersion string `json:"discoveryVersion"`
	APIs             []*API `json:"items"`
}

// NewList unmarshals the bytes into a Document.
func NewList(bytes []byte) (*List, error) {
	var listResponse List
	err := json.Unmarshal(bytes, &listResponse)
	return &listResponse, err
}

// An API represents the an API description returned by the apis/list API.
type API struct {
	Kind              string            `json:"kind"`
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	Version           string            `json:"version"`
	Title             string            `json:"title"`
	Description       string            `json:"description"`
	DiscoveryRestURL  string            `json:"discoveryRestUrl"`
	DiscoveryLink     string            `json:"discoveryLink"`
	Icons             map[string]string `json:"icons"`
	DocumentationLink string            `json:"documentationLink"`
	Labels            []string          `json:"labels"`
	Preferred         bool              `json:"preferred"`
}

// APIWithName returns the first API with a specified name or nil if none exists.
func (a *List) APIWithName(name string) *API {
	for _, item := range a.APIs {
		if item.Name == name {
			return item
		}
	}
	return nil
}

// APIWithID returns the first API with a specified ID or nil if none exists.
func (a *List) APIWithID(id string) *API {
	for _, item := range a.APIs {
		if item.ID == id {
			return item
		}
	}
	return nil
}
