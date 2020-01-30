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

package compiler

import (
	. "gopkg.in/check.v1"
	"io"
	"net/http"
	"testing"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) {
	TestingT(t)
}

var mockServer *http.Server

type ReaderTestingSuite struct{}

var _ = Suite(&ReaderTestingSuite{})

func (s *ReaderTestingSuite) SetUpSuite(c *C) {
	mockServer = &http.Server{Addr: "127.0.0.1:8080", Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		yamlBytes, err := ReadBytesForFile("testdata/petstore.yaml")
		c.Assert(err, IsNil)
		io.WriteString(w, string(yamlBytes))

	})}
	go func() {
		mockServer.ListenAndServe()
	}()
}

func (s *ReaderTestingSuite) TearDownSuite(c *C) {
	mockServer.Close()
}

func (s *ReaderTestingSuite) TestRemoveFromInfoCache(c *C) {
	fileName := "testdata/petstore.yaml"
	yamlBytes, err := ReadBytesForFile(fileName)
	c.Assert(err, IsNil)
	c.Assert(len(yamlBytes) > 0, Equals, true)
	petstore, err := ReadInfoFromBytes(fileName, yamlBytes)
	c.Assert(err, IsNil)
	c.Assert(petstore, NotNil)
	c.Assert(len(infoCache), Equals, 1)
	RemoveFromInfoCache(fileName)
	c.Assert(len(infoCache), Equals, 0)
}

func (s *ReaderTestingSuite) TestDisableInfoCache(c *C) {
	fileName := "testdata/petstore.yaml"
	yamlBytes, err := ReadBytesForFile(fileName)
	c.Assert(err, IsNil)
	c.Assert(len(yamlBytes) > 0, Equals, true)
	DisableInfoCache()
	petstore, err := ReadInfoFromBytes(fileName, yamlBytes)
	c.Assert(err, IsNil)
	c.Assert(petstore, NotNil)
	c.Assert(len(infoCache), Equals, 0)
}

func (s *ReaderTestingSuite) TestRemoveFromFileCache(c *C) {
	fileUrl := "http://127.0.0.1:8080/petstore"
	yamlBytes, err := FetchFile(fileUrl)
	c.Assert(err, IsNil)
	c.Assert(len(yamlBytes) > 0, Equals, true)
	c.Assert(len(fileCache), Equals, 1)
	RemoveFromFileCache(fileUrl)
	c.Assert(len(fileCache), Equals, 0)
}

func (s *ReaderTestingSuite) TestDisableFileCache(c *C) {
	DisableFileCache()
	fileUrl := "http://127.0.0.1:8080/petstore"
	yamlBytes, err := FetchFile(fileUrl)
	c.Assert(err, IsNil)
	c.Assert(len(yamlBytes) > 0, Equals, true)
	c.Assert(len(fileCache), Equals, 0)
}
