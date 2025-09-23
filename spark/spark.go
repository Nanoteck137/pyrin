package spark

import (
	"encoding/json"
	"fmt"
	goast "go/ast"
	goparser "go/parser"
	"os"
	"reflect"
	"sort"

	"github.com/maruel/natural"
)

type Generator interface {
	Generate(serverDef *ServerDef, resolver *Resolver, outputDir string) error
}

type StructFieldDef struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	OmitEmpty bool   `json:"omitEmpty"`
}

type StructDef struct {
	Name   string           `json:"name"`
	Fields []StructFieldDef `json:"fields"`
}

type EndpointType string

const (
	EndpointTypeApi    EndpointType = "api"
	EndpointTypeForm   EndpointType = "form"
	EndpointTypeNormal EndpointType = "normal"
)

type Endpoint struct {
	Type     EndpointType `json:"type"`
	Name     string       `json:"name"`
	Method   string       `json:"method"`
	Path     string       `json:"path"`
	Response string       `json:"response,omitempty"`
	Body     string       `json:"body,omitempty"`
	// TODO(patrik): Add form constrains
}

type ServerDefVersion int

const (
	ServerDefVersion1 ServerDefVersion = 1
)

type ServerDef struct {
	Version ServerDefVersion `json:"version"`

	Structures []StructDef `json:"structures"`
	Endpoints  []Endpoint  `json:"endpoints"`
}

func (s *ServerDef) SaveToFile(p string) error {
	d, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(p, d, 0644)
	if err != nil {
		return err
	}

	return err
}

func parseTypespecBase(ty goast.Expr) Typespec {
	switch ty := ty.(type) {
	case *goast.Ident:
		return &IdentTypespec{
			Ident: ty.Name,
		}
	case *goast.StarExpr:
		base := parseTypespecBase(ty.X)
		return &PtrTypespec{
			Base: base,
		}
	case *goast.ArrayType:
		element := parseTypespecBase(ty.Elt)
		return &ArrayTypespec{
			Element: element,
		}
	case *goast.MapType:
		key := parseTypespecBase(ty.Key)
		value := parseTypespecBase(ty.Value)

		return &MapTypespec{
			Key:   key,
			Value: value,
		}
	default:
		panic(fmt.Sprintf("parseTypespecBase: unknown type: %T", ty))
	}
}

func ParseTypespec(s string) (Typespec, error) {
	e, err := goparser.ParseExpr(s)
	if err != nil {
		return nil, err
	}

	return parseTypespecBase(e), nil
}

func fieldTypeToString(ty FieldType) (string, error) {
	switch ty := ty.(type) {
	case *FieldTypeString:
		return "string", nil
	case *FieldTypeInt:
		return "int", nil
	case *FieldTypeFloat:
		return "float", nil
	case *FieldTypeBoolean:
		return "bool", nil
	case *FieldTypeArray:
		s, err := fieldTypeToString(ty.ElementType)
		if err != nil {
			return "", err
		}
		return "[]" + s, nil
	case *FieldTypePtr:
		s, err := fieldTypeToString(ty.BaseType)
		if err != nil {
			return "", err
		}
		return "*" + s, nil
	case *FieldTypeStructRef:
		return ty.Name, nil
	case *FieldTypeMap:
		key, err := fieldTypeToString(ty.KeyType)
		if err != nil {
			return "", err
		}

		value, err := fieldTypeToString(ty.ValueType)
		if err != nil {
			return "", err
		}

		return "map[" + key + "]" + value, nil
	default:
		return "", fmt.Errorf("Unknown resolved type: %T", ty)
	}
}

type NameFilter map[string]bool

func (n *NameFilter) AddName(name string) {
	(*n)[name] = true
}

func (n *NameFilter) LoadDefault() {
	// NOTE(patrik): Javascript keywords
	// Taken from: https://www.w3schools.com/js/js_reserved.asp
	n.AddName("abstract")
	n.AddName("arguments")
	n.AddName("await*")
	n.AddName("boolean")
	n.AddName("break")
	n.AddName("byte")
	n.AddName("case")
	n.AddName("catch")
	n.AddName("char")
	n.AddName("class*")
	n.AddName("const*")
	n.AddName("continue")
	n.AddName("debugger")
	n.AddName("default")
	n.AddName("delete")
	n.AddName("do")
	n.AddName("double")
	n.AddName("else")
	n.AddName("enum*")
	n.AddName("eval")
	n.AddName("export*")
	n.AddName("extends*")
	n.AddName("false")
	n.AddName("final")
	n.AddName("finally")
	n.AddName("float")
	n.AddName("for")
	n.AddName("function")
	n.AddName("goto")
	n.AddName("if")
	n.AddName("implements")
	n.AddName("import*")
	n.AddName("in")
	n.AddName("instanceof")
	n.AddName("int")
	n.AddName("interface")
	n.AddName("let*")
	n.AddName("long")
	n.AddName("native")
	n.AddName("new")
	n.AddName("null")
	n.AddName("package")
	n.AddName("private")
	n.AddName("protected")
	n.AddName("public")
	n.AddName("return")
	n.AddName("short")
	n.AddName("static")
	n.AddName("super*")
	n.AddName("switch")
	n.AddName("synchronized")
	n.AddName("this")
	n.AddName("throw")
	n.AddName("throws")
	n.AddName("transient")
	n.AddName("true")
	n.AddName("try")
	n.AddName("typeof")
	n.AddName("var")
	n.AddName("void")
	n.AddName("volatile")
	n.AddName("while")
	n.AddName("with")
	n.AddName("yield")

	// NOTE(patrik): Golang keywords
	// Taken from: https://handhikayp.medium.com/golang-101-6-the-reserved-keywords-in-go-1c8ef12d0fbf
	n.AddName("break")
	n.AddName("case")
	n.AddName("chan")
	n.AddName("const")
	n.AddName("continue")
	n.AddName("default")
	n.AddName("defer")
	n.AddName("else")
	n.AddName("fallthrough")
	n.AddName("for")
	n.AddName("func")
	n.AddName("go")
	n.AddName("goto")
	n.AddName("if")
	n.AddName("import")
	n.AddName("interface")
	n.AddName("map")
	n.AddName("package")
	n.AddName("range")
	n.AddName("return")
	n.AddName("select")
	n.AddName("struct")
	n.AddName("switch")
	n.AddName("type")
	n.AddName("var")
	n.AddName("true")
	n.AddName("false")
	n.AddName("iota")
	n.AddName("nil")
	n.AddName("int")
	n.AddName("int8")
	n.AddName("int16")
	n.AddName("int32")
	n.AddName("int64")
	n.AddName("uint")
	n.AddName("uint8")
	n.AddName("uint16")
	n.AddName("uint32")
	n.AddName("uint64")
	n.AddName("uintptr")
	n.AddName("float32")
	n.AddName("float64")
	n.AddName("complex128")
	n.AddName("complex64")
	n.AddName("bool")
	n.AddName("byte")
	n.AddName("rune")
	n.AddName("string")
	n.AddName("error")
	n.AddName("make")
	n.AddName("len")
	n.AddName("cap")
	n.AddName("new")
	n.AddName("append")
	n.AddName("copy")
	n.AddName("close")
	n.AddName("delete")
	n.AddName("complex")
	n.AddName("real")
	n.AddName("imag")
	n.AddName("panic")
	n.AddName("recover")

	// NOTE(patrik): Dart keywords
	// Taken from: https://www.geeksforgeeks.org/dart-keywords/
	n.AddName("abstract")
	n.AddName("else")
	n.AddName("import")
	n.AddName("super")
	n.AddName("as")
	n.AddName("enum")
	n.AddName("in")
	n.AddName("switch")
	n.AddName("assert")
	n.AddName("export")
	n.AddName("interface")
	n.AddName("sync")
	n.AddName("async")
	n.AddName("extends")
	n.AddName("is")
	n.AddName("this")
	n.AddName("await")
	n.AddName("extension")
	n.AddName("library")
	n.AddName("throw")
	n.AddName("break")
	n.AddName("external")
	n.AddName("mixin")
	n.AddName("true")
	n.AddName("case")
	n.AddName("factory")
	n.AddName("new")
	n.AddName("try")
	n.AddName("class")
	n.AddName("final")
	n.AddName("catch")
	n.AddName("false")
	n.AddName("null")
	n.AddName("typedef")
	n.AddName("on")
	n.AddName("var")
	n.AddName("const")
	n.AddName("finally")
	n.AddName("operator")
	n.AddName("void")
	n.AddName("continue")
	n.AddName("for")
	n.AddName("part")
	n.AddName("while")
	n.AddName("covariant")
	n.AddName("Function")
	n.AddName("rethrow")
	n.AddName("with")
	n.AddName("default")
	n.AddName("get")
	n.AddName("return")
	n.AddName("yield")
	n.AddName("deferred")
	n.AddName("hide")
	n.AddName("set")
	n.AddName("do")
	n.AddName("if")
	n.AddName("show")
	n.AddName("dynamic")
	n.AddName("implements")
	n.AddName("static")
}

func CreateServerDef(router *Router, fieldNameFilter NameFilter) (ServerDef, error) {
	res := ServerDef{
		Version: ServerDefVersion1,
	}

	resolver := NewResolver()
	structRegistry := NewStructRegistry()

	for _, route := range router.Routes {
		switch route := route.(type) {
		case ApiRoute:
			err := structRegistry.Register(route.ResponseType)
			if err != nil {
				return ServerDef{}, err
			}

			err = structRegistry.Register(route.BodyType)
			if err != nil {
				return ServerDef{}, err
			}
		case FormApiRoute:
			err := structRegistry.Register(route.ResponseType)
			if err != nil {
				return ServerDef{}, err
			}

			err = structRegistry.Register(route.Spec.BodyType)
			if err != nil {
				return ServerDef{}, err
			}
		case NormalRoute:
		default:
			panic(fmt.Sprintf("Unimplemented route type: %T", route))
		}
	}

	decls, err := structRegistry.GetStructDecls()
	if err != nil {
		return ServerDef{}, err
	}

	for _, decl := range decls {
		for _, field := range decl.Fields {
			if fieldNameFilter[field.Name] {
				return ServerDef{}, fmt.Errorf("%s uses banned field name: %s", decl.Name, field.Name)
			}
		}
	}

	for _, decl := range decls {
		resolver.AddStructDecl(decl)
	}

	getTypeName := func(ty any) (string, error) {
		if ty == nil {
			return "", nil
		}

		reflectedType := reflect.TypeOf(ty)
		name, err := structRegistry.TranslateName(reflectedType)
		if err != nil {
			return "", err
		}

		_, err = resolver.Resolve(name)
		if err != nil {
			return "", err
		}

		return name, nil
	}

	for _, route := range router.Routes {
		switch route := route.(type) {
		case ApiRoute:
			responseType, err := getTypeName(route.ResponseType)
			if err != nil {
				return ServerDef{}, err
			}

			bodyType, err := getTypeName(route.BodyType)
			if err != nil {
				return ServerDef{}, err
			}

			res.Endpoints = append(res.Endpoints, Endpoint{
				Type:     EndpointTypeApi,
				Name:     route.Name,
				Method:   route.Method,
				Path:     route.Path,
				Response: responseType,
				Body:     bodyType,
			})
		case FormApiRoute:
			responseType, err := getTypeName(route.ResponseType)
			if err != nil {
				return ServerDef{}, err
			}

			bodyType, err := getTypeName(route.Spec.BodyType)
			if err != nil {
				return ServerDef{}, err
			}

			res.Endpoints = append(res.Endpoints, Endpoint{
				Type:     EndpointTypeForm,
				Name:     route.Name,
				Method:   route.Method,
				Path:     route.Path,
				Response: responseType,
				Body:     bodyType,
			})
		case NormalRoute:
			res.Endpoints = append(res.Endpoints, Endpoint{
				Type:   EndpointTypeNormal,
				Name:   route.Name,
				Method: route.Method,
				Path:   route.Path,
			})
		default:
			panic(fmt.Sprintf("Unimplemented route type: %T", route))
		}
	}

	for _, st := range resolver.Symbols {
		if st.State != SymbolResolved {
			continue
		}

		rs := st.ResolvedStruct

		fields := make([]StructFieldDef, 0, len(rs.Fields))

		for _, f := range rs.Fields {
			s, err := fieldTypeToString(f.Type)
			if err != nil {
				return ServerDef{}, err
			}

			fields = append(fields, StructFieldDef{
				Name:      f.Name,
				Type:      s,
				OmitEmpty: f.OmitEmpty,
			})
		}

		res.Structures = append(res.Structures, StructDef{
			Name:   st.Name,
			Fields: fields,
		})
	}

	sort.SliceStable(res.Structures, func(i, j int) bool {
		return natural.Less(res.Structures[i].Name, res.Structures[j].Name)
	})


	sort.SliceStable(res.Endpoints, func(i, j int) bool {
		return natural.Less(res.Endpoints[i].Name, res.Endpoints[j].Name)
	})

	return res, nil
}

func CreateResolverFromServerDef(s *ServerDef) (*Resolver, error) {
	resolver := NewResolver()

	// TODO(patrik): Handle better
	for _, t := range s.Structures {
		fields := make([]*FieldDecl, 0, len(t.Fields))

		for _, f := range t.Fields {
			t, err := ParseTypespec(f.Type)
			if err != nil {
				return nil, err
			}

			fields = append(fields, &FieldDecl{
				Name:      f.Name,
				Type:      t,
				OmitEmpty: f.OmitEmpty,
			})
		}

		resolver.AddStructDecl(StructDecl{
			Name:   t.Name,
			Fields: fields,
		})
	}

	err := resolver.ResolveAll()
	if err != nil {
		return nil, err
	}

	return resolver, nil
}
