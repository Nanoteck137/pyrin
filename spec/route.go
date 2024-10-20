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

type Route struct {
	Name        string
	Path        string
	Method      string
	ErrorTypes  []pyrin.ErrorType
	Data        any
	Body        any
	RequireForm bool
}

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
			r.Router.AddRoute(Route{
				Name:        h.Name,
				Path:        r.Prefix + h.Path,
				Method:      h.Method,
				ErrorTypes:  h.Errors,
				Data:        h.DataType,
				Body:        h.BodyType,
				RequireForm: h.RequireForm,
			})
		}
	}
}
