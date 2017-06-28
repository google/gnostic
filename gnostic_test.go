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
	var cmd = exec.Command(
		"gnostic",
		input_file,
		"--text-out=.",
		"--errors-out=.",
		"--resolve-refs")
	//t.Log(cmd.Args)
	err = cmd.Run()
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
		"examples/v2.0/json/petstore.json",
		"test/v2.0/petstore.text")
}

func TestPetstoreYAML(t *testing.T) {
	test_normal(t,
		"examples/v2.0/yaml/petstore.yaml",
		"test/v2.0/petstore.text")
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
		"https://raw.githubusercontent.com/googleapis/openapi-compiler/master/examples/v2.0/json/petstore.json",
		"test/v2.0/petstore.text")
}

func TestRemotePetstoreYAML(t *testing.T) {
	test_normal(t,
		"https://raw.githubusercontent.com/googleapis/openapi-compiler/master/examples/v2.0/yaml/petstore.yaml",
		"test/v2.0/petstore.text")
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

func TestErrorMissingVersion(t *testing.T) {
	test_errors(t,
		"examples/errors/petstore-missingversion.yaml",
		"test/errors/petstore-missingversion.errors")
}

func test_plugin(t *testing.T, plugin string, input_file string, output_file string, reference_file string) {
	// remove any preexisting output files
	os.Remove(output_file)
	// run the compiler
	var err error
	output, err := exec.Command(
		"gnostic",
		"--"+plugin+"-out=-",
		input_file).Output()
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
		"go-sample",
		"examples/v2.0/yaml/petstore.yaml",
		"sample-petstore.out",
		"test/v2.0/yaml/sample-petstore.out")
}

func TestErrorInvalidPluginInvocations(t *testing.T) {
	var err error
	output, err := exec.Command(
		"gnostic",
		"examples/v2.0/yaml/petstore.yaml",
		"--errors-out=-",
		"--plugin-out=foo=bar,:abc",
		"--plugin-out=,foo=bar:abc",
		"--plugin-out=foo=:abc",
		"--plugin-out==bar:abc",
		"--plugin-out=,,:abc",
		"--plugin-out=foo=bar=baz:abc",
	).Output()
	if err == nil {
		t.Logf("Invalid invocations were accepted")
		t.FailNow()
	}
	output_file := "invalid-plugin-invocation.errors"
	_ = ioutil.WriteFile(output_file, output, 0644)
	err = exec.Command("diff", output_file, "test/errors/invalid-plugin-invocation.errors").Run()
	if err != nil {
		t.Logf("Diff failed: %+v", err)
		t.FailNow()
	} else {
		// if the test succeeded, clean up
		os.Remove(output_file)
	}
}

func TestValidPluginInvocations(t *testing.T) {
	var err error
	output, err := exec.Command(
		"gnostic",
		"examples/v2.0/yaml/petstore.yaml",
		"--errors-out=-",
		// verify an invocation with no parameters
		"--go-sample-out=!", // "!" indicates that no output should be generated
		// verify single pair of parameters
		"--go-sample-out=a=b:!",
		// verify multiple parameters
		"--go-sample-out=a=b,c=123,xyz=alphabetagammadelta:!",
		// verify that special characters / . - _ can be included in parameter keys and values
		"--go-sample-out=a/b/c=x/y/z:!",
		"--go-sample-out=a.b.c=x.y.z:!",
		"--go-sample-out=a-b-c=x-y-z:!",
		"--go-sample-out=a_b_c=x_y_z:!",
	).Output()
	if len(output) != 0 {
		t.Logf("Valid invocations generated invalid errors\n%s", string(output))
		t.FailNow()
	}
	if err != nil {
		t.Logf("Valid invocations were not accepted")
		t.FailNow()
	}
}

func TestExtensionHandlerWithLibraryExample(t *testing.T) {
	output_file := "library-example-with-ext.text.out"
	input_file := "test/library-example-with-ext.json"
	reference_file := "test/library-example-with-ext.text.out"

	os.Remove(output_file)
	// run the compiler
	var err error

	command := exec.Command(
		"gnostic",
		"--x-sampleone",
		"--x-sampletwo",
		"--text-out="+output_file,
		"--resolve-refs",
		input_file)

	_, err = command.Output()
	if err != nil {
		t.Logf("Compile failed for command %v: %+v", command, err)
		t.FailNow()
	}
	//_ = ioutil.WriteFile(output_file, output, 0644)
	err = exec.Command("diff", output_file, reference_file).Run()
	if err != nil {
		t.Logf("Diff failed: %+v", err)
		t.FailNow()
	} else {
		// if the test succeeded, clean up
		os.Remove(output_file)
	}
}

func TestJSONOutput(t *testing.T) {
	input_file := "test/library-example-with-ext.json"

	text_file := "sample.text"
	json_file := "sample.json"
	text_file2 := "sample2.text"
	json_file2 := "sample2.json"

	os.Remove(text_file)
	os.Remove(json_file)
	os.Remove(text_file2)
	os.Remove(json_file2)

	var err error

	// Run the compiler once.
	command := exec.Command(
		"gnostic",
		"--text-out="+text_file,
		"--json-out="+json_file,
		input_file)
	_, err = command.Output()
	if err != nil {
		t.Logf("Compile failed for command %v: %+v", command, err)
		t.FailNow()
	}

	// Run the compiler again, this time on the generated output.
	command = exec.Command(
		"gnostic",
		"--text-out="+text_file2,
		"--json-out="+json_file2,
		json_file)
	_, err = command.Output()
	if err != nil {
		t.Logf("Compile failed for command %v: %+v", command, err)
		t.FailNow()
	}

	// Verify that both models have the same internal representation.
	err = exec.Command("diff", text_file, text_file2).Run()
	if err != nil {
		t.Logf("Diff failed: %+v", err)
		t.FailNow()
	} else {
		// if the test succeeded, clean up
		os.Remove(text_file)
		os.Remove(json_file)
		os.Remove(text_file2)
		os.Remove(json_file2)
	}
}

func TestYAMLOutput(t *testing.T) {
	input_file := "test/library-example-with-ext.json"

	text_file := "sample.text"
	yaml_file := "sample.yaml"
	text_file2 := "sample2.text"
	yaml_file2 := "sample2.yaml"

	os.Remove(text_file)
	os.Remove(yaml_file)
	os.Remove(text_file2)
	os.Remove(yaml_file2)

	var err error

	// Run the compiler once.
	command := exec.Command(
		"gnostic",
		"--text-out="+text_file,
		"--yaml-out="+yaml_file,
		input_file)
	_, err = command.Output()
	if err != nil {
		t.Logf("Compile failed for command %v: %+v", command, err)
		t.FailNow()
	}

	// Run the compiler again, this time on the generated output.
	command = exec.Command(
		"gnostic",
		"--text-out="+text_file2,
		"--yaml-out="+yaml_file2,
		yaml_file)
	_, err = command.Output()
	if err != nil {
		t.Logf("Compile failed for command %v: %+v", command, err)
		t.FailNow()
	}

	// Verify that both models have the same internal representation.
	err = exec.Command("diff", text_file, text_file2).Run()
	if err != nil {
		t.Logf("Diff failed: %+v", err)
		t.FailNow()
	} else {
		// if the test succeeded, clean up
		os.Remove(text_file)
		os.Remove(yaml_file)
		os.Remove(text_file2)
		os.Remove(yaml_file2)
	}
}

func test_builder(version string, t *testing.T) {
	var err error

	pb_file := "petstore-" + version + ".pb"
	yaml_file := "petstore.yaml"
	json_file := "petstore.json"
	text_file := "petstore.text"
	text_reference := "test/" + version + ".0/petstore.text"

	os.Remove(pb_file)
	os.Remove(text_file)
	os.Remove(yaml_file)
	os.Remove(json_file)

	// Generate petstore.pb.
	command := exec.Command(
		"petstore-builder",
		"--"+version)
	_, err = command.Output()
	if err != nil {
		t.Logf("Command %v failed: %+v", command, err)
		t.FailNow()
	}

	// Convert petstore.pb to yaml and json.
	command = exec.Command(
		"gnostic",
		pb_file,
		"--json-out="+json_file,
		"--yaml-out="+yaml_file)
	_, err = command.Output()
	if err != nil {
		t.Logf("Command %v failed: %+v", command, err)
		t.FailNow()
	}

	// Read petstore.yaml, resolve references, and export text.
	command = exec.Command(
		"gnostic",
		yaml_file,
		"--resolve-refs",
		"--text-out="+text_file)
	_, err = command.Output()
	if err != nil {
		t.Logf("Command %v failed: %+v", command, err)
		t.FailNow()
	}

	// Verify that the generated text matches our reference.
	err = exec.Command("diff", text_file, text_reference).Run()
	if err != nil {
		t.Logf("Diff failed: %+v", err)
		t.FailNow()
	}

	// Read petstore.json, resolve references, and export text.
	command = exec.Command(
		"gnostic",
		json_file,
		"--resolve-refs",
		"--text-out="+text_file)
	_, err = command.Output()
	if err != nil {
		t.Logf("Command %v failed: %+v", command, err)
		t.FailNow()
	}

	// Verify that the generated text matches our reference.
	err = exec.Command("diff", text_file, text_reference).Run()
	if err != nil {
		t.Logf("Diff failed: %+v", err)
		t.FailNow()
	}

	// if the test succeeded, clean up
	os.Remove(pb_file)
	os.Remove(text_file)
	os.Remove(yaml_file)
	os.Remove(json_file)
}

func TestBuilderV2(t *testing.T) {
	test_builder("v2", t)
}

func TestBuilderV3(t *testing.T) {
	test_builder("v3", t)
}

// OpenAPI 3.0 tests

func TestPetstoreYAML_30(t *testing.T) {
	test_normal(t,
		"examples/v3.0/yaml/petstore.yaml",
		"test/v3.0/petstore.text")
}

func TestPetstoreJSON_30(t *testing.T) {
	test_normal(t,
		"examples/v3.0/json/petstore.json",
		"test/v3.0/petstore.text")
}
