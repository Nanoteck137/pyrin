package main

import (
	"errors"
	goast "go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"log"

	"github.com/kr/pretty"
	"github.com/nanoteck137/pyrin/ast"
	"github.com/nanoteck137/pyrin/resolve"
)

func main() {
	// cmd.Execute()

	fset := token.NewFileSet()

	src := `
	package test

	type TestStruct struct {
		Field1 string;
	}

	type TestStruct2 struct { 
		TestStruct

		Field2, Hello int;
	}
	`

	f, err := parser.ParseFile(fset, "types.go", src, parser.ParseComments)
	if err != nil {
		var errorList scanner.ErrorList
		if errors.As(err, &errorList) {
			pretty.Println(errorList)
			log.Fatal()
		} else {
			log.Fatal(err)
		}
	}

	var decls []ast.Decl

	for _, d := range f.Decls {
		switch d := d.(type) {
		case *goast.GenDecl:
			for _, spec := range d.Specs {
				switch spec := spec.(type) {
				case *goast.TypeSpec:
					pretty.Println(spec)

					switch ty := spec.Type.(type) {
					case *goast.StructType:
						var fields []*ast.Field
						extend := ""

						for _, field := range ty.Fields.List {
							if field.Names == nil {
								if extend == "" {
									if ident, ok := field.Type.(*goast.Ident); ok {
										extend = ident.Name
									}
								}

								continue
							}

							var ty ast.Typespec

							switch t := field.Type.(type) {
							case *goast.Ident:
								ty = &ast.IdentTypespec{
									Ident: t.Name,
								}
							}

							for _, name := range field.Names {

								fields = append(fields, &ast.Field{
									Name:  name.Name,
									Type:  ty,
									Unset: false,
								})
							}
						}

						decls = append(decls, &ast.StructDecl{
							Name:   spec.Name.Name,
							Extend: extend,
							Fields: fields,
						})
					}
				}
			}
		}
	}

	pretty.Println(decls)

	resolver := resolve.New()

	for _, decl := range decls {
		resolver.AddSymbolDecl(decl)
	}

	resolver.ResolveAll()

	pretty.Println(resolver.Symbols)
}
