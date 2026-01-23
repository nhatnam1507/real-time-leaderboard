// Package response provides HTTP response utilities and error handling.
package response

import (
	"fmt"
	"net/http"
)

// ErrorCode represents application error codes
type ErrorCode string

const (
	// CodeValidation represents a validation error.
	CodeValidation ErrorCode = "VALIDATION_ERROR"
	// CodeNotFound represents a not found error.
	CodeNotFound ErrorCode = "NOT_FOUND"
	// CodeUnauthorized represents an unauthorized error.
	CodeUnauthorized ErrorCode = "UNAUTHORIZED"
	// CodeForbidden represents a forbidden error.
	CodeForbidden ErrorCode = "FORBIDDEN"
	// CodeConflict represents a conflict error.
	CodeConflict ErrorCode = "CONFLICT"
	// CodeInternal represents an internal server error.
	CodeInternal ErrorCode = "INTERNAL_ERROR"
	// CodeBadRequest represents a bad request error.
	CodeBadRequest ErrorCode = "BAD_REQUEST"
	// CodeTooManyRequests represents a too many requests error.
	CodeTooManyRequests ErrorCode = "TOO_MANY_REQUESTS"
)

// APIError represents an API error (user-facing only)
type APIError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	HTTPStatus int       `json:"-"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return e.Message
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *APIError {
	return &APIError{
		Code:       CodeValidation,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
	}
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(resource string) *APIError {
	return &APIError{
		Code:       CodeNotFound,
		Message:    fmt.Sprintf("%s not found", resource),
		HTTPStatus: http.StatusNotFound,
	}
}

// NewUnauthorizedError creates a new unauthorized error
func NewUnauthorizedError(message string) *APIError {
	if message == "" {
		message = "Unauthorized"
	}
	return &APIError{
		Code:       CodeUnauthorized,
		Message:    message,
		HTTPStatus: http.StatusUnauthorized,
	}
}

// NewForbiddenError creates a new forbidden error
func NewForbiddenError(message string) *APIError {
	if message == "" {
		message = "Forbidden"
	}
	return &APIError{
		Code:       CodeForbidden,
		Message:    message,
		HTTPStatus: http.StatusForbidden,
	}
}

// NewConflictError creates a new conflict error
func NewConflictError(message string) *APIError {
	return &APIError{
		Code:       CodeConflict,
		Message:    message,
		HTTPStatus: http.StatusConflict,
	}
}

// NewInternalError creates a new internal server error
func NewInternalError(message string) *APIError {
	return &APIError{
		Code:       CodeInternal,
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
	}
}

// NewBadRequestError creates a new bad request error
func NewBadRequestError(message string) *APIError {
	return &APIError{
		Code:       CodeBadRequest,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
	}
}

// NewTooManyRequestsError creates a new too many requests error
func NewTooManyRequestsError(message string) *APIError {
	if message == "" {
		message = "Too many requests"
	}
	return &APIError{
		Code:       CodeTooManyRequests,
		Message:    message,
		HTTPStatus: http.StatusTooManyRequests,
	}
}

// IsAPIError checks if an error is an APIError
func IsAPIError(err error) bool {
	_, ok := err.(*APIError)
	return ok
}

// AsAPIError converts an error to APIError (for internal errors, returns generic message)
func AsAPIError(err error) *APIError {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr
	}
	return NewInternalError("Internal server error")
}
