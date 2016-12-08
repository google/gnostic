package main

import (
	"log"
	"os"
	"os/exec"
	"testing"
)

func TestPetstoreJSON(t *testing.T) {
	var err error
	err = exec.Command("rm", "-f", "petstore.text").Run()
	if err != nil {
		panic(err)
	}
	err = exec.Command("./openapi-compiler", "-input", "examples/petstore.json", "-text").Run()
	if err != nil {
		log.Printf("JSON compile failed: %+v", err)
		os.Exit(-1)
	}
	err = exec.Command("diff", "petstore.text", "test/petstore.text").Run()
	if err != nil {
		log.Printf("JSON diff failed: %+v", err)
		os.Exit(-1)
	}
}

func TestPetstoreYAML(t *testing.T) {
	var err error
	err = exec.Command("rm", "-f", "petstore.text").Run()
	if err != nil {
		panic(err)
	}
	err = exec.Command("./openapi-compiler", "-input", "examples/petstore.yaml", "-text").Run()
	if err != nil {
		log.Printf("YAML compile failed: %+v", err)
		os.Exit(-1)
	}
	err = exec.Command("diff", "petstore.text", "test/petstore.text").Run()
	if err != nil {
		log.Printf("YAML diff failed: %+v", err)
		os.Exit(-1)
	}
}

func TestSeparateYAML(t *testing.T) {
	var err error
	err = exec.Command("rm", "-f", "swagger.text").Run()
	if err != nil {
		panic(err)
	}
	err = exec.Command("./openapi-compiler", "-input", "examples/v2.0/yaml/petstore-separate/spec/swagger.yaml", "-text").Run()
	if err != nil {
		log.Printf("YAML compile failed: %+v", err)
		os.Exit(-1)
	}
	err = exec.Command("diff", "swagger.text", "test/v2.0/yaml/petstore-separate/spec/swagger.text").Run()
	if err != nil {
		log.Printf("YAML diff failed: %+v", err)
		os.Exit(-1)
	}
}
