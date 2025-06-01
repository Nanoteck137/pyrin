package spark

import (
	"errors"
	"fmt"
)

type FieldType interface {
	typeType()
}

type FieldTypeString struct{}
type FieldTypeInt struct{}
type FieldTypeFloat struct{}
type FieldTypeBoolean struct{}
type FieldTypeArray struct {
	ElementType FieldType
}
type FieldTypePtr struct {
	BaseType FieldType
}

type FieldTypeStructRef struct {
	Name string
}

func (t *FieldTypeString) typeType()    {}
func (t *FieldTypeInt) typeType()       {}
func (t *FieldTypeFloat) typeType()     {}
func (t *FieldTypeBoolean) typeType()   {}
func (t *FieldTypeArray) typeType()     {}
func (t *FieldTypePtr) typeType()       {}
func (t *FieldTypeStructRef) typeType() {}

type ResolvedField struct {
	Name      string
	Type      FieldType
	OmitEmpty bool
}

type ResolvedStruct struct {
	Name   string
	Fields []ResolvedField
}

type SymbolState int

const (
	SymbolUnresolved SymbolState = iota
	SymbolResolving
	SymbolResolved
)

type Symbol struct {
	State SymbolState
	Name  string

	Decl           StructDecl
	ResolvedStruct *ResolvedStruct
}

type Resolver struct {
	Symbols         []*Symbol
	ResolvedSymbols []*Symbol
}

func NewResolver() *Resolver {
	return &Resolver{}
}

var intType = &FieldTypeInt{}
var floatType = &FieldTypeFloat{}
var stringType = &FieldTypeString{}
var boolType = &FieldTypeBoolean{}

func (resolver *Resolver) resolveTypespecBase(typespec Typespec, isFromPointer bool) (FieldType, error) {
	switch t := typespec.(type) {
	case *IdentTypespec:
		switch t.Ident {
		case "int":
			return intType, nil
		case "float":
			return floatType, nil
		case "string":
			return stringType, nil
		case "bool":
			return boolType, nil
		default:
			_, err := resolver.Resolve(t.Ident)
			if err != nil {
				return nil, err
			}

			return &FieldTypeStructRef{
				Name: t.Ident,
			}, nil
		}
	case *ArrayTypespec:
		elementTy, err := resolver.resolveTypespecBase(t.Element, false)
		if err != nil {
			return nil, err
		}

		return &FieldTypeArray{
			ElementType: elementTy,
		}, nil
	case *PtrTypespec:
		if isFromPointer {
			return nil, errors.New("single pointer is only allowed")
		}

		baseTy, err := resolver.resolveTypespecBase(t.Base, true)
		if err != nil {
			return nil, err
		}

		return &FieldTypePtr{
			BaseType: baseTy,
		}, nil
	default:
		// TODO(patrik): Better error
		panic("Unknown typespec")
	}
}

func (resolver *Resolver) resolveTypespec(typespec Typespec) (FieldType, error) {
	return resolver.resolveTypespecBase(typespec, false)
}

func (resolver *Resolver) ResolveField(field *FieldDecl) (ResolvedField, error) {
	ty, err := resolver.resolveTypespec(field.Type)
	if err != nil {
		return ResolvedField{}, err
	}

	return ResolvedField{
		Name:      field.Name,
		Type:      ty,
		OmitEmpty: field.OmitEmpty,
	}, nil
}

func (resolver *Resolver) getSymbol(name string) *Symbol {
	for _, s := range resolver.Symbols {
		if s.Name == name {
			return s
		}
	}

	return nil
}

func (resolver *Resolver) resolveExtendedStruct(decl *StructDecl) (*ResolvedStruct, error) {
	s, err := resolver.Resolve(decl.Extend)
	if err != nil {
		return nil, err
	}

	fields := make([]ResolvedField, len(s.Fields))
	copy(fields, s.Fields)

	for _, df := range decl.Fields {
		found := false
		index := 0
		for i, f := range fields {
			if df.Name == f.Name {
				found = true
				index = i
				break
			}
		}

		ty, err := resolver.resolveTypespec(df.Type)
		if err != nil {
			return nil, err
		}

		if found {
			fields[index].Type = ty
		} else {
			fields = append(fields, ResolvedField{
				Name:      df.Name,
				Type:      ty,
				OmitEmpty: df.OmitEmpty,
			})
		}
	}

	return &ResolvedStruct{
		Name:   decl.Name,
		Fields: fields,
	}, nil
}

func (resolver *Resolver) resolveStruct(decl *StructDecl) (*ResolvedStruct, error) {
	if decl.Extend != "" {
		return resolver.resolveExtendedStruct(decl)
	}

	var fields []ResolvedField
	for _, f := range decl.Fields {
		f, err := resolver.ResolveField(f)
		if err != nil {
			return nil, err
		}

		fields = append(fields, f)
	}

	return &ResolvedStruct{
		Name:   decl.Name,
		Fields: fields,
	}, nil
}

func (resolver *Resolver) Resolve(name string) (*ResolvedStruct, error) {
	s := resolver.getSymbol(name)
	if s == nil {
		return nil, fmt.Errorf("symbol not found: %s", name)
	}

	if s.State == SymbolResolved {
		return s.ResolvedStruct, nil
	}

	if s.State == SymbolResolving {
		return nil, fmt.Errorf("cyclic dependency found: %s", name)
	}

	s.State = SymbolResolving

	var err error
	s.ResolvedStruct, err = resolver.resolveStruct(&s.Decl)
	if err != nil {
		s.State = SymbolUnresolved
		return nil, err
	}

	s.State = SymbolResolved

	resolver.ResolvedSymbols = append(resolver.ResolvedSymbols, s)

	return s.ResolvedStruct, nil
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

func (resolver *Resolver) AddStructDecl(decl StructDecl) {
	resolver.Symbols = append(resolver.Symbols, &Symbol{
		Name: decl.Name,
		Decl: decl,
	})
}
