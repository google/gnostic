package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
)

func TestErrorExtensionGeneratorUnsupportedPrimitive(t *testing.T) {
	var err error

	output, err := exec.Command(
		"generator",
		"--extension",
		"test/x-unsupportedprimitives.json",
		"--out_dir=/tmp",
	).Output()

	output_file := "x-unsupportedprimitives.errors"
	_ = ioutil.WriteFile(output_file, output, 0644)
	err = exec.Command("diff", output_file, "test/errors/x-unsupportedprimitives.errors").Run()
	if err != nil {
		t.Logf("Diff failed: %+v", err)
		t.FailNow()
	} else {
		// if the test succeeded, clean up
		os.Remove(output_file)
	}
}

func TestErrorExtensionGeneratorNameCollision(t *testing.T) {
	var err error

	output, err := exec.Command(
		"generator",
		"--extension",
		"test/x-extension-name-collision.json",
		"--out_dir=/tmp",
	).Output()

	output_file := "x-extension-name-collision.errors"
	_ = ioutil.WriteFile(output_file, output, 0644)
	err = exec.Command("diff", output_file, "test/errors/x-extension-name-collision.errors").Run()
	if err != nil {
		t.Logf("Diff failed: %+v", err)
		t.FailNow()
	} else {
		// if the test succeeded, clean up
		os.Remove(output_file)
	}
}
