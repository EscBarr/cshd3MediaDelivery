package response

import "cshdMediaDelivery/internal/lib/errs"

const (
	StatusOK    = "OK"
	StatusError = "Error"
)

type ApiResponse struct {
	Status     string      `json:"status"`
	StatusCode int         `json:"status_code"`
	Data       interface{} `json:"data,omitempty"`
	ErrorCode  string      `json:"error,omitempty"`
	Message    string      `json:"message,omitempty"`
	Metadata   interface{} `json:"metadata,omitempty"`
}

func ErrorApiResponse(err *errs.ApiError) ApiResponse {
	return ApiResponse{
		Status:     StatusError,
		StatusCode: err.HttpCode,
		ErrorCode:  err.Code,
		Message:    err.Message,
	}
}

func SuccessResponse(msg string) ApiResponse {
	return ApiResponse{
		Status:     StatusOK,
		StatusCode: 200,
		Message:    msg,
	}
}
