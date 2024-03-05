package resolve

import (
	"fmt"
)

type SymbolState int

const (
	SymbolUnresolved SymbolState = iota
	SymbolResolving
	SymbolResolved
)

type TypespecArray struct {
	Element any
}

type TypespecIdent struct {
	Ident string
}

type TypeString struct{}
type TypeInt struct{}
type TypeArray struct {
	ElementType any
}

type Field struct {
	Name string
	Type any
}

type TypeStruct struct {
	Name string
	Fields []Field
}

type FieldDecl struct {
	Name string
	Typespec any
}

type StructDecl struct {
	Name string
	Fields []FieldDecl
}


type Struct struct {
	State SymbolState
	Decl StructDecl
	Type  any
}

type Resolver struct {
	Types map[string]any

	Structs map[string]*Struct
	ResolvedStructs []*Struct
}

func New() *Resolver {
	resolver := &Resolver{
		Types:   map[string]any{},
		Structs: make(map[string]*Struct),
	}

	resolver.Types["int"] = TypeInt{}
	resolver.Types["string"] = TypeString{}

	return resolver
}

func (resolver *Resolver) ResolveTypespec(typespec any) (any, error) {
	switch t := typespec.(type) {
	case TypespecIdent:
		return resolver.Resolve(t.Ident)
	case TypespecArray:
		elementTy, err := resolver.ResolveTypespec(t.Element)
		if err != nil {
			return nil, err
		}

		return TypeArray{
			ElementType: elementTy,
		}, nil
	}

	return nil, nil
}

func (resolver *Resolver) ResolveField(field FieldDecl) (Field, error) {
	ty, err := resolver.ResolveTypespec(field.Typespec)
	if err != nil {
		return Field{}, err
	}

	return Field{
		Name: field.Name,
		Type: ty,
	}, nil
}

func (resolver *Resolver) Resolve(name string) (any, error) {
	if ty, exists := resolver.Types[name]; exists {
		return ty, nil
	}

	s, exists := resolver.Structs[name]
	if !exists {
		return nil, fmt.Errorf("'%s' don't exist", name)
	}

	if s.State == SymbolResolved {
		return s.Type, nil
	}

	if s.State == SymbolResolving {
		return nil, fmt.Errorf("'%s' cyclic dependency", name)
	}

	s.State = SymbolResolving

	var fields []Field

	for _, f := range s.Decl.Fields {
		f, err := resolver.ResolveField(f)
		if err != nil {
			return nil, err
		}

		fields = append(fields, f)
	}

	s.State = SymbolResolved
	s.Type = TypeStruct{
		Name: s.Decl.Name,
		Fields: fields,
	}

	resolver.ResolvedStructs = append(resolver.ResolvedStructs, s)

	return s.Type, nil
}

func (resolver *Resolver) ResolveAll() error {
	for n := range resolver.Structs {
		_, err := resolver.Resolve(n)
		if err != nil {
			return err
		}
	}

	return nil
}

//
// // func ResolveConfig(config *Config) ([]*ResolvedStruct, error) {
// // 	resolver := &Resolver{
// // 		config: config,
// // 	}
// //
// // 	for _, s := range config.Structs {
// // 		_, err := resolver.ResolveStruct(s.Name)
// // 		if err != nil {
// // 			return nil, err
// // 		}
// // 	}
// //
// // 	return resolver.ResolvedStructs, nil
// // }
