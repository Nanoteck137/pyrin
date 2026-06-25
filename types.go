package pyrin

import "net/http"

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

type Router interface {
	Group(prefix string, middlewares ...MiddlewareFunc) Group
}

type Group interface {
	Register(handlers ...Handler)
}
