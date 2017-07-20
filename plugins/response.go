package openapi_plugin_v1

import (
	"fmt"
	"io"
	"os"
	"path"
)

func HandleResponse(response *Response, outputLocation string) error {
	if response.Errors != nil {
		return fmt.Errorf("Plugin error: %+v", response.Errors)
	}

	// Write files to the specified directory.
	var writer io.Writer
	switch {
	case outputLocation == "!":
		// Write nothing.
	case outputLocation == "-":
		writer = os.Stdout
		for _, file := range response.Files {
			writer.Write([]byte("\n\n" + file.Name + " -------------------- \n"))
			writer.Write(file.Data)
		}
	case isFile(outputLocation):
		return fmt.Errorf("unable to overwrite %s", outputLocation)
	default: // write files into a directory named by outputLocation
		if !isDirectory(outputLocation) {
			os.Mkdir(outputLocation, 0755)
		}
		for _, file := range response.Files {
			p := outputLocation + "/" + file.Name
			dir := path.Dir(p)
			os.MkdirAll(dir, 0755)
			f, _ := os.Create(p)
			defer f.Close()
			f.Write(file.Data)
		}
	}
	return nil
}

func isFile(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !fileInfo.IsDir()
}

func isDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}
