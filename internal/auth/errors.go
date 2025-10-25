package auth

import "errors"

type ErrorCode string

const (
	ErrorInvalidCredentials ErrorCode = "invalid_credentials"
	ErrorInvalidScope       ErrorCode = "invalid_scope"
	ErrorInvalidToken       ErrorCode = "invalid_token"
	ErrorInternal           ErrorCode = "internal_error"
)

type Error struct {
	Code    ErrorCode
	Message string
	err     error
}

func (e *Error) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.err != nil {
		return e.err.Error()
	}
	return string(e.Code)
}

func (e *Error) Unwrap() error {
	return e.err
}

func NewError(code ErrorCode, message string, err error) *Error {
	return &Error{Code: code, Message: message, err: err}
}

func AsError(err error) (*Error, bool) {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
