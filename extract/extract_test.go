package extract_test

import (
	"testing"

	"github.com/kr/pretty"
	"github.com/nanoteck137/pyrin/extract"
	"github.com/nanoteck137/pyrin/testdata"
)

type Inner struct {
	A bool
	B *string
	C []float64
}

type Inner2 struct {
	A int
}

type TestStruct struct {
	Inner
	A int `json:"a"`
	B float32 `json:"b,omitempty"`
	c string
}

func TestExtract(t *testing.T) {
	c := extract.NewContext()

	err := c.ExtractTypes(TestStruct{})
	if err != nil {
		t.Errorf("Failed to extract type: %v", err)
	}

	err = c.ExtractTypes(testdata.Inner{})
	if err != nil {
		t.Errorf("Failed to extract type: %v", err)
	}

	pretty.Println(c)

	decls, err := c.ConvertToDecls()
	if err != nil {
		t.Errorf("Failed to convert to decls: %v", err)
	}

	pretty.Println(decls)
}
