// Package errors provides application-specific error types and utilities.
package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError represents an application error with HTTP status code and optional details.
type AppError struct {
	Code    int               // HTTP status code
	Message string            // User-friendly error message
	Err     error             // Original wrapped error
	Details map[string]string // Optional details (e.g., validation errors)
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the wrapped error for errors.Is and errors.As support.
func (e *AppError) Unwrap() error {
	return e.Err
}

// Is reports whether any error in err's tree matches target.
// This allows AppError to work with errors.Is().
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// As finds the first error in err's tree that matches target.
// This is a helper function that wraps errors.As().
func As(err error, target any) bool {
	return errors.As(err, target)
}

// Is reports whether any error in err's tree matches target.
// This is a helper function that wraps errors.Is().
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// NewBadRequest creates a new AppError with HTTP 400 Bad Request status.
func NewBadRequest(message string) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: message,
	}
}

// NewUnauthorized creates a new AppError with HTTP 401 Unauthorized status.
func NewUnauthorized(message string) *AppError {
	return &AppError{
		Code:    http.StatusUnauthorized,
		Message: message,
	}
}

// NewForbidden creates a new AppError with HTTP 403 Forbidden status.
func NewForbidden(message string) *AppError {
	return &AppError{
		Code:    http.StatusForbidden,
		Message: message,
	}
}

// NewNotFound creates a new AppError with HTTP 404 Not Found status.
func NewNotFound(message string) *AppError {
	return &AppError{
		Code:    http.StatusNotFound,
		Message: message,
	}
}

// NewInternalError creates a new AppError with HTTP 500 Internal Server Error status.
// The original error is wrapped for debugging purposes.
func NewInternalError(err error) *AppError {
	return &AppError{
		Code:    http.StatusInternalServerError,
		Message: "internal server error",
		Err:     err,
	}
}

// NewValidationError creates a new AppError with HTTP 400 Bad Request status
// and includes validation error details.
func NewValidationError(details map[string]string) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: "validation failed",
		Details: details,
	}
}

// WithDetails adds details to an existing AppError and returns it.
func (e *AppError) WithDetails(details map[string]string) *AppError {
	e.Details = details
	return e
}

// WithError wraps an original error and returns the AppError.
func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	return e
}

// HTTPStatus returns the HTTP status code for the error.
// If err is not an AppError, it returns 500 Internal Server Error.
func HTTPStatus(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return http.StatusInternalServerError
}

// GetDetails returns the error details if err is an AppError with details.
// Returns nil if err is not an AppError or has no details.
func GetDetails(err error) map[string]string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Details
	}
	return nil
}

// Wrap wraps an error with additional context message.
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}
