package pyrin

import "net/http"

// TODO(patrik): Capture the original error when returning api errors

const (
	ErrTypeUnknownError        ErrorType = "UNKNOWN_ERROR"
	ErrTypeRouteNotFound       ErrorType = "ROUTE_NOT_FOUND"
	ErrTypeValidationError     ErrorType = "VALIDATION_ERROR"
	ErrTypeFormValidationError ErrorType = "FORM_VALIDATION_ERROR"
	ErrTypeEmptyBody           ErrorType = "EMPTY_BODY_ERROR"
	ErrTypeBadContentType      ErrorType = "BAD_CONTENT_TYPE_ERROR"
)

var GlobalErrors = []ErrorType{
	ErrTypeUnknownError,
	ErrTypeRouteNotFound,
	ErrTypeValidationError,
	ErrTypeFormValidationError,
	ErrTypeEmptyBody,
	ErrTypeBadContentType,
}

type ErrorType string

func (e ErrorType) String() string {
	return string(e)
}

type Error struct {
	Code    int       `json:"code"`
	Type    ErrorType `json:"type"`
	Message string    `json:"message"`
	Extra   any       `json:"extra,omitempty"`
}

func (e *Error) Error() string {
	return e.Message
}

type Response struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

func SuccessResponse(data any) Response {
	return Response{
		Success: true,
		Data:    data,
	}
}

func ErrorResponse(err Error) Response {
	return Response{
		Success: false,
		Error:   &err,
	}
}

func RouteNotFound() *Error {
	return &Error{
		Code:    http.StatusNotFound,
		Type:    ErrTypeRouteNotFound,
		Message: "Route not found",
	}
}

func ValidationError(extra any) *Error {
	return &Error{
		Code:    http.StatusBadRequest,
		Type:    ErrTypeValidationError,
		Message: "Validation error",
		Extra:   extra,
	}
}

func FormValidationError(extra any) *Error {
	return &Error{
		Code:    http.StatusBadRequest,
		Type:    ErrTypeFormValidationError,
		Message: "Form Validation error",
		Extra:   extra,
	}
}

func EmptyBody() *Error {
	return &Error{
		Code:    http.StatusBadRequest,
		Type:    ErrTypeEmptyBody,
		Message: "Empty body",
	}
}

func BadContentType(expected string) *Error {
	return &Error{
		Code:    http.StatusBadRequest,
		Type:    ErrTypeBadContentType,
		Message: "Bad Content-Type expected: " + expected,
	}
}
