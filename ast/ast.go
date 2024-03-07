package ast

type Decl interface {
	declType()
}

type Typespec interface {
	typespecType()
}

type IdentTypespec struct {
	Ident string
}

type ArrayTypespec struct {
	Element Typespec
}

func (ty *IdentTypespec) typespecType() {}
func (ty *ArrayTypespec) typespecType() {}

type Field struct {
	Name  string
	Type  Typespec
	Unset bool
}

type StructDecl struct {
	Name   string
	Extend string
	Fields []*Field
}

func (decl *StructDecl) declType() {}
