package pyrin

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/MadAppGang/httplog/echolog"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nanoteck137/validate"
)

const formBodyKey = "body"

const jsonMimeType = "application/json"
const multipartFormMimeType = "multipart/form-data"

// TODO(patrik): Move to server config?
const defaultMemory = 32 << 20

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
	ReturnType  any
	BodyType    any
	Errors      []ErrorType
	Middlewares []echo.MiddlewareFunc
	HandlerFunc ApiHandlerFunc
}

func (h ApiHandler) handlerType() {}

type FormFileSpec struct {
	NumExpected int
}

type FormSpec struct {
	BodyType any
	Files    map[string]FormFileSpec
}

type FormApiHandler struct {
	Name        string
	Method      string
	Path        string
	ReturnType  any
	Spec        FormSpec
	Errors      []ErrorType
	Middlewares []echo.MiddlewareFunc
	HandlerFunc ApiHandlerFunc
}

func (h FormApiHandler) handlerType() {}

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

	formSpec *FormSpec
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

func (w *wrapperContext) checkContentType(expected string) error {
	contentType := w.c.Request().Header.Get("Content-Type")
	if contentType == "" {
		return BadContentType(expected)

	}

	typ, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return BadContentType(expected)
	}

	if typ != expected {
		return BadContentType(expected)
	}

	return nil
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

func validateForm(spec *FormSpec, form *multipart.Form) error {
	extra := make(map[string]string)

	if spec.BodyType != nil {
		data, exists := form.Value[formBodyKey]
		if !exists && len(data) < 1 {
			extra[formBodyKey] = "contains no data"
		}
	}

	for field, spec := range spec.Files {
		files := form.File[field]
		// TODO(patrik): Should this be more strict when checking num expected
		if len(files) < spec.NumExpected {
			extra[field] = fmt.Sprintf("expected %d or more files, got %d", spec.NumExpected, len(files))
			continue
		}
	}

	if len(extra) > 0 {
		return FormValidationError(extra)
	}

	return nil
}

func (g *ServerGroup) Register(handlers ...Handler) {
	for _, h := range handlers {
		switch h := h.(type) {
		case ApiHandler:
			wrapHandler := func(c echo.Context) error {
				context := &wrapperContext{
					c: c,
				}

				if h.BodyType != nil {
					err := context.checkContentType(jsonMimeType)
					if err != nil {
						return err
					}
				}

				data, err := h.HandlerFunc(context)
				if err != nil {
					return err
				}

				return c.JSON(200, SuccessResponse(data))
			}

			g.group.Add(h.Method, h.Path, wrapHandler, h.Middlewares...)
		case FormApiHandler:
			wrapHandler := func(c echo.Context) error {
				context := &wrapperContext{
					c:        c,
					formSpec: &h.Spec,
				}

				err := context.checkContentType(multipartFormMimeType)
				if err != nil {
					return err
				}

				err = c.Request().ParseMultipartForm(defaultMemory)
				if err != nil {
					return err
				}

				err = validateForm(context.formSpec, c.Request().MultipartForm)
				if err != nil {
					return err
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

type Files []*multipart.FileHeader

func FormFiles(c Context, key string) (Files, error) {
	wrapperContext := c.(*wrapperContext)

	spec := wrapperContext.formSpec

	if spec == nil {
		// NOTE(patrik): These are developer faults
		return nil, errors.New("handler cannot use forms use 'FormApiHandler'")
	}

	_, exists := spec.Files[key]
	if !exists {
		// NOTE(patrik): These are developer faults
		return nil, fmt.Errorf("%s: is not valid, key is not defined in spec", key)
	}

	form := c.Request().MultipartForm
	files := form.File[key]

	return files, nil
}

func Body[T any](c Context) (T, error) {
	var res T

	wrapperContext := c.(*wrapperContext)

	var body io.Reader
	if wrapperContext.formSpec == nil {
		body = c.Request().Body
	} else {
		data := c.Request().FormValue(formBodyKey)
		body = strings.NewReader(data)

		// TODO(patrik): Add check for body type to match the spec
	}

	decoder := json.NewDecoder(body)

	if !decoder.More() {
		return res, EmptyBody()
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
