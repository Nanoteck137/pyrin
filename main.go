package main

import (
	"log"

	"github.com/kr/pretty"
	"github.com/nanoteck137/pyrin/gen"
	"github.com/nanoteck137/pyrin/gen/gogen"
	"github.com/nanoteck137/pyrin/gen/jsgen"
	"github.com/nanoteck137/pyrin/resolve"
)

func main() {
	stack := gen.GeneratorStack{}
	stack.AddGenerator(gogen.New(gogen.Options{
		Output:      "./work/types/types.go",
	}))
	stack.AddGenerator(jsgen.New(jsgen.Options{
		Output: "./work/types.js",
	}))

	resolver := resolve.New()
	resolver.Structs["ApiGetArtists"] = &resolve.Struct{
		Decl: resolve.StructDecl{
			Name: "ApiGetArtists",
			Fields: []resolve.FieldDecl{
				{
					Name: "artists",
					Typespec: resolve.TypespecArray{
						Element: resolve.TypespecIdent{
							Ident: "ApiArtist",
						},
					},
				},
			},
		},
	}

	resolver.Structs["ApiArtist"] = &resolve.Struct{
		Decl: resolve.StructDecl{
			Name: "ApiArtist",
			Fields: []resolve.FieldDecl{
				{
					Name: "id",
					Typespec: resolve.TypespecIdent{
						Ident: "string",
					},
				},
				{
					Name: "name",
					Typespec: resolve.TypespecIdent{
						Ident: "string",
					},
				},
				{
					Name: "picture",
					Typespec: resolve.TypespecIdent{
						Ident: "string",
					},
				},
			},
		},
	}

	pretty.Println(resolver)

	resolver.ResolveAll()

	pretty.Println(resolver.ResolvedStructs)

	err := stack.Generate(resolver)
	if err != nil {
		log.Fatal(err)
	}
}
