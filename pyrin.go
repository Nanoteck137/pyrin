package pyrin

import (
	"encoding/json"
	"net/http"

	"github.com/go-viper/mapstructure/v2"
	"github.com/kr/pretty"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nanoteck137/pyrin/api"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Body interface {
	validation.Validatable
}

type Context interface {
	Request() *http.Request
	Param(name string) string
	Body() any
}

type HandlerFunc func(c Context) (any, error)

type Handler struct {
	Name        string
	Method      string
	Path        string
	DataType    any
	BodyType    Body
	RequireForm bool
	Errors      []api.ErrorType
	Middlewares []echo.MiddlewareFunc
	HandlerFunc HandlerFunc
}

type Group interface {
	Register(handlers ...Handler)
}

var _ Context = (*wrapperContext)(nil)

type wrapperContext struct {
	c    echo.Context
	body any
}

func (w *wrapperContext) Body() any {
	return w.body
}

func (w *wrapperContext) Request() *http.Request {
	return w.c.Request()
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
		// log.Debug("Registering", "method", h.Method, "name", h.Name, "path", g.Prefix+h.Path)
		wrapHandler := func(c echo.Context) error {
			context := &wrapperContext{
				c: c,
			}

			if h.BodyType != nil && !h.RequireForm {
				b := h.BodyType

				var data map[string]any

				decoder := json.NewDecoder(context.Request().Body)
				err := decoder.Decode(&data)
				if err != nil {
					return err
				}

				d, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
					Result:  &b,
					TagName: "json",
				})
				if err != nil {
					return err
				}

				err = d.Decode(data)
				if err != nil {
					return err
				}

				pretty.Println(b)

				err = b.Validate()
				if err != nil {
					extra := make(map[string]string)

					if e, ok := err.(validation.Errors); ok {
						for k, v := range e {
							extra[k] = v.Error()
						}
					}

					return &api.Error{
						Code:    400,
						Type:    "VALIDATION_ERROR",
						Message: "Body Validation error",
						Extra:   extra,
					}
				}

				context.body = b
			}

			data, err := h.HandlerFunc(context)
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
			Code: 500,
			Type: ErrTypeUnknownError,
			// Message: "Internal Server Error",
			Message: err.Error(),
		}))
	}
}

func NewServer() *Server {
	e := echo.New()
	e.HTTPErrorHandler = errorHandler

	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

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
