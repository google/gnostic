// Copyright 2017 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package disco represents Google API discovery documents.
package disco

import (
	"encoding/json"
)

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
