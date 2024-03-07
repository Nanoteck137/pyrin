package main

import (
	"log"
	"os"

	"github.com/kr/pretty"
	"github.com/nanoteck137/pyrin/gen"
	"github.com/nanoteck137/pyrin/gen/gogen"
	"github.com/nanoteck137/pyrin/gen/jsgen"
	"github.com/nanoteck137/pyrin/parser"
	"github.com/nanoteck137/pyrin/resolve"
)

func main() {
	file, err := os.Open("./test.pyrin")
	if err != nil {
		log.Fatal(err)
	}

	parser := parser.New(file)
	decls := parser.Parse()
	pretty.Println(decls)

	resolver := resolve.New()

	for _, decl := range decls {
		resolver.AddSymbolDecl(decl)
		// switch decl := decl.(type) {
		// case *ast.StructDecl:
		// 	resolver
		// 	// resolver.Symbols[decl.Name] = &resolve.Symbol{
		// 	// 	Decl:  decl,
		// 	// };
		// }
	}

	err = resolver.ResolveAll()
	if err != nil {
		log.Fatal(err)
	}

	// pretty.Println(resolver)

	stack := gen.GeneratorStack{}
	stack.AddGenerator(gogen.New(gogen.Options{
		Output: "./work/types/types.go",
	}))
	stack.AddGenerator(jsgen.New(jsgen.Options{
		Output: "./work/types.ts",
	}))

	stack.Generate(resolver)
	//
	// // resolver.Structs["ApiGetArtists"] = &resolve.Struct{
	// // 	Decl: resolve.StructDecl{
	// // 		Name: "ApiGetArtists",
	// // 		Fields: []resolve.FieldDecl{
	// // 			{
	// // 				Name: "artists",
	// // 				Typespec: resolve.TypespecArray{
	// // 					Element: resolve.TypespecIdent{
	// // 						Ident: "ApiArtist",
	// // 					},
	// // 				},
	// // 			},
	// // 		},
	// // 	},
	// // }
	// //
	// // resolver.Structs["ApiArtist"] = &resolve.Struct{
	// // 	Decl: resolve.StructDecl{
	// // 		Name: "ApiArtist",
	// // 		Fields: []resolve.FieldDecl{
	// // 			{
	// // 				Name: "id",
	// // 				Typespec: resolve.TypespecIdent{
	// // 					Ident: "string",
	// // 				},
	// // 			},
	// // 			{
	// // 				Name: "name",
	// // 				Typespec: resolve.TypespecIdent{
	// // 					Ident: "string",
	// // 				},
	// // 			},
	// // 			{
	// // 				Name: "picture",
	// // 				Typespec: resolve.TypespecIdent{
	// // 					Ident: "string",
	// // 				},
	// // 			},
	// // 		},
	// // 	},
	// // }
	//
	// err = resolver.ResolveAll()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// pretty.Println(resolver.ResolvedStructs)
	//
	// err = stack.Generate(resolver)
	// if err != nil {
	// 	log.Fatal(err)
	// }
}
