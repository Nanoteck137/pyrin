package api

type ErrorType string

type ApiError struct {
	Code    int       `json:"code"`
	Type    ErrorType `json:"type"`
	Message string    `json:"message"`
	Extra   any       `json:"extra,omitempty"`
}

func (e *ApiError) Error() string {
	return e.Message
}

type ApiResponse struct {
	Success bool      `json:"status"`
	Data    any       `json:"data,omitempty"`
	Error   *ApiError `json:"error,omitempty"`
}

func SuccessResponse(data any) ApiResponse {
	return ApiResponse{
		Success: true,
		Data:    data,
	}
}
