package zod

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/nanoteck137/pyrin/resolve"
)

type Options struct {
	Output string
}

type Generator struct {
	options Options
}

func New(options ...Options) *Generator {
	gen := &Generator{}

	if len(options) > 0 {
		gen.options = options[0]
	}

	if gen.options.Output == "" {
		gen.options.Output = "./src/types/types.ts"
	}

	return gen
}

func (gen *Generator) Name() string {
	return "Zod"
}

func GenerateType(w io.Writer, ty resolve.Type) {
	switch t := ty.(type) {
	case *resolve.TypeString:
		fmt.Fprint(w, "z.string()")
	case *resolve.TypeInt:
		fmt.Fprint(w, "z.number()")
	case *resolve.TypeBoolean:
		fmt.Fprint(w, "z.boolean()")
	case *resolve.TypeArray:
		fmt.Fprint(w, "z.array(")
		GenerateType(w, t.ElementType)
		fmt.Fprint(w, ")")
	case *resolve.TypePtr:
		GenerateType(w, t.BaseType)
		fmt.Fprint(w, ".nullable()")
	case *resolve.TypeStruct:
		fmt.Fprint(w, t.Name)
	}
}

func GenerateField(w io.Writer, field *resolve.Field) {
	fmt.Fprint(w, "  ", field.Name, ": ")
	GenerateType(w, field.Type)
	if field.Optional {
		fmt.Fprint(w, ".optional()")
	}
	fmt.Fprintln(w, ",")
}

func (gen *Generator) Generate(resolver *resolve.Resolver) error {
	var b strings.Builder

	fmt.Fprintln(&b, "// THIS FILE IS GENERATED BY PYRIN ZOD CODE GENERATOR")
	fmt.Fprintln(&b, `import { z } from "zod";`)
	fmt.Fprintln(&b)

	for _, s := range resolver.ResolvedStructs {

		switch ty := s.Type.(type) {
		case *resolve.TypeStruct:
			fmt.Fprintf(&b, "export const %s = z.object({\n", s.Name)
			for _, f := range ty.Fields {
				GenerateField(&b, &f)
			}
			fmt.Fprintln(&b, "});")
			fmt.Fprintf(&b, "export type %s = z.infer<typeof %s>;\n", s.Name, s.Name)
			fmt.Fprintln(&b)
		case *resolve.TypeSameStruct:
			fmt.Fprintf(&b, "export const %s = %s;\n", s.Name, ty.Type.Name)
			fmt.Fprintf(&b, "export type %s = z.infer<typeof %s>;\n", s.Name, s.Name)
			fmt.Fprintln(&b)
		}
	}

	dir := path.Dir(gen.options.Output)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(gen.options.Output, []byte(b.String()), 0644)
	if err != nil {
		return err
	}

	return nil
}
