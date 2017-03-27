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

// summarize is a tool for summarizing the results of gnostic_analyze runs.
package main

import (
	"encoding/json"
	"fmt"
	"github.com/googleapis/gnostic/plugins/go/gnostic_analyze/statistics"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

// Results are collected in this global slice.
var stats []statistics.DocumentStatistics

// walker is called for each summary file found.
func walker(p string, info os.FileInfo, err error) error {
	basename := path.Base(p)
	if basename != "summary.json" {
		return nil
	}
	data, err := ioutil.ReadFile(p)
	if err != nil {
		return err
	}
	var s statistics.DocumentStatistics
	err = json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	stats = append(stats, s)
	return nil
}

func main() {
	// Collect all statistics in the current directory and its subdirectories.
	stats = make([]statistics.DocumentStatistics, 0)
	filepath.Walk(".", walker)

	// Compute some interesting properties.
	apis_with_anonymous_ops := 0
	op_frequencies := make(map[string]int, 0)
	parameter_type_frequencies := make(map[string]int, 0)
	result_type_frequencies := make(map[string]int, 0)
	definition_field_type_frequencies := make(map[string]int, 0)
	definition_array_type_frequencies := make(map[string]int, 0)

	for _, api := range stats {
		if api.Operations["anonymous"] != 0 {
			apis_with_anonymous_ops += 1
		}
		for k, v := range api.Operations {
			op_frequencies[k] += v
		}
		for k, v := range api.ParameterTypes {
			parameter_type_frequencies[k] += v
		}
		for k, v := range api.ResultTypes {
			result_type_frequencies[k] += v
		}
		for k, v := range api.DefinitionFieldTypes {
			definition_field_type_frequencies[k] += v
		}
		for k, v := range api.DefinitionArrayTypes {
			definition_array_type_frequencies[k] += v
		}
	}

	// Report the results.
	fmt.Printf("Collected information on %d APIs.\n\n",
		len(stats))
	fmt.Printf("apis with anonymous ops: %d\n\n",
		apis_with_anonymous_ops)
	fmt.Printf("op frequencies: %+v\n\n",
		op_frequencies)
	fmt.Printf("parameter type frequencies: %+v\n\n",
		parameter_type_frequencies)
	fmt.Printf("result type frequencies: %+v\n\n",
		result_type_frequencies)
	fmt.Printf("definition field type frequencies: %+v\n\n",
		definition_field_type_frequencies)
	fmt.Printf("definition array type frequencies: %+v\n\n",
		definition_array_type_frequencies)
}
