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
	"github.com/nanoteck137/pyrin/gen/zod"
	"github.com/nanoteck137/pyrin/parser"
	"github.com/nanoteck137/pyrin/resolve"
	"github.com/spf13/cobra"
)

var genCmd = &cobra.Command{
	Use: "gen",
}

var genZodCmd = &cobra.Command{
	Use:  "zod <SERVER_CONFIG>",
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
		err = zod.GenerateTypeCode(buf, resolver)
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
		zod.GenerateClientCode(buf, &server)

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

func init() {
	genZodCmd.Flags().StringP("output", "o", "./src/api", "Output directory")
	genCmd.AddCommand(genZodCmd)

	rootCmd.AddCommand(genCmd)
}
