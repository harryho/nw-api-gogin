package catalog

import (
	"errors"
	"net/http"
)

type ErrorCode string

const (
	ErrorInternal   ErrorCode = "internal_error"
	ErrorValidation ErrorCode = "validation_error"
	ErrorNotFound   ErrorCode = "not_found"
	ErrorConflict   ErrorCode = "conflict"
)

type Error struct {
	Code    ErrorCode
	Message string
	Err     error
	Status  int
}

func (e *Error) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return ""
}

func (e *Error) Unwrap() error {
	return e.Err
}

func NewValidationError(message string, err error) *Error {
	return &Error{Code: ErrorValidation, Message: message, Err: err, Status: http.StatusUnprocessableEntity}
}

func NewNotFoundError(message string, err error) *Error {
	return &Error{Code: ErrorNotFound, Message: message, Err: err, Status: http.StatusNotFound}
}

func NewConflictError(message string, err error) *Error {
	return &Error{Code: ErrorConflict, Message: message, Err: err, Status: http.StatusConflict}
}

func NewInternalError(message string, err error) *Error {
	return &Error{Code: ErrorInternal, Message: message, Err: err, Status: http.StatusInternalServerError}
}

func Wrap(err error, code ErrorCode, message string, status int) *Error {
	return &Error{Code: code, Message: message, Err: err, Status: status}
}

func AsError(err error) (*Error, bool) {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
