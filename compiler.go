//go:generate ./COMPILE-PROTOS.sh

package main

import (
	"fmt"
        "jsonschema"
)


func main() {
   fmt.Printf("Hello\n")

   var s jsonschema.Schema
   fmt.Printf("%+v\n", s)
}
