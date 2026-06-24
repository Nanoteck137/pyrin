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

	"github.com/go-chi/chi/v5"
	"github.com/nanoteck137/validate"
)

const formBodyKey = "body"

const jsonMimeType = "application/json"
const multipartFormMimeType = "multipart/form-data"

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

type MiddlewareFunc func(http.Handler) http.Handler

type ApiHandlerFunc func(c Context) (any, error)

type ApiHandler struct {
	Name         string
	Method       string
	Path         string
	ResponseType any
	BodyType     any
	Errors       []ErrorType
	Middlewares  []MiddlewareFunc
	HandlerFunc  ApiHandlerFunc
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
	Name         string
	Method       string
	Path         string
	ResponseType any
	Spec         FormSpec
	Errors       []ErrorType
	Middlewares  []MiddlewareFunc
	HandlerFunc  ApiHandlerFunc
}

func (h FormApiHandler) handlerType() {}

type NormalHandlerFunc func(c Context) error

type NormalHandler struct {
	Name        string
	Method      string
	Path        string
	Middlewares []MiddlewareFunc
	HandlerFunc NormalHandlerFunc
}

func (h NormalHandler) handlerType() {}

var _ Context = (*wrapperContext)(nil)

type wrapperContext struct {
	w http.ResponseWriter
	r *http.Request

	formSpec *FormSpec
}

func (w *wrapperContext) Response() http.ResponseWriter {
	return w.w
}

func (w *wrapperContext) Request() *http.Request {
	return w.r
}

func (w *wrapperContext) Param(name string) string {
	return chi.URLParam(w.r, name)
}

func (w *wrapperContext) checkContentType(expected string) error {
	contentType := w.r.Header.Get("Content-Type")
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
	Group(prefix string, middlewares ...MiddlewareFunc) Group
}

type Group interface {
	Register(handlers ...Handler)
}

type serverGroup struct {
	router       chi.Router
	errorHandler func(err error, w http.ResponseWriter, r *http.Request)
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
		if len(files) < spec.NumExpected {
			extra[field] = fmt.Sprintf(
				"expected %d or more files, got %d",
				spec.NumExpected,
				len(files),
			)
			continue
		}
	}

	if len(extra) > 0 {
		return FormValidationError(extra)
	}

	return nil
}

func convertPath(path string) string {
	var b strings.Builder
	b.Grow(len(path) + 8)

	i := 0
	for i < len(path) {
		if path[i] == ':' {
			b.WriteByte('{')
			i++
			for i < len(path) && path[i] != '/' {
				b.WriteByte(path[i])
				i++
			}
			b.WriteByte('}')
		} else {
			b.WriteByte(path[i])
			i++
		}
	}

	return b.String()
}

func writeJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func (g *serverGroup) Register(handlers ...Handler) {
	for _, h := range handlers {
		switch h := h.(type) {
		case ApiHandler:
			handlerFn := func(w http.ResponseWriter, r *http.Request) {
				ctx := &wrapperContext{
					w: w,
					r: r,
				}

				if h.BodyType != nil {
					err := ctx.checkContentType(jsonMimeType)
					if err != nil {
						g.errorHandler(err, w, r)
						return
					}
				}

				data, err := h.HandlerFunc(ctx)
				if err != nil {
					g.errorHandler(err, w, r)
					return
				}

				writeJSON(w, http.StatusOK, SuccessResponse(data))
			}

			var handler http.Handler = http.HandlerFunc(handlerFn)
			for i := len(h.Middlewares) - 1; i >= 0; i-- {
				handler = h.Middlewares[i](handler)
			}

			g.router.Method(h.Method, convertPath(h.Path), handler)

		case FormApiHandler:
			handlerFn := func(w http.ResponseWriter, r *http.Request) {
				ctx := &wrapperContext{
					w:        w,
					r:        r,
					formSpec: &h.Spec,
				}

				err := ctx.checkContentType(multipartFormMimeType)
				if err != nil {
					g.errorHandler(err, w, r)
					return
				}

				err = r.ParseMultipartForm(defaultMemory)
				if err != nil {
					g.errorHandler(err, w, r)
					return
				}

				err = validateForm(ctx.formSpec, r.MultipartForm)
				if err != nil {
					g.errorHandler(err, w, r)
					return
				}

				data, err := h.HandlerFunc(ctx)
				if err != nil {
					g.errorHandler(err, w, r)
					return
				}

				writeJSON(w, http.StatusOK, SuccessResponse(data))
			}

			var handler http.Handler = http.HandlerFunc(handlerFn)
			for i := len(h.Middlewares) - 1; i >= 0; i-- {
				handler = h.Middlewares[i](handler)
			}

			g.router.Method(h.Method, convertPath(h.Path), handler)

		case NormalHandler:
			handlerFn := func(w http.ResponseWriter, r *http.Request) {
				ctx := &wrapperContext{
					w: w,
					r: r,
				}

				err := h.HandlerFunc(ctx)
				if err != nil {
					g.errorHandler(err, w, r)
					return
				}
			}

			var handler http.Handler = http.HandlerFunc(handlerFn)
			for i := len(h.Middlewares) - 1; i >= 0; i-- {
				handler = h.Middlewares[i](handler)
			}

			g.router.Method(h.Method, convertPath(h.Path), handler)
		}
	}
}

type Server struct {
	mux          *chi.Mux
	errorHandler func(err error, w http.ResponseWriter, r *http.Request)
}

type ErrorCallback func(err error)

func errorHandler(err error, w http.ResponseWriter, r *http.Request, errorCallback ErrorCallback) {
	switch e := err.(type) {
	case *Error:
		if errorCallback != nil {
			errorCallback(e)
		}

		writeJSON(w, e.Code, ErrorResponse(*e))
	case *NoContentError:
		w.WriteHeader(e.Code)
	default:
		if errorCallback != nil {
			errorCallback(e)
		}

		writeJSON(w, http.StatusInternalServerError, ErrorResponse(Error{
			Code:    http.StatusInternalServerError,
			Type:    ErrTypeUnknownError,
			Message: "Internal Server Error",
		}))
	}
}

type ServerConfig struct {
	RegisterHandlers func(router Router)
	ErrorCallback    ErrorCallback
	Middlewares      []MiddlewareFunc
}

func NewServer(config *ServerConfig) *Server {
	mux := chi.NewMux()

	errHandler := func(err error, w http.ResponseWriter, r *http.Request) {
		errorHandler(err, w, r, config.ErrorCallback)
	}

	mux.NotFound(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errHandler(RouteNotFound(), w, r)
		// writeJSON(w, http.StatusNotFound, ErrorResponse(*RouteNotFound()))
	}))

	for _, m := range config.Middlewares {
		mux.Use(m)
	}

	return &Server{
		mux:          mux,
		errorHandler: errHandler,
	}
}

func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, s.mux)
}

func (s *Server) Group(
	prefix string,
	middlewares ...MiddlewareFunc,
) Group {
	sub := chi.NewRouter()

	for _, middleware := range middlewares {
		sub.Use(middleware)
	}

	s.mux.Mount(prefix, sub)

	return &serverGroup{
		router:       sub,
		errorHandler: s.errorHandler,
	}
}

func FormFiles(c Context, key string) ([]*multipart.FileHeader, error) {
	wrapperContext := c.(*wrapperContext)

	spec := wrapperContext.formSpec

	if spec == nil {
		// TODO(patrik): Internal error or panic
		return nil, errors.New("handler cannot use forms use 'FormApiHandler'")
	}

	_, exists := spec.Files[key]
	if !exists {
		// TODO(patrik): Internal error or panic
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

func ServeFile(c Context, filesystem fs.FS, file string) error {
	f, err := filesystem.Open(file)
	if err != nil {
		return NoContentNotFound()
	}
	defer f.Close()

	fi, _ := f.Stat()

	ff, ok := f.(io.ReadSeeker)
	if !ok {
		return errors.New("file does not implement io.ReadSeeker")
	}

	http.ServeContent(c.Response(), c.Request(), fi.Name(), fi.ModTime(), ff)

	return nil
}

type hookedResponseWriter struct {
	http.ResponseWriter
	got404 bool
}

func (hrw *hookedResponseWriter) WriteHeader(status int) {
	if status == http.StatusNotFound {
		hrw.got404 = true
	} else {
		hrw.ResponseWriter.WriteHeader(status)
	}
}

func (hrw *hookedResponseWriter) Write(p []byte) (int, error) {
	if hrw.got404 {
		return len(p), nil
	}

	return hrw.ResponseWriter.Write(p)
}

func SpaHandler(root fs.FS, indexFilename string) Handler {
	return NormalHandler{
		Method: http.MethodGet,
		Path:   "/*",
		HandlerFunc: func(c Context) error {
			fs := http.FileServer(http.FS(root))

			hookedWriter := &hookedResponseWriter{ResponseWriter: c.Response()}
			fs.ServeHTTP(hookedWriter, c.Request())

			if hookedWriter.got404 {
				accept := c.Request().Header.Get("Accept")
				if !strings.Contains(accept, "text/html") {
					c.Response().WriteHeader(http.StatusNotFound)
					fmt.Fprint(c.Response(), "404 not found")
				} else {
					c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
					ServeFile(c, root, indexFilename)
				}
			}

			return nil
		},
	}
}
