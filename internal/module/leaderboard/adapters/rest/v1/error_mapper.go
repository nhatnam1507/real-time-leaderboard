// Package v1 provides REST API v1 handlers for the leaderboard module.
package v1

import (
	"errors"

	"real-time-leaderboard/internal/shared/response"
	"real-time-leaderboard/internal/shared/validator"
)

// toAPIError converts leaderboard errors to APIError (internal helper)
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

	// If it's already an APIError, return it as-is
	if apiErr, ok := err.(*response.APIError); ok {
		return apiErr
	}

	// Default to internal error for wrapped errors
	return response.NewInternalError("An error occurred")
}
