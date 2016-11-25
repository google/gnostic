package main

import (
	"os/exec"
	"testing"
)

func TestPetstoreJSON(t *testing.T) {
	var err error
	err = exec.Command("rm", "petstore.text").Run()
	if err != nil {
		panic(err)
	}
	err = exec.Command("./openapi-compiler", "-input", "examples/petstore.json", "-text").Run()
	if err != nil {
		panic(err)
	}
	err = exec.Command("diff", "petstore.text", "test/petstore.text").Run()
	if err != nil {
		panic(err)
	}
}

func TestPetstoreYAML(t *testing.T) {
	var err error
	err = exec.Command("rm", "petstore.text").Run()
	if err != nil {
		panic(err)
	}
	err = exec.Command("./openapi-compiler", "-input", "examples/petstore.yaml", "-text").Run()
	if err != nil {
		panic(err)
	}
	err = exec.Command("diff", "petstore.text", "test/petstore.text").Run()
	if err != nil {
		panic(err)
	}
}
