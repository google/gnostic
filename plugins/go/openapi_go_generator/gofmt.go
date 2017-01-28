package main

import (
	"io/ioutil"
	"os/exec"
	"runtime"
)

func gofmt(inputBytes []byte) (outputBytes []byte, err error) {
	cmd := exec.Command(runtime.GOROOT() + "/bin/gofmt")
	input, _ := cmd.StdinPipe()
	output, _ := cmd.StdoutPipe()
	cmderr, _ := cmd.StderrPipe()
	err = cmd.Start()
	if err != nil {
		return
	}
	input.Write(inputBytes)
	input.Close()
	errors, err := ioutil.ReadAll(cmderr)
	if len(errors) > 0 {
		return inputBytes, nil
	} else {
		outputBytes, err = ioutil.ReadAll(output)
	}
	return
}
