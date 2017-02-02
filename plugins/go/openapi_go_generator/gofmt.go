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

package main

import (
	"io/ioutil"
	"log"
	"os/exec"
	"runtime"
	"strings"
)

// Remove lines containing only "//-" after templates have been expanded.
// Code templates use "//-" prefixes to mark template operators
// that otherwise would add unnecessary blank lines.
func stripMarkers(inputBytes []byte) (outputBytes []byte) {
	inputString := string(inputBytes)
	inputLines := strings.Split(inputString, "\n")
	outputLines := make([]string, 0)
	for _, line := range inputLines {
		if strings.Contains(line, "//-") {
			removed := strings.TrimSpace(strings.Replace(line, "//-", "", 1))
			if removed != "" {
				outputLines = append(outputLines, removed)
			}
		} else {
			outputLines = append(outputLines, line)
		}
	}
	outputString := strings.Join(outputLines, "\n")
	return []byte(outputString)
}

// Run the gofmt tool to format generated code.
func gofmt(inputBytes []byte) (outputBytes []byte, err error) {
	cmd := exec.Command(runtime.GOROOT() + "/bin/gofmt")
	input, _ := cmd.StdinPipe()
	output, _ := cmd.StdoutPipe()
	cmderr, _ := cmd.StderrPipe()
	err = cmd.Start()
	if err != nil {
		return
	}
	strippedBytes := stripMarkers(inputBytes)
	input.Write(strippedBytes)
	input.Close()
	errors, err := ioutil.ReadAll(cmderr)
	if len(errors) > 0 {
		log.Printf("gofmt errors: %s", errors)
		return strippedBytes, nil
	} else {
		outputBytes, err = ioutil.ReadAll(output)
	}
	return
}
