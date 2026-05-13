package apperror

import (
	"errors"
	"net/http"
)

const (
	CodeOK               = "OK_000"
	CodeInvalidBody      = "REQ_001"
	CodeValidationFailed = "REQ_002"
	CodeInvalidParam     = "REQ_003"
	CodeUnauthorized     = "AUTH_001"
	CodeForbidden        = "AUTH_003"
	CodeActorNotFound    = "AUTH_004"
	CodeRefreshFailed    = "AUTH_005"
	CodeRateLimited      = "RATE_001"
	CodeRequestTimeout   = "REQ_TIMEOUT"
	CodeInternal         = "SYS_001"
	CodeDatabase         = "DB_000"
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type AppError struct {
	Code       string
	Message    string
	HTTPStatus int
	Err        error
	Detail     interface{}
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}

	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}

	return e.Message
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}

func New(status int, code, message string) *AppError {
	return &AppError{
		HTTPStatus: status,
		Code:       code,
		Message:    message,
	}
}

func Wrap(err error, status int, code, message string) *AppError {
	appErr := New(status, code, message)
	appErr.Err = err
	if err != nil {
		appErr.Detail = err.Error()
	}
	return appErr
}

func WithDetail(err *AppError, detail interface{}) *AppError {
	if err == nil {
		return nil
	}

	err.Detail = detail
	return err
}

func Validation(details []FieldError) *AppError {
	return &AppError{
		HTTPStatus: http.StatusBadRequest,
		Code:       CodeValidationFailed,
		Message:    "validation failed",
		Detail:     details,
	}
}

func From(err error) *AppError {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	return Wrap(err, http.StatusInternalServerError, CodeInternal, "internal server error")
}
