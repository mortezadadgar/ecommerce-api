package domain

import (
	"errors"
	"fmt"
)

// error codes analogous to http error codes.
const (
	EINTERNAL     = 500
	ETOOLARGE     = 413
	ECONFLICT     = 409
	EINVALID      = 400
	ENOTFOUND     = 404
	EUNAUTHORIZED = 401
)

// Error represents error messages shown to users.
type Error struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// Error implements the error interface. Not used by the application.
func (e *Error) Error() string {
	return fmt.Sprintf("error: code:%d message:%s", e.Code, e.Message)
}

// ErrorCode unwraps an application error and returns its code.
func ErrorCode(err error) int {
	var e *Error
	if err == nil {
		return 0
	} else if errors.As(err, &e) {
		return e.Code
	}

	return EINTERNAL
}

// ErrorMessage unwraps an application error and returns its message.
func ErrorMessage(err error) string {
	var e *Error
	if err == nil {
		return ""
	} else if errors.As(err, &e) {
		return e.Message
	}

	return "internal server error"
}

// Errorf returns a error with giving code and formatted message.
func Errorf(code int, format string, args ...interface{}) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}
