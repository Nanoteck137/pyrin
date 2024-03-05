package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/kr/pretty"
	"github.com/nanoteck137/pyrin/gen"
	"github.com/nanoteck137/pyrin/gen/gogen"
	"github.com/nanoteck137/pyrin/gen/jsgen"
	"github.com/nanoteck137/pyrin/resolve"
)

type FieldDef struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type StructDef struct {
	Name   string     `json:"name"`
	Fields []FieldDef `json:"fields"`
}

type Config struct {
	Structs []StructDef `json:"structs"`
}

func parseFieldType(ty string) (any, error) {
	ty = strings.TrimSpace(ty)
	switch ty {
	case "string", "str":
		return resolve.TypespecIdent{
			Ident: "string",
		}, nil
	case "int", "number":
		return resolve.TypespecIdent{
			Ident: "int",
		}, nil
	default:
		if strings.HasPrefix(ty, "array") {
			splits := strings.SplitN(ty, " ", 2)

			element, err :=parseFieldType(splits[1])
			if err != nil {
				return nil, err
			}

			return resolve.TypespecArray{
				Element: element,
			}, nil
		}

		return resolve.TypespecIdent{
			Ident: ty,
		},nil
	}
}

func main() {
	data, err := os.ReadFile("./test.json")
	if err != nil {
		log.Fatal(err)
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatal(err)
	}

	pretty.Println(config)
	resolver := resolve.New()

	for _, s := range config.Structs {
		fields := make([]resolve.FieldDecl, len(s.Fields))

		for i, f := range s.Fields {
			ty, err := parseFieldType(f.Type)
			if err != nil {
				log.Fatal(err)
			}

			fields[i] = resolve.FieldDecl{
				Name:     f.Name,
				Typespec: ty,
			}
		}

		resolver.Structs[s.Name] = &resolve.Struct{
			Decl:  resolve.StructDecl{
				Name:   s.Name,
				Fields: fields,
			},
		}
	}

	pretty.Println(resolver)

	stack := gen.GeneratorStack{}
	stack.AddGenerator(gogen.New(gogen.Options{
		Output: "./work/types/types.go",
	}))
	stack.AddGenerator(jsgen.New(jsgen.Options{
		Output: "./work/types.ts",
	}))

	// resolver.Structs["ApiGetArtists"] = &resolve.Struct{
	// 	Decl: resolve.StructDecl{
	// 		Name: "ApiGetArtists",
	// 		Fields: []resolve.FieldDecl{
	// 			{
	// 				Name: "artists",
	// 				Typespec: resolve.TypespecArray{
	// 					Element: resolve.TypespecIdent{
	// 						Ident: "ApiArtist",
	// 					},
	// 				},
	// 			},
	// 		},
	// 	},
	// }
	//
	// resolver.Structs["ApiArtist"] = &resolve.Struct{
	// 	Decl: resolve.StructDecl{
	// 		Name: "ApiArtist",
	// 		Fields: []resolve.FieldDecl{
	// 			{
	// 				Name: "id",
	// 				Typespec: resolve.TypespecIdent{
	// 					Ident: "string",
	// 				},
	// 			},
	// 			{
	// 				Name: "name",
	// 				Typespec: resolve.TypespecIdent{
	// 					Ident: "string",
	// 				},
	// 			},
	// 			{
	// 				Name: "picture",
	// 				Typespec: resolve.TypespecIdent{
	// 					Ident: "string",
	// 				},
	// 			},
	// 		},
	// 	},
	// }

	err = resolver.ResolveAll()
	if err != nil {
		log.Fatal(err)
	}

	pretty.Println(resolver.ResolvedStructs)

	err = stack.Generate(resolver)
	if err != nil {
		log.Fatal(err)
	}
}
