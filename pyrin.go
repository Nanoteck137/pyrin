package pyrin

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net/http"

	"github.com/MadAppGang/httplog/echolog"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nanoteck137/validate"
)

type Context interface {
	Request() *http.Request
	Response() http.ResponseWriter
	Param(name string) string
}

type Transformable interface {
	Transform()
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
	BodyType    any
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
			// TODO(patrik): Check for Content-Type
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

func Body[T any](c Context) (T, error) {
	var res T

	decoder := json.NewDecoder(c.Request().Body)

	if !decoder.More() {
		// TODO(patrik): Better error
		return res, errors.New("Empty body")
	}

	err := decoder.Decode(&res)
	if err != nil {
		return res, err
	}

	var p any = &res
	if t, ok := p.(Transformable); ok {
		t.Transform()
	}

	if v, ok := p.(validate.Validatable); ok {
		err = v.Validate()
		if err != nil {
			extra := make(map[string]string)

			if e, ok := err.(validate.Errors); ok {
				for k, v := range e {
					extra[k] = v.Error()
				}
			}

			return res, ValidationError(extra)
		}
	}

	return res, nil
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
