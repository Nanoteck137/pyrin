package resolve

import (
	"fmt"

	"github.com/nanoteck137/servine/ast"
)

type SymbolState int

const (
	SymbolUnresolved SymbolState = iota
	SymbolResolving
	SymbolResolved
)

type Type interface {
	typeType()
}

type TypeString struct{}
type TypeInt struct{}
type TypeBoolean struct{}
type TypeArray struct {
	ElementType Type
}

type Field struct {
	Name string
	Type Type
}

type TypeStruct struct {
	Name   string
	Fields []Field
}

func (t *TypeString) typeType()  {}
func (t *TypeInt) typeType()     {}
func (t *TypeBoolean) typeType() {}
func (t *TypeArray) typeType()   {}
func (t *TypeStruct) typeType()  {}

type Symbol struct {
	State SymbolState
	Name  string
	Decl  ast.Decl
	Type  Type
}

type Resolver struct {
	Symbols         []*Symbol
	ResolvedStructs []*Symbol
}

func New() *Resolver {
	resolver := &Resolver{}

	resolver.AddType("int", &TypeInt{})
	resolver.AddType("string", &TypeString{})
	resolver.AddType("bool", &TypeBoolean{})

	return resolver
}

func (resolver *Resolver) ResolveTypespec(typespec any) (Type, error) {
	switch t := typespec.(type) {
	case *ast.IdentTypespec:
		return resolver.Resolve(t.Ident)
	case *ast.ArrayTypespec:
		elementTy, err := resolver.ResolveTypespec(t.Element)
		if err != nil {
			return nil, err
		}

		return &TypeArray{
			ElementType: elementTy,
		}, nil
	default:
		panic("Unknown typespec")
	}
}

func (resolver *Resolver) ResolveField(field *ast.Field) (Field, error) {
	ty, err := resolver.ResolveTypespec(field.Type)
	if err != nil {
		return Field{}, err
	}

	return Field{
		Name: field.Name,
		Type: ty,
	}, nil
}

func (resolver *Resolver) GetSymbol(name string) (*Symbol, bool) {
	for _, s := range resolver.Symbols {
		if s.Name == name {
			return s, true
		}
	}

	return nil, false
}

func (resolver *Resolver) Resolve(name string) (Type, error) {
	s, exists := resolver.GetSymbol(name)
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

	switch decl := s.Decl.(type) {
	case *ast.StructDecl:
		if decl.Extend != "" {
			ty, err := resolver.Resolve(decl.Extend)
			if err != nil {
				return nil, err
			}

			// TODO(patrik): Resolve fields

			s.State = SymbolResolved
			s.Type = ty
		} else {
			var fields []Field
			for _, f := range decl.Fields {
				f, err := resolver.ResolveField(f)
				if err != nil {
					return nil, err
				}

				fields = append(fields, f)
			}

			s.State = SymbolResolved
			s.Type = &TypeStruct{
				Name:   decl.Name,
				Fields: fields,
			}
		}
	}

	resolver.ResolvedStructs = append(resolver.ResolvedStructs, s)

	return s.Type, nil
}

func (resolver *Resolver) ResolveAll() error {
	for _, s := range resolver.Symbols {
		_, err := resolver.Resolve(s.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

func (resolver *Resolver) AddSymbol(sym *Symbol) {
	resolver.Symbols = append(resolver.Symbols, sym)
}

func (resolver *Resolver) AddSymbolDecl(decl ast.Decl) {
	switch decl := decl.(type) {
	case *ast.StructDecl:
		resolver.AddSymbol(&Symbol{
			Name: decl.Name,
			Decl: decl,
		})
	}
}

func (resolver *Resolver) AddType(name string, ty Type) {
	resolver.AddSymbol(&Symbol{
		State: SymbolResolved,
		Name:  name,
		Type:  ty,
	})
}
