package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	goparser "go/parser"
	"log"
	"os"
	"path"

	"github.com/nanoteck137/pyrin/ast"
	"github.com/nanoteck137/pyrin/base"
	"github.com/nanoteck137/pyrin/client"
	"github.com/nanoteck137/pyrin/gen/gog"
	"github.com/nanoteck137/pyrin/gen/tsg"
	"github.com/nanoteck137/pyrin/parser"
	"github.com/nanoteck137/pyrin/resolve"
	"github.com/spf13/cobra"
)

var genTsCmd = &cobra.Command{
	Use:  "ts <SERVER_CONFIG>",
	Short: "Generate Typescript code",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		input := args[0]
		output, _ := cmd.Flags().GetString("output")

		err := os.MkdirAll(output, 0755)
		if err != nil {
			log.Fatal(err)
		}

		d, err := os.ReadFile(input)
		if err != nil {
			log.Fatal(err)
		}

		// TODO(patrik): Add checks
		var server client.Server
		err = json.Unmarshal(d, &server)
		if err != nil {
			log.Fatal(err)
		}

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

		err = resolver.ResolveAll()
		if err != nil {
			log.Fatal(err)
		}

		buf := &bytes.Buffer{}
		err = tsg.GenerateTypeCode(buf, resolver)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%v\n", buf.String())

		p := path.Join(output, "types.ts")
		err = os.WriteFile(p, buf.Bytes(), 0644)
		if err != nil {
			log.Fatal(err)
		}

		buf = &bytes.Buffer{}
		tsg.GenerateClientCode(buf, &server)

		fmt.Printf("%v\n", buf.String())

		p = path.Join(output, "client.ts")
		err = os.WriteFile(p, buf.Bytes(), 0644)
		if err != nil {
			log.Fatal(err)
		}

		p = path.Join(output, "base-client.ts")
		err = os.WriteFile(p, []byte(base.BaseClientSource), 0644)
		if err != nil {
			log.Fatal(err)
		}
	},
}

var genGoCmd = &cobra.Command{
	Use:  "go <SERVER_CONFIG>",
	Short: "Generate Golang code",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		input := args[0]
		output, _ := cmd.Flags().GetString("output")

		err := os.MkdirAll(output, 0755)
		if err != nil {
			log.Fatal(err)
		}

		d, err := os.ReadFile(input)
		if err != nil {
			log.Fatal(err)
		}

		// TODO(patrik): Add checks
		var server client.Server
		err = json.Unmarshal(d, &server)
		if err != nil {
			log.Fatal(err)
		}

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

		err = resolver.ResolveAll()
		if err != nil {
			log.Fatal(err)
		}

		buf := &bytes.Buffer{}
		err = gog.GenerateTypeCode(buf, resolver)
		if err != nil {
			log.Fatal(err)
		}

		p := path.Join(output, "types.go")
		err = os.WriteFile(p, buf.Bytes(), 0644)
		if err != nil {
			log.Fatal(err)
		}

		buf = &bytes.Buffer{}
		gog.GenerateClientCode(buf, &server)

		p = path.Join(output, "client.go")
		err = os.WriteFile(p, buf.Bytes(), 0644)
		if err != nil {
			log.Fatal(err)
		}

		p = path.Join(output, "base.go")
		err = os.WriteFile(p, []byte(base.BaseGoClient), 0644)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	genTsCmd.Flags().StringP("output", "o", "./src/api", "Output directory")
	genGoCmd.Flags().StringP("output", "o", "./api", "Output directory")
	rootCmd.AddCommand(genTsCmd)
	rootCmd.AddCommand(genGoCmd)
}
