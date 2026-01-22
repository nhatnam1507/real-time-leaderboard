// Package v1 provides REST API v1 handlers for the auth module.
package v1

import (
	"errors"

	"real-time-leaderboard/internal/module/auth/domain"
	"real-time-leaderboard/internal/shared/response"
	"real-time-leaderboard/internal/shared/validator"
)

// toAPIError converts auth domain errors to APIError (internal helper)
func toAPIError(err error) *response.APIError {
	if err == nil {
		return nil
	}

	// Check for validation errors
	if validator.IsValidationError(err) {
		var validationErr *validator.ValidationError
		if errors.As(err, &validationErr) {
			return response.NewValidationError(validationErr.Message)
		}
	}

	// Check for domain errors
	if errors.Is(err, domain.ErrUserNotFound) {
		return response.NewNotFoundError("User")
	}
	if errors.Is(err, domain.ErrUserAlreadyExists) {
		return response.NewConflictError("User already exists")
	}
	if errors.Is(err, domain.ErrInvalidCredentials) {
		return response.NewUnauthorizedError("Invalid credentials")
	}
	if errors.Is(err, domain.ErrInvalidToken) {
		return response.NewUnauthorizedError("Invalid or expired token")
	}

	// If it's already an APIError, return it as-is
	if apiErr, ok := err.(*response.APIError); ok {
		return apiErr
	}

	// Default to internal error for wrapped errors
	return response.NewInternalError("An error occurred")
}
