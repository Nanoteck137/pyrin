package cmd

import (
	"bytes"
	"fmt"
	goparser "go/parser"
	"log"
	"net/http"

	"github.com/nanoteck137/pyrin/ast"
	"github.com/nanoteck137/pyrin/client"
	"github.com/nanoteck137/pyrin/gen/zod"
	"github.com/nanoteck137/pyrin/parser"
	"github.com/nanoteck137/pyrin/resolve"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use: "test",
	Run: func(cmd *cobra.Command, args []string) {
		server := client.Server{
			Types: []client.MetadataType{
				{
					Name:   "Track",
					Extend: "",
					Fields: []client.TypeField{
						{
							Name: "name",
							Type: "string",
							Omit: false,
						},
						{
							Name: "num",
							Type: "int",
							Omit: false,
						},
					},
				},
				{
					Name:   "GetTracks",
					Extend: "",
					Fields: []client.TypeField{
						{
							Name: "tracks",
							Type: "[]Track",
							Omit: false,
						},
					},
				},
				{
					Name:   "GetTrackById",
					Extend: "Track",
					Fields: []client.TypeField{
						{
							Name: "num",
							Type: "string",
							Omit: false,
						},
					},
				},
			},
			Endpoints: []client.Endpoint{
				{
					Name:         "GetTracks",
					Method:       http.MethodGet,
					Path:         "/api/v1/tracks",
					ResponseType: "GetTracks",
					BodyType:     "",
				},
				{
					Name:         "GetTrackById",
					Method:       http.MethodGet,
					Path:         "/api/v1/tracks/:id",
					ResponseType: "GetTrackById",
					BodyType:     "",
				},
			},
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

		err := resolver.ResolveAll()
		if err != nil {
			log.Fatal(err)
		}

		buf := &bytes.Buffer{}
		err = zod.GenerateTypeCode(buf, resolver)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%v\n", buf.String())

		buf = &bytes.Buffer{}
		zod.GenerateClientCode(buf, &server)

		fmt.Printf("%v\n", buf.String())
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
