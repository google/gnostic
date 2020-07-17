package linenumbers

import (
	"testing"
)

//TestFindLineNumbers runs unit tests on the linenumbers package
func TestFindLineNumbers(t *testing.T) {
	keys := []string{"paths", "/alerts?lat={lat}\u0026lon={lon}", "get", "parameters", "1", "name"}
	token := "lon_100"
	file := "../../examples/v2.0/yaml/weatherbit.yaml"
	result, err := FindYamlLine(file, keys, token)
	if err != nil {
		t.Errorf("%+v\n", err)
	}
	if result.Line != 38 {
		t.Errorf("Given token \"lon_100\", FindYamlLine() returned %d, expected 38", result.Line)
	}

	keys = []string{"paths", "/alerts?lat={lat}\u0026lon={lon}", "get", "parameters", "0", "name"}
	token = "latSour"
	result, err = FindYamlLine(file, keys, token)
	if err != nil {
		t.Errorf("%+v\n", err)
	}
	if result.Line != 32 {
		t.Errorf("Given token \"latSour\", FindYamlLine() returned %d, expected 32", result.Line)
	}
}
