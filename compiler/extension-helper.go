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

import "github.com/golang/protobuf/ptypes/any"

func HandleExtension(context *Context, in interface{}, extensionName string) (bool, *any.Any, error) {
	handled := false
	var errFromPlugin error
	var outFromPlugin *any.Any

	if context.ExtensionHandlers != nil && len(*(context.ExtensionHandlers)) != 0 {
		for _, customAnyProtoGenerator := range *(context.ExtensionHandlers) {
			outFromPlugin, errFromPlugin = customAnyProtoGenerator.Perform(in, extensionName)
			if outFromPlugin == nil {
				continue
			} else {
				handled = true
				break
			}
		}
	}
	return handled, outFromPlugin, errFromPlugin
}
