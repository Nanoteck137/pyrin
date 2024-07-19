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

type TestStruct struct {
	A int
	B float32
	C string
	D Inner
	E *Inner
	F []Inner
}

func TestExtract(t *testing.T) {
	c := extract.NewContext()

	err := c.ExtractType(TestStruct{})
	if err != nil {
		t.Errorf("Failed to extract type: %v", err)
	}

	err = c.ExtractType(testdata.Inner{})
	if err != nil {
		t.Errorf("Failed to extract type: %v", err)
	}

	pretty.Println(c)
}
