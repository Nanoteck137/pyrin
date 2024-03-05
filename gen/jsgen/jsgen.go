package jsgen

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
		gen.options.Output = "./src/types/types.js"
	}

	return gen
}

func (gen *Generator) Name() string {
	return "JsGenerator"
}

func GenerateType(w io.Writer, ty any) {
	switch t := ty.(type) {
	case resolve.TypeString:
		fmt.Fprint(w, "z.string()")
	case resolve.TypeInt:
		fmt.Fprint(w, "z.number()")
	case resolve.TypeArray:
		fmt.Fprint(w, "z.array(")
		GenerateType(w, t.ElementType)
		fmt.Fprint(w, ")")
	case resolve.TypeStruct:
		fmt.Fprint(w, t.Name)
	}
}

func GenerateField(w io.Writer, field *resolve.Field) {
	fmt.Fprint(w, "  ", field.Name, ": ")
	GenerateType(w, field.Type)
	fmt.Fprintln(w, ",")
}

func (gen *Generator) Generate(resolver *resolve.Resolver) error {
	var b strings.Builder

	for _, s := range resolver.ResolvedStructs {
		fmt.Fprintf(&b, "export const %s = z.object({\n", s.Decl.Name)
		st := s.Type.(resolve.TypeStruct)
		for _, f := range st.Fields {
			GenerateField(&b, &f)
		}
		fmt.Fprintln(&b, "});")
		fmt.Fprintf(&b, "export type %s = z.infer<typeof %s>;\n", s.Decl.Name, s.Decl.Name)
		fmt.Fprintln(&b)
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
