package spark

import (
	"strings"

	"github.com/nanoteck137/pyrin"
)

var _ pyrin.Router = (*Router)(nil)

type Router struct {
	Routes []Route
}

func (r *Router) AddRoute(route Route) {
	r.Routes = append(r.Routes, route)
}

func (r *Router) Group(
	prefix string, 
	middlewares ...pyrin.MiddlewareFunc,
) pyrin.Group {
	return NewRouteGroup(r, prefix)
}

type Route interface {
	routeType()
}

type ApiRoute struct {
	Name         string
	Path         string
	Method       string
	ErrorTypes   []pyrin.ErrorType
	ResponseType any
	BodyType     any
}

func (r ApiRoute) routeType() {}

type FormApiRoute struct {
	Name         string
	Path         string
	Method       string
	ErrorTypes   []pyrin.ErrorType
	ResponseType any
	Spec         pyrin.FormSpec
}

func (r FormApiRoute) routeType() {}

type NormalRoute struct {
	Name   string
	Path   string
	Method string
}

func (r NormalRoute) routeType() {}

type RouteGroup struct {
	Router *Router
	Prefix string
}

func joinPaths(prefix, path string) string {
	return strings.TrimRight(prefix, "/") + "/" + strings.TrimLeft(path, "/")
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
			if h.Name == "" {
				continue
			}

			r.Router.AddRoute(ApiRoute{
				Name:         h.Name,
			Path:         joinPaths(r.Prefix, h.Path),
			Method:       h.Method,
			ErrorTypes:   h.Errors,
			ResponseType: h.ResponseType,
			BodyType:     h.BodyType,
			})
		case pyrin.FormApiHandler:
			if h.Name == "" {
				continue
			}

			r.Router.AddRoute(FormApiRoute{
				Name:         h.Name,
			Path:         joinPaths(r.Prefix, h.Path),
			Method:       h.Method,
			ErrorTypes:   h.Errors,
			ResponseType: h.ResponseType,
			Spec:         h.Spec,
			})
		case pyrin.NormalHandler:
			if h.Name == "" {
				continue
			}

			r.Router.AddRoute(NormalRoute{
				Name:   h.Name,
				Path:   joinPaths(r.Prefix, h.Path),
				Method: h.Method,
			})
		}
	}
}
