package errs

import (
	"fmt"
	"net/http"
)

type ApiError struct {
	Code     string
	HttpCode int
	Message  string
}

func (e *ApiError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *ApiError) Wrap(message string) *ApiError {
	return &ApiError{
		Code:     e.Code,
		HttpCode: e.HttpCode,
		Message:  message,
	}
}

func IsApiError(err error) (*ApiError, bool) {
	apiErr, ok := err.(*ApiError)
	return apiErr, ok
}

// Common errors
var (
	ErrInternalError = &ApiError{Code: "INTERNAL_SERVER_ERROR", HttpCode: http.StatusInternalServerError, Message: "Something was wrong"}
	ErrBadRequest    = &ApiError{Code: "BAD_REQUEST", HttpCode: http.StatusBadRequest, Message: "Invalid data"}
	ErrNotFound      = &ApiError{Code: "RECORD_NOT_FOUND", HttpCode: http.StatusNotFound, Message: "record not found"}
)
