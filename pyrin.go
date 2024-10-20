package pyrin

import (
	"errors"
	"io"
	"io/fs"
	"net/http"
	"reflect"
	"sort"

	"github.com/MadAppGang/httplog/echolog"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nanoteck137/pyrin/spec"
	"github.com/nanoteck137/pyrin/tools/extract"
	"github.com/nanoteck137/pyrin/tools/resolve"
	"github.com/nanoteck137/pyrin/tools/validate"
	"github.com/nanoteck137/pyrin/utils"
)

type Body interface {
	validate.Validatable
}

type Context interface {
	Request() *http.Request
	Response() http.ResponseWriter
	Param(name string) string
}

type Handler interface {
	handlerType()
}

type ApiHandlerFunc func(c Context) (any, error)

type ApiHandler struct {
	Name        string
	Method      string
	Path        string
	DataType    any
	BodyType    Body
	RequireForm bool
	Errors      []ErrorType
	Middlewares []echo.MiddlewareFunc
	HandlerFunc ApiHandlerFunc
}

func (h ApiHandler) handlerType() {}

type NormalHandlerFunc func(c Context) error

type NormalHandler struct {
	Method      string
	Path        string
	Middlewares []echo.MiddlewareFunc
	HandlerFunc NormalHandlerFunc
}

func (h NormalHandler) handlerType() {}

var _ Context = (*wrapperContext)(nil)

type wrapperContext struct {
	c echo.Context
}

func (w *wrapperContext) Response() http.ResponseWriter {
	return w.c.Response()
}

func (w *wrapperContext) Request() *http.Request {
	return w.c.Request()
}

func (w *wrapperContext) Param(name string) string {
	return w.c.Param(name)
}

type Router interface {
	Group(prefix string) Group
}

type ServerRouter struct {
	e *echo.Echo
}

func (r *ServerRouter) Group(prefix string) Group {
	return newServerGroup(r.e, prefix)
}

type Group interface {
	Register(handlers ...Handler)
}

type ServerGroup struct {
	prefix string
	group  *echo.Group
}

func (g *ServerGroup) Register(handlers ...Handler) {
	for _, h := range handlers {
		switch h := h.(type) {
		case ApiHandler:
			wrapHandler := func(c echo.Context) error {
				context := &wrapperContext{
					c: c,
				}

				data, err := h.HandlerFunc(context)
				if err != nil {
					return err
				}

				return c.JSON(200, SuccessResponse(data))
			}

			g.group.Add(h.Method, h.Path, wrapHandler, h.Middlewares...)
		case NormalHandler:
			wrapHandler := func(c echo.Context) error {
				context := &wrapperContext{
					c: c,
				}
				return h.HandlerFunc(context)
			}
			g.group.Add(h.Method, h.Path, wrapHandler, h.Middlewares...)
		}
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

func errorHandler(err error, c echo.Context) {
	switch err := err.(type) {
	case *Error:
		c.JSON(err.Code, ErrorResponse(*err))
	case *echo.HTTPError:
		c.JSON(err.Code, ErrorResponse(Error{
			Code:    err.Code,
			Type:    ErrTypeUnknownError,
			Message: err.Error(),
		}))
	default:
		c.JSON(500, ErrorResponse(Error{
			Code: 500,
			Type: ErrTypeUnknownError,
			// Message: "Internal Server Error",
			Message: err.Error(),
		}))
	}
}

type ServerConfig struct {
	RegisterHandlers func(router Router)
	LogName          string
}

func NewServer(config *ServerConfig) *Server {
	e := echo.New()
	e.HTTPErrorHandler = errorHandler

	e.RouteNotFound("/*", func(c echo.Context) error {
		return RouteNotFound()
	})

	if config.LogName == "" {
		config.LogName = "Pyrin Server"
	}

	e.Use(echolog.LoggerWithName(config.LogName))
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	router := ServerRouter{
		e: e,
	}

	if config.RegisterHandlers != nil {
		config.RegisterHandlers(&router)
	}

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
	ErrorTypes  []ErrorType
	Data        any
	Body        any
	RequireForm bool
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
		switch h := h.(type) {
		case ApiHandler:
			r.Routes = append(r.Routes, Route{
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

func ServeFile(w http.ResponseWriter, r *http.Request, filesystem fs.FS, file string) error {
	f, err := filesystem.Open(file)
	if err != nil {
		// TODO(patrik): Add NoContentError to pyrin
		return echo.ErrNotFound
	}
	defer f.Close()

	fi, _ := f.Stat()

	ff, ok := f.(io.ReadSeeker)
	if !ok {
		return errors.New("file does not implement io.ReadSeeker")
	}

	http.ServeContent(w, r, fi.Name(), fi.ModTime(), ff)

	return nil
}

func GenerateSpec(routes []Route) (*spec.Server, error) {
	s := &spec.Server{}

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

		s.Endpoints = append(s.Endpoints, spec.Endpoint{
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
			fields := make([]spec.TypeField, 0, len(t.Fields))

			for _, f := range t.Fields {
				s, err := utils.TypeToString(f.Type)
				if err != nil {
					return nil, err
				}

				fields = append(fields, spec.TypeField{
					Name: f.Name,
					Type: s,
					Omit: f.Optional,
				})
			}

			s.Types = append(s.Types, spec.Type{
				Name:   st.Name,
				Extend: "",
				Fields: fields,
			})
		case *resolve.TypeSameStruct:
			s.Types = append(s.Types, spec.Type{
				Name:   st.Name,
				Extend: t.Type.Name,
			})
		}
	}


	return s, nil
}
