package pyrin

import (
	"github.com/labstack/echo/v4"
	"github.com/nanoteck137/pyrin/api"
)

type Context interface {
	Param(name string) string
}

type HandlerFunc func(c Context) (any, error)

type Handler struct {
	Name        string
	Method      string
	Path        string
	DataType    any
	BodyType    any
	IsMultiForm bool
	Errors      []api.ErrorType
	HandlerFunc HandlerFunc
	Middlewares []echo.MiddlewareFunc
}

type Group interface {
	Register(handlers ...Handler)
}

var _ Context = (*wrapperContext)(nil)

type wrapperContext struct {
	c echo.Context
}

func (w *wrapperContext) Param(name string) string {
	return w.c.Param(name)
}

type ServerGroup struct {
	prefix string
	group  *echo.Group
}

func (g *ServerGroup) Register(handlers ...Handler) {
	for _, h := range handlers {
		f := h.HandlerFunc

		// log.Debug("Registering", "method", h.Method, "name", h.Name, "path", g.Prefix+h.Path)
		wrapHandler := func(c echo.Context) error {
			context := &wrapperContext{
				c: c,
			}

			data, err := f(context)
			if err != nil {
				return err
			}

			return c.JSON(200, api.SuccessResponse(data))
		}

		g.group.Add(h.Method, h.Path, wrapHandler, h.Middlewares...)
	}
}

func newServerGroup(e *echo.Echo, prefix string, m ...echo.MiddlewareFunc) *ServerGroup {
	g := e.Group(prefix, m...)
	return &ServerGroup{
		prefix: prefix,
		group:  g,
	}
}

type Server struct {
	e *echo.Echo
}

const ErrTypeUnknownError api.ErrorType = "UNKNOWN_ERROR"

func errorHandler(err error, c echo.Context) {
	switch err := err.(type) {
	case *api.Error:
		c.JSON(err.Code, api.ErrorResponse(*err))
	case *echo.HTTPError:
		c.JSON(err.Code, api.ErrorResponse(api.Error{
			Code:    err.Code,
			Type:    ErrTypeUnknownError,
			Message: err.Error(),
		}))
	default:
		c.JSON(500, api.ErrorResponse(api.Error{
			Code:    500,
			Type:    ErrTypeUnknownError,
			Message: "Internal Server Error",
		}))
	}
}

func NewServer() *Server {
	e := echo.New()
	e.HTTPErrorHandler = errorHandler

	return &Server{
		e: e,
	}
}

func (s *Server) Start(addr string) error {
	return s.e.Start(addr)
}

func (s *Server) Group(prefix string, m ...echo.MiddlewareFunc) *ServerGroup {
	return newServerGroup(s.e, prefix, m...)
}

type Route struct {
	Name        string
	Path        string
	Method      string
	ErrorTypes  []api.ErrorType
	Data        any
	Body        any
	IsMultiForm bool
}

type RouteGroup struct {
	Prefix string
	Routes []Route
}

func NewRouteGroup(prefix string) *RouteGroup {
	return &RouteGroup{
		Prefix: prefix,
		Routes: []Route{},
	}
}

func (r *RouteGroup) Register(handlers ...Handler) {
	for _, h := range handlers {
		r.Routes = append(r.Routes, Route{
			Name:        h.Name,
			Path:        r.Prefix + h.Path,
			Method:      h.Method,
			ErrorTypes:  h.Errors,
			Data:        h.DataType,
			Body:        h.BodyType,
			IsMultiForm: h.IsMultiForm,
		})
	}
}
