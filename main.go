package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/kr/pretty"
	"github.com/nanoteck137/pyrin/gen"
	"github.com/nanoteck137/pyrin/resolve"
)

type Status int

const (
	StatusUnfinished Status = iota
	StatusProcessing
	StatusFinished
)

// type StructDecl struct {
// 	Name   string
// 	Status Status
// 	Fields []FieldDef
// }
//
// type Generator struct {
// 	Structs map[string]StructDecl
// }
//
// func (gen *Generator) GetType(w io.Writer, fieldType string) (string, error) {
// 	switch fieldType {
// 	case "string", "str":
// 		return "string", nil
// 	case "int", "number":
// 		return "int", nil
// 	default:
// 		splits := strings.Split(fieldType, " ")
// 		if len(splits) == 2 {
// 			if splits[0] == "array" {
// 				elementTy, err := gen.GenerateStruct(w, splits[1])
// 				if err != nil {
// 					return "", err
// 				}
//
// 				return "[]" + elementTy, nil
// 			}
// 		}
// 		return "", fmt.Errorf("Unknown field type '%s'", fieldType)
// 	}
// }
//
// func (gen *Generator) GenerateField(w io.Writer, field FieldDef) error {
// 	ty, err := gen.GetType(w, field.Typ)
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Fprintf(w, "\t%s %s\n", field.Name, ty)
//
// 	return nil
// }
//
// func (gen *Generator) GenerateStruct(w io.Writer, name string) (string, error) {
// 	decl, exists := gen.Structs[name]
// 	if !exists {
// 		return "", fmt.Errorf("'%s' don't exist", name)
// 	}
//
// 	if decl.Status == StatusProcessing {
// 		return "", fmt.Errorf("'%s' cyclic dependency", name)
// 	}
//
// 	if decl.Status == StatusFinished {
// 		return decl.Name, nil
// 	}
//
// 	decl.Status = StatusProcessing
//
// 	fmt.Fprintf(w, "type %s struct {\n", decl.Name)
// 	for _, f := range decl.Fields {
// 		err := gen.GenerateField(w, f)
// 		if err != nil {
// 			return "", err
// 		}
// 	}
// 	fmt.Fprintf(w, "}\n")
//
// 	decl.Status = StatusFinished
//
// 	return decl.Name, nil
// }
//
// func Generate(config Config) error {
// 	var b strings.Builder
//
// 	gen := Generator{
// 		Structs: map[string]StructDecl{},
// 	}
//
// 	for _, s := range config.Structs {
// 		gen.Structs[s.Name] = StructDecl{
// 			Name:   s.Name,
// 			Status: StatusUnfinished,
// 			Fields: s.Fields,
// 		}
// 	}
//
// 	pretty.Println(gen)
//
// 	for _, s := range config.Structs {
// 		_, err := gen.GenerateStruct(&b, s.Name)
// 		if err != nil {
// 			return err
// 		}
// 	}
//
// 	fmt.Println(b.String())
//
// 	return nil
// }

type FieldDef struct {
	Name string `json:"name"`
	Typ  string `json:"type"`
}

type StructDef struct {
	Name   string     `json:"name"`
	Fields []FieldDef `json:"fields"`
}

type Config struct {
	Structs []StructDef `json:"structs"`
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

	ty, err := resolver.Resolve("ApiGetArtists")
	if err != nil {
		log.Fatal(err)
	}

	pretty.Println(ty)

	pretty.Println(resolver.ResolvedStructs)

	s := gen.Generate(resolver)
	fmt.Println(s)
}
