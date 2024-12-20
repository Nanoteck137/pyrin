package spec

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/nanoteck137/pyrin"
	"github.com/nanoteck137/pyrin/tools/extract"
	"github.com/nanoteck137/pyrin/tools/resolve"
	"github.com/nanoteck137/pyrin/utils"
)

func GenerateSpec(routes []Route) (*Server, error) {
	s := &Server{}

	resolver := resolve.New()

	c := extract.NewContext()

	for _, route := range routes {
		switch route := route.(type) {
		case ApiRoute:
			err := c.ExtractTypes(route.ReturnType)
			if err != nil {
				return nil, err
			}

			err = c.ExtractTypes(route.BodyType)
			if err != nil {
				return nil, err
			}
		case FormApiRoute:
			err := c.ExtractTypes(route.ReturnType)
			if err != nil {
				return nil, err
			}

			err = c.ExtractTypes(route.Spec.BodyType)
			if err != nil {
				return nil, err
			}
		case NormalRoute:
		default:
			panic(fmt.Sprintf("Unimplemented route type: %T", route))
		}
	}

	decls, err := c.ConvertToDecls()
	if err != nil {
		return nil, err
	}

	for _, decl := range decls {
		resolver.AddSymbolDecl(decl)
	}

	errorTypes := make(map[pyrin.ErrorType]struct{})

	addErrorTypes := func(types []pyrin.ErrorType) {
		for _, t := range types {
			errorTypes[t] = struct{}{}
		}
	}

	addErrorTypes(pyrin.GlobalErrors)

	getTypeName := func(ty any) (string, error) {
		if ty == nil {
			return "", nil
		}

		reflectedType := reflect.TypeOf(ty)
		name, err := c.TranslateName(reflectedType.Name(), reflectedType.PkgPath())
		if err != nil {
			return "", err
		}

		_, err = resolver.Resolve(name)
		if err != nil {
			return "", err
		}

		return name, nil
	}

	for _, route := range routes {
		switch route := route.(type) {
		case ApiRoute:
			addErrorTypes(route.ErrorTypes)

			responseType, err := getTypeName(route.ReturnType)
			if err != nil {
				return nil, err
			}

			bodyType, err := getTypeName(route.BodyType)
			if err != nil {
				return nil, err
			}

			s.ApiEndpoints = append(s.ApiEndpoints, ApiEndpoint{
				Name:         route.Name,
				Method:       route.Method,
				Path:         route.Path,
				ResponseType: responseType,
				BodyType:     bodyType,
			})
		case FormApiRoute:
			addErrorTypes(route.ErrorTypes)

			responseType, err := getTypeName(route.ReturnType)
			if err != nil {
				return nil, err
			}

			bodyType, err := getTypeName(route.Spec.BodyType)
			if err != nil {
				return nil, err
			}

			s.FormApiEndpoints = append(s.FormApiEndpoints, FormApiEndpoint{
				Name:         route.Name,
				Method:       route.Method,
				Path:         route.Path,
				ResponseType: responseType,
				BodyType:     bodyType,
			})
		case NormalRoute:
			s.NormalEndpoints = append(s.NormalEndpoints, NormalEndpoint{
				Name:   route.Name,
				Method: route.Method,
				Path:   route.Path,
			})
		default:
			panic(fmt.Sprintf("Unimplemented route type: %T", route))
		}
	}

	presentErrorType := make([]string, 0, len(errorTypes))

	for errType := range errorTypes {
		presentErrorType = append(presentErrorType, errType.String())
	}

	sort.Strings(presentErrorType)

	s.ErrorTypes = presentErrorType

	for _, st := range resolver.ResolvedStructs {
		switch t := st.Type.(type) {
		case *resolve.TypeStruct:
			fields := make([]TypeField, 0, len(t.Fields))

			for _, f := range t.Fields {
				s, err := utils.TypeToString(f.Type)
				if err != nil {
					return nil, err
				}

				fields = append(fields, TypeField{
					Name: f.Name,
					Type: s,
					Omit: f.Optional,
				})
			}

			s.Types = append(s.Types, Type{
				Name:   st.Name,
				Extend: "",
				Fields: fields,
			})
		case *resolve.TypeSameStruct:
			s.Types = append(s.Types, Type{
				Name:   st.Name,
				Extend: t.Type.Name,
			})
		}
	}

	return s, nil
}
