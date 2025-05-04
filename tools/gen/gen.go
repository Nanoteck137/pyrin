package gen

import (
	"bytes"
	"encoding/json"
	"fmt"
	goparser "go/parser"
	"log"
	"os"
	"path"

	"github.com/nanoteck137/pyrin/spec"
	"github.com/nanoteck137/pyrin/tools/ast"
	"github.com/nanoteck137/pyrin/tools/gen/base"
	"github.com/nanoteck137/pyrin/tools/gen/dartg"
	"github.com/nanoteck137/pyrin/tools/gen/gog"
	"github.com/nanoteck137/pyrin/tools/gen/tsg"
	"github.com/nanoteck137/pyrin/tools/parser"
	"github.com/nanoteck137/pyrin/tools/resolve"
)

func ReadSpec(input string) (*spec.Server, error) {
	d, err := os.ReadFile(input)
	if err != nil {
		return nil, err
	}

	// TODO(patrik): Add checks
	var server spec.Server
	err = json.Unmarshal(d, &server)
	if err != nil {
		return nil, err
	}

	return &server, nil
}

func GenerateTypescript(server *spec.Server, output string) error {
	resolver := resolve.New()

	for _, t := range server.Types {
		fields := make([]*ast.Field, 0, len(t.Fields))

		for _, f := range t.Fields {
			e, err := goparser.ParseExpr(f.Type)
			if err != nil {
				return err
			}

			t := parser.ParseTypespec(e)

			fields = append(fields, &ast.Field{
				Name: f.Name,
				Type: t,
				Omit: f.Omit,
			})
		}

		resolver.AddSymbolDecl(&ast.StructDecl{
			Name:   t.Name,
			Extend: t.Extend,
			Fields: fields,
		})
	}

	err := resolver.ResolveAll()
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	err = tsg.GenerateTypeCode(buf, resolver)
	if err != nil {
		return err
	}

	fmt.Printf("%v\n", buf.String())

	p := path.Join(output, "types.ts")
	err = os.WriteFile(p, buf.Bytes(), 0644)
	if err != nil {
		return err
	}

	buf = &bytes.Buffer{}
	tsg.GenerateClientCode(buf, server)

	fmt.Printf("%v\n", buf.String())

	p = path.Join(output, "client.ts")
	err = os.WriteFile(p, buf.Bytes(), 0644)
	if err != nil {
		return err
	}

	p = path.Join(output, "base-client.ts")
	err = os.WriteFile(p, []byte(base.BaseClientSource), 0644)
	if err != nil {
		return err
	}

	return nil
}

func GenerateGolang(server *spec.Server, output string) error {
	resolver := resolve.New()

	for _, t := range server.Types {
		fields := make([]*ast.Field, 0, len(t.Fields))

		for _, f := range t.Fields {
			e, err := goparser.ParseExpr(f.Type)
			if err != nil {
				log.Fatal(err)
			}

			t := parser.ParseTypespec(e)

			fields = append(fields, &ast.Field{
				Name: f.Name,
				Type: t,
				Omit: f.Omit,
			})
		}

		resolver.AddSymbolDecl(&ast.StructDecl{
			Name:   t.Name,
			Extend: t.Extend,
			Fields: fields,
		})
	}

	err := resolver.ResolveAll()
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	err = gog.GenerateTypeCode(buf, resolver)
	if err != nil {
		return err
	}

	p := path.Join(output, "types.go")
	err = os.WriteFile(p, buf.Bytes(), 0644)
	if err != nil {
		return err
	}

	buf = &bytes.Buffer{}
	gog.GenerateClientCode(buf, server)

	p = path.Join(output, "client.go")
	err = os.WriteFile(p, buf.Bytes(), 0644)
	if err != nil {
		return err
	}

	p = path.Join(output, "base.go")
	err = os.WriteFile(p, []byte(base.BaseGoClient), 0644)
	if err != nil {
		return err
	}

	return nil
}

func GenerateDart(server *spec.Server, output string) error {
	resolver := resolve.New()

	for _, t := range server.Types {
		fields := make([]*ast.Field, 0, len(t.Fields))

		for _, f := range t.Fields {
			e, err := goparser.ParseExpr(f.Type)
			if err != nil {
				log.Fatal(err)
			}

			t := parser.ParseTypespec(e)

			fields = append(fields, &ast.Field{
				Name: f.Name,
				Type: t,
				Omit: f.Omit,
			})
		}

		resolver.AddSymbolDecl(&ast.StructDecl{
			Name:   t.Name,
			Extend: t.Extend,
			Fields: fields,
		})
	}

	err := resolver.ResolveAll()
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	err = dartg.GenerateTypeCode(buf, resolver)
	if err != nil {
		return err
	}

	p := path.Join(output, "types.dart")
	err = os.WriteFile(p, buf.Bytes(), 0644)
	if err != nil {
		return err
	}

	buf = &bytes.Buffer{}
	dartg.GenerateClientCode(buf, server)

	p = path.Join(output, "client.dart")
	err = os.WriteFile(p, buf.Bytes(), 0644)
	if err != nil {
		return err
	}

	p = path.Join(output, "base_client.dart")
	err = os.WriteFile(p, []byte(base.BaseDartClient), 0644)
	if err != nil {
		return err
	}

	return nil
}
