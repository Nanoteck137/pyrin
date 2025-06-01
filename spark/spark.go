package spark

import (
	"encoding/json"
	"fmt"
	goast "go/ast"
	goparser "go/parser"
	"os"
	"reflect"
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
	default:
		panic(fmt.Sprintf("parseTypespecBase: unknown type: %t", ty))
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
	default:
		return "", fmt.Errorf("Unknown resolved type: %T", ty)
	}
}

func CreateServerDef(router *Router) (ServerDef, error) {
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
				Name: f.Name,
				Type: t,
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
