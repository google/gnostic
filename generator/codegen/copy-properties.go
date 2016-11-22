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

// this generates the contents of the copyProperties function

package main

import (
	"fmt"
	//"sort"
)

func main() {

	properties := []string{"Schema",
		"Id",
		"MultipleOf",
		"Maximum",
		"ExclusiveMaximum",
		"Minimum",
		"ExclusiveMinimum",
		"MaxLength",
		"MinLength",
		"Pattern",
		"AdditionalItems",
		"Items",
		"MaxItems",
		"MinItems",
		"UniqueItems",
		"MaxProperties",
		"MinProperties",
		"Required",
		"AdditionalProperties",
		"Properties",
		"PatternProperties",
		"Dependencies",
		"Enumeration",
		"Type",
		"AllOf",
		"AnyOf",
		"OneOf",
		"Not",
		"Definitions",
		"Title",
		"Description",
		"Default",
		"Format",
		"Ref"}
	//sort.Strings(properties)

	for _, property := range properties {
		fmt.Printf("if source.%s != nil {", property)
		fmt.Printf("destination.%s = source.%s", property, property)
		fmt.Printf("}\n")
	}
}
