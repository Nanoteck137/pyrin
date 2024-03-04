package gogen

import (
	"fmt"
	"io"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/nanoteck137/pyrin/resolve"
)

type GoGenerator struct {}

func (gen *GoGenerator) Name() string {
	return "GoGenerator"
}

func (gen *GoGenerator) Generate(resolver *resolve.Resolver) error {
	var b strings.Builder

	fmt.Fprintln(&b, "package types")
	fmt.Fprintln(&b)

	for _, s := range resolver.ResolvedStructs {
		fmt.Fprintf(&b, "type %s struct {\n", s.Decl.Name)

		st := s.Type.(resolve.TypeStruct)
		for _, f := range st.Fields {
			GenerateField(&b, &f)
		}

		fmt.Fprintln(&b, "}");
		fmt.Fprintln(&b)
	}

	fmt.Println(b.String())

	return nil
}

func GenerateType(typ any) string {
	switch t := typ.(type) {
	case resolve.TypeString:
		return "string"
	case resolve.TypeInt:
		return "int"
	case resolve.TypeStruct:
		return t.Name
	case resolve.TypeArray:
		return "[]"+GenerateType(t.ElementType)
	}

	return ""
}

func GenerateField(w io.Writer, field *resolve.Field) {
	jsonName := field.Name
	name := strcase.ToCamel(field.Name)

	fmt.Fprintf(w, "\t%s %s `json:\"%s\"`\n", name, GenerateType(field.Type), jsonName);
}
