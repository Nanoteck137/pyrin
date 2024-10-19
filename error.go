package pyrin

import "net/http"

// TODO(patrik): Capture the original error when returning api errors

const (
	ErrTypeUnknownError    ErrorType = "UNKNOWN_ERROR"
	ErrTypeRouteNotFound   ErrorType = "ROUTE_NOT_FOUND"
	ErrTypeValidationError ErrorType = "VALIDATION_ERROR"
)

type ErrorType string

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
