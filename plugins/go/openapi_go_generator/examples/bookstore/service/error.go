/*
 Copyright 2017 Google Inc. All Rights Reserved.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package main

import (
	"fmt"
	"net/http"
)

// Represents an error with an attached code. The zero value is valid, and represents an HTTP 200
// StatusOK with an empty error message.
type httperror struct {
	StatusCode int
	Err        error
}

func (e httperror) Error() string {
	if e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

func (e httperror) Code() int {
	if e.StatusCode == 0 {
		return http.StatusOK
	}
	return e.StatusCode
}

// Returns an Error value with the indicated code and formatted message. This is a convenience
// wrapper for &httperror{StatusCode: code, fmt.Errorf(format, a...)}.
func httpErrorf(code int, format string, a ...interface{}) error {
	return httperror{StatusCode: code, Err: fmt.Errorf(format, a...)}
}

// Returns the HTTP code for the supplied error, falling back to http.StatusInternalServerError if
// the error is not an httperror.
func getCode(err error) int {
	return getCodeWithDefault(err, http.StatusInternalServerError)
}

// Returns the HTTP code for the supplied error, with the given default code if the error is not an
// httperror.
func getCodeWithDefault(err error, code int) int {
	if err == nil {
		return http.StatusOK
	}
	if e, ok := err.(httperror); ok {
		return e.Code()
	}
	return code
}

// Complete the given HTTP request with the supplied error. If the error value is not
// httperror, http.StatusInternalServerError will be used.
func completeRequestWithError(w http.ResponseWriter, err error) {
	msg := err.Error()
	if msg != "" {
		msg = "Error: " + msg
	}
	http.Error(w, msg, getCode(err))
}
