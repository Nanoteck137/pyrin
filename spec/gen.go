package spec

import (
	"reflect"
	"sort"

	"github.com/nanoteck137/pyrin/tools/extract"
	"github.com/nanoteck137/pyrin/tools/resolve"
	"github.com/nanoteck137/pyrin/utils"
)

func GenerateSpec(routes []Route) (*Server, error) {
	s := &Server{}

	resolver := resolve.New()

	c := extract.NewContext()

	for _, route := range routes {
		if route.Data != nil {
			c.ExtractTypes(route.Data)
		}

		if route.Body != nil {
			c.ExtractTypes(route.Body)
		}
	}

	decls, err := c.ConvertToDecls()
	if err != nil {
		return nil, err
	}

	for _, decl := range decls {
		resolver.AddSymbolDecl(decl)
	}

	for _, route := range routes {
		responseType := ""
		bodyType := ""

		if route.Data != nil {
			t := reflect.TypeOf(route.Data)

			name, err := c.TranslateName(t.Name(), t.PkgPath())
			if err != nil {
				return nil, err
			}

			_, err = resolver.Resolve(name)
			if err != nil {
				return nil, err
			}

			responseType = name
		}

		if route.Body != nil {
			t := reflect.TypeOf(route.Body)

			name, err := c.TranslateName(t.Name(), t.PkgPath())
			if err != nil {
				return nil, err
			}

			_, err = resolver.Resolve(name)
			if err != nil {
				return nil, err
			}

			bodyType = name
		}

		types := make(map[ErrorType]struct{})

		for _, t := range globalErrors {
			types[t] = struct{}{}
		}

		for _, t := range route.ErrorTypes {
			types[t] = struct{}{}
		}

		errorTypes := make([]string, 0, len(types))

		for k := range types {
			errorTypes = append(errorTypes, string(k))
		}

		sort.Strings(errorTypes)

		s.Endpoints = append(s.Endpoints, Endpoint{
			Name:            route.Name,
			Method:          route.Method,
			Path:            route.Path,
			ErrorTypes:      errorTypes,
			ResponseType:    responseType,
			BodyType:        bodyType,
			RequireFormData: route.RequireForm,
		})
	}

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
