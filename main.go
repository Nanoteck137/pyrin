package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
)

func main() {
	// cmd.Execute()

	fset := token.NewFileSet()

	src := `
	package test

	type TestStruct struct {
		Field1 string;
	}
	`

	f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	for _, d := range f.Decls {
		switch d := d.(type) {
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch spec := spec.(type) {
				case *ast.TypeSpec:
					fmt.Printf("spec.Name.Name: %v\n", spec.Name.Name)
				}
			}
			_ = d
		default:
		}
	}
}
