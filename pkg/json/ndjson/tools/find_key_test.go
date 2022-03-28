package tools

import (
	"testing"

	"github.com/cagnosolutions/go-data/pkg/json/ndjson/tools/jsonv1/scanner"
)

var fileName = "small_file.json"

func Test_Tools_JsonV1_FindKey(t *testing.T) {
	data := LoadFileData(fileName)
	bfound, err := scanner.FindKey(data, 0, []byte("secret-sauce"))
	if err != nil {
		t.Error(err)
	}
	expected := `"the number is 42!"`
	if string(bfound) != expected {
		t.Errorf("expected=%v, got=%s\n", expected, bfound)
	}
}

func Test_Tools_JsonV1_FindKeyV2(t *testing.T) {
	data := LoadFileData(fileName)
	bfound, err := scanner.FindKey2(data, []byte("secret-sauce1"))
	if err != nil {
		t.Error(err)
	}
	expected := `"the number is 42!"`
	if string(bfound) != expected {
		t.Errorf("expected=%v, got=%s\n", expected, bfound)
	}
}
