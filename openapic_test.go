package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func test_compiler(t *testing.T, input_file string, reference_file string, expect_errors bool) {
	text_file := strings.Replace(filepath.Base(input_file), filepath.Ext(input_file), ".text", 1)
	errors_file := strings.Replace(filepath.Base(input_file), filepath.Ext(input_file), ".errors", 1)
	// remove any preexisting output files
	os.Remove(text_file)
	os.Remove(errors_file)
	// run the compiler
	var err error
	err = exec.Command(
		"openapic",
		input_file,
		"--text_out="+text_file,
		"--errors_out="+errors_file,
		"--resolve_refs").Run()
	if err != nil && !expect_errors {
		t.Logf("Compile failed: %+v", err)
		t.FailNow()
	}
	// verify the output against a reference
	var output_file string
	if expect_errors {
		output_file = errors_file
	} else {
		output_file = text_file
	}
	err = exec.Command("diff", output_file, reference_file).Run()
	if err != nil {
		t.Logf("Diff failed: %+v", err)
		t.FailNow()
	} else {
		// if the test succeeded, clean up
		os.Remove(text_file)
		os.Remove(errors_file)
	}
}

func test_normal(t *testing.T, input_file string, reference_file string) {
	test_compiler(t, input_file, reference_file, false)
}

func test_errors(t *testing.T, input_file string, reference_file string) {
	test_compiler(t, input_file, reference_file, true)
}

func TestPetstoreJSON(t *testing.T) {
	test_normal(t,
		"examples/petstore.json",
		"test/petstore.text")
}

func TestPetstoreYAML(t *testing.T) {
	test_normal(t,
		"examples/petstore.yaml",
		"test/petstore.text")
}

func TestSeparateYAML(t *testing.T) {
	test_normal(t,
		"examples/v2.0/yaml/petstore-separate/spec/swagger.yaml",
		"test/v2.0/yaml/petstore-separate/spec/swagger.text")
}

func TestSeparateJSON(t *testing.T) {
	test_normal(t,
		"examples/v2.0/json/petstore-separate/spec/swagger.json",
		"test/v2.0/yaml/petstore-separate/spec/swagger.text") // yaml and json results should be identical
}

func TestRemotePetstoreJSON(t *testing.T) {
	test_normal(t,
		"https://raw.githubusercontent.com/googleapis/openapi-compiler/master/examples/petstore.json",
		"test/petstore.text")
}

func TestRemotePetstoreYAML(t *testing.T) {
	test_normal(t,
		"https://raw.githubusercontent.com/googleapis/openapi-compiler/master/examples/petstore.yaml",
		"test/petstore.text")
}

func TestRemoteSeparateYAML(t *testing.T) {
	test_normal(t,
		"https://raw.githubusercontent.com/googleapis/openapi-compiler/master/examples/v2.0/yaml/petstore-separate/spec/swagger.yaml",
		"test/v2.0/yaml/petstore-separate/spec/swagger.text")
}

func TestRemoteSeparateJSON(t *testing.T) {
	test_normal(t,
		"https://raw.githubusercontent.com/googleapis/openapi-compiler/master/examples/v2.0/json/petstore-separate/spec/swagger.json",
		"test/v2.0/yaml/petstore-separate/spec/swagger.text")
}

func TestErrorBadProperties(t *testing.T) {
	test_errors(t,
		"examples/errors/petstore-badproperties.yaml",
		"test/errors/petstore-badproperties.errors")
}

func TestErrorUnresolvedRefs(t *testing.T) {
	test_errors(t,
		"examples/errors/petstore-unresolvedrefs.yaml",
		"test/errors/petstore-unresolvedrefs.errors")
}

func test_plugin(t *testing.T, plugin string, input_file string, output_file string, reference_file string) {
	// remove any preexisting output files
	os.Remove(output_file)
	// run the compiler
	var err error
	output, err := exec.Command("openapic", "--"+plugin+"_out=-", input_file).Output()
	if err != nil {
		t.Logf("Compile failed: %+v", err)
		t.FailNow()
	}
	_ = ioutil.WriteFile(output_file, output, 0644)
	err = exec.Command("diff", output_file, reference_file).Run()
	if err != nil {
		t.Logf("Diff failed: %+v", err)
		t.FailNow()
	} else {
		// if the test succeeded, clean up
		os.Remove(output_file)
	}
}

func TestSamplePluginWithPetstore(t *testing.T) {
	test_plugin(t,
		"go_sample",
		"examples/petstore.yaml",
		"sample-petstore.out",
		"test/sample-petstore.out")
}
