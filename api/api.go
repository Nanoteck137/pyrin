package api

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
	Success bool   `json:"status"`
	Data    any    `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

func SuccessResponse(data any) Response {
	return Response{
		Success: true,
		Data:    data,
	}
}
