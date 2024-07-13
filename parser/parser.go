package parser

import (
	"errors"
	goast "go/ast"
	goparser "go/parser"
	"go/scanner"
	"go/token"
	"log"
	"strconv"

	"github.com/fatih/structtag"
	"github.com/kr/pretty"
	"github.com/nanoteck137/pyrin/ast"
)

func ParseTypespec(ty goast.Expr) ast.Typespec {
	switch ty := ty.(type) {
	case *goast.Ident:
		return &ast.IdentTypespec{
			Ident: ty.Name,
		}
	case *goast.StarExpr:
		base := ParseTypespec(ty.X)
		return &ast.PtrTypespec{
			Base: base,
		}
	case *goast.ArrayType:
		element := ParseTypespec(ty.Elt)
		return &ast.ArrayTypespec{
			Element: element,
		}
	}

	return nil
}

func parseStruct(name string, ty *goast.StructType) *ast.StructDecl {
	var fields []*ast.Field
	extend := ""

	for _, field := range ty.Fields.List {
		if field.Names == nil {
			// TODO(patrik): Add error when multiple embedded structs is detected
			if extend == "" {
				if ident, ok := field.Type.(*goast.Ident); ok {
					extend = ident.Name
				}
			}

			continue
		}

		jsonName := ""
		omit := false

		if field.Tag != nil {
			lit, err := strconv.Unquote(field.Tag.Value)
			if err != nil {
				// TODO(patrik): Remove
				log.Fatal(err)
			}

			tags, err := structtag.Parse(lit)
			if err != nil {
				// TODO(patrik): Remove
				log.Fatal(err)
			}

			tag, err := tags.Get("json")
			if err == nil {
				jsonName = tag.Name
				omit = tag.HasOption("omitempty")
			}
		}

		ty := ParseTypespec(field.Type)

		// TODO(patrik): Better errors
		if len(field.Names) > 1 {
			log.Fatal("More then one name for field")
		}

		name := field.Names[0].Name

		if jsonName != "" {
			name = jsonName
		}

		fields = append(fields, &ast.Field{
			Name: name,
			Type: ty,
			Omit: omit,
		})
	}

	return &ast.StructDecl{
		Name:   name,
		Extend: extend,
		Fields: fields,
	}
}

func Parse(source string) []ast.Decl {
	fset := token.NewFileSet()

	f, err := goparser.ParseFile(fset, "", source, goparser.ParseComments)
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
					switch ty := spec.Type.(type) {
					case *goast.StructType:
						decls = append(decls, parseStruct(spec.Name.Name, ty))
					}
				}
			}
		}
	}

	return decls
}
