// Package errs provides error functionality with HTTP status codes.
package errs

import (
	"net/http"
)

// Common errors
var (
	NotImplemented = New("not implemented", http.StatusNotImplemented)
)

// Error is an error interface with a HTTP status code.
// This can be useful for providing a hint to the caller
// of an appropriate HTTP error code to be sent to any
// HTTP client that might have initiated the request.
type Error interface {
	error

	// Status returns a HTTP status code associated
	// with this error.
	StatusCode() int
}

type errorImpl struct {
	message    string
	statusCode int
	code       string
}

// Error implements the error interface.
func (e errorImpl) Error() string {
	return e.message
}

// StatusCode returns the suggested HTTP status code to return
// to the HTTP client.
func (e errorImpl) StatusCode() int {
	return e.statusCode
}

// BadRequest returns an errs.Error object with a HTTP status code
// of Bad Request (400)
func BadRequest(message string) error {
	return New(message, http.StatusBadRequest)
}

// Forbidden returns an errs.Error object with a HTTP status code
// of Forbidden (403)
func Forbidden(message string) error {
	return New(message, http.StatusForbidden)
}

// ServerError returns an errs.Error object with a HTTP status code
// of Internal Server Errror (500)
func ServerError(message string) error {
	return New(message, http.StatusInternalServerError)
}

// New returns an errs.Error object with the specified message and
// HTTP status code.
func New(message string, statusCode int) error {
	return errorImpl{
		message:    message,
		statusCode: statusCode,
	}
}

// Code returns the code associated with the error, or a blank string
// if the error does not have a code. Useful for AWS and other packages
// that have error types with a Code() method.
func Code(err error) string {
	type ErrorWithCode interface {
		Code() string
	}
	if errWithCode, ok := err.(ErrorWithCode); ok {
		return errWithCode.Code()
	}
	return ""
}

// HasErrorCode determines whether an error has the specified error code.
// Useful for AWS and other packages that have error types with a Code() method.
func HasErrorCode(err error, code string) bool {
	return Code(err) == code
}
