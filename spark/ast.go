package spark

type Typespec interface {
	typespecType()
}

type IdentTypespec struct {
	Ident string
}

type ArrayTypespec struct {
	Element Typespec
}

type PtrTypespec struct {
	Base Typespec
}

func (ty *IdentTypespec) typespecType() {}
func (ty *ArrayTypespec) typespecType() {}
func (ty *PtrTypespec) typespecType()   {}

type FieldDecl struct {
	Name      string
	Type      Typespec
	OmitEmpty bool
}

type StructDecl struct {
	Name   string
	Extend string
	Fields []*FieldDecl
}
