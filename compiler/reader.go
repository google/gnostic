// Copyright 2016 Google Inc. All Rights Reserved.
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
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// read a file and unmarshal it as a yaml.MapSlice
func ReadFile(filename string) interface{} {
	// is the filename a url?
	fileurl, _ := url.Parse(filename)
	if fileurl.Scheme != "" {
		// yes it is, so fetch it
		log.Printf("fetching %s", filename)
		response, err := http.Get(filename)
		if err != nil {
			log.Fatal(err)
		} else {
			defer response.Body.Close()
			bytes, err := ioutil.ReadAll(response.Body)
			if err == nil {
				var info yaml.MapSlice
				yaml.Unmarshal(bytes, &info)
				return info
			}
		}
	} else {
		// no, it's a local filename
		file, e := ioutil.ReadFile(filename)
		if e != nil {
			fmt.Printf("File error: %v\n", e)
			os.Exit(1)
		}
		var info yaml.MapSlice
		yaml.Unmarshal(file, &info)
		return info
	}
	return nil
}

var info_cache map[string]interface{}
var count int64

// read a file and return the fragment needed to resolve a $ref
func ReadInfoForRef(basefile string, ref string) interface{} {
	if info_cache == nil {
		log.Printf("making cache")
		info_cache = make(map[string]interface{}, 0)
	}
	{
		info, ok := info_cache[ref]
		if ok {
			return info
		}
	}

	log.Printf("%d Resolving %s", count, ref)
	count = count + 1
	basedir, _ := filepath.Split(basefile)
	parts := strings.Split(ref, "#")
	var filename string
	if parts[0] != "" {
		filename = basedir + parts[0]
	} else {
		filename = basefile
	}
	info := ReadFile(filename)
	if len(parts) > 1 {
		path := strings.Split(parts[1], "/")
		for i, key := range path {
			if i > 0 {
				m, ok := info.(yaml.MapSlice)
				if ok {
					for _, section := range m {
						if section.Key == key {
							info = section.Value
						}
					}
				}
			}
		}
	}

	info_cache[ref] = info
	return info
}
