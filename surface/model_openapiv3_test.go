package surface_v1

import (
	"os"
	"testing"

	openapiv3 "github.com/fern-api/protoc-gen-openapi/openapiv3"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestModelOpenAPIV3(t *testing.T) {
	refFile := "testdata/v3.0/petstore.json"
	modelFile := "testdata/v3.0/petstore.model.json"

	bFile, err := os.ReadFile(refFile)
	if err != nil {
		t.Logf("Failed to read file: %+v", err)
		t.FailNow()
	}
	bModel, err := os.ReadFile(modelFile)
	if err != nil {
		t.Logf("Failed to read file: %+v", err)
		t.FailNow()
	}

	docv3, err := openapiv3.ParseDocument(bFile)
	if err != nil {
		t.Logf("Failed to parse document: %+v", err)
		t.FailNow()
	}

	m, err := NewModelFromOpenAPI3(docv3, refFile)
	if err != nil {
		t.Logf("Failed to create model: %+v", err)
		t.FailNow()
	}

	var model Model
	if err := protojson.Unmarshal(bModel, &model); err != nil {
		t.Logf("Failed to unmarshal model: %+v", err)
		t.FailNow()
	}

	cmpOpts := []cmp.Option{
		protocmp.Transform(),
	}
	if diff := cmp.Diff(&model, m, cmpOpts...); diff != "" {
		t.Errorf("Model mismatch (-want +got):\n%s", diff)
	}
	x, _ := protojson.Marshal(m)
	t.Logf("Model: %s", x)
}
