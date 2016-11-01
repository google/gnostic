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
