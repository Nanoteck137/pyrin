package spec

import "github.com/nanoteck137/pyrin"

var _ pyrin.Router = (*Router)(nil)

type Router struct {
	Routes []Route
}

func (r *Router) AddRoute(route Route) {
	r.Routes = append(r.Routes, route)
}

func (r *Router) Group(prefix string) pyrin.Group {
	return NewRouteGroup(r, prefix)
}

type Route interface {
	routeType()
}

type ApiRoute struct {
	Name       string
	Path       string
	Method     string
	ErrorTypes []pyrin.ErrorType
	ReturnType any
	BodyType   any
}

func (r ApiRoute) routeType() {}

type FormApiRoute struct {
	Name       string
	Path       string
	Method     string
	ErrorTypes []pyrin.ErrorType
	ReturnType any
	Spec       pyrin.FormSpec
}

func (r FormApiRoute) routeType() {}

type NormalRoute struct {
	Name   string
	Path   string
	Method string
}

func (r NormalRoute) routeType() {}

// type Route struct {
// 	Name        string
// 	Path        string
// 	Method      string
// 	ErrorTypes  []pyrin.ErrorType
// 	ReturnType  any
// 	BodyType    any
// 	RequireForm bool
// }

type RouteGroup struct {
	Router *Router
	Prefix string
}

func NewRouteGroup(router *Router, prefix string) *RouteGroup {
	return &RouteGroup{
		Router: router,
		Prefix: prefix,
	}
}

func (r *RouteGroup) Register(handlers ...pyrin.Handler) {
	for _, h := range handlers {
		switch h := h.(type) {
		case pyrin.ApiHandler:
			r.Router.AddRoute(ApiRoute{
				Name:       h.Name,
				Path:       r.Prefix + h.Path,
				Method:     h.Method,
				ErrorTypes: h.Errors,
				ReturnType: h.ReturnType,
				BodyType:   h.BodyType,
			})
		case pyrin.FormApiHandler:
			r.Router.AddRoute(FormApiRoute{
				Name:       h.Name,
				Path:       r.Prefix + h.Path,
				Method:     h.Method,
				ErrorTypes: h.Errors,
				ReturnType: h.ReturnType,
				Spec:       h.Spec,
			})
		case pyrin.NormalHandler:
			r.Router.AddRoute(NormalRoute{
				Name:        h.Name,
				Path:        r.Prefix + h.Path,
				Method:      h.Method,
			})
		}
	}
}
