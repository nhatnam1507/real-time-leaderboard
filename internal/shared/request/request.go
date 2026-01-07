// Package request provides common request structures for API endpoints.
package request

import (
	"real-time-leaderboard/internal/shared/validator"

	"github.com/gin-gonic/gin"
)

// ListRequest represents a common request structure for list endpoints
// It embeds PaginationRequest and can be extended with additional components
// like SortingRequest, FilteringRequest, SearchingRequest, etc.
type ListRequest struct {
	PaginationRequest
	// Future: SortingRequest, FilteringRequest, SearchingRequest, etc.
}

// FromGinContext parses the list request from Gin context query parameters
func (r *ListRequest) FromGinContext(c *gin.Context) error {
	if err := c.ShouldBindQuery(r); err != nil {
		return err
	}
	return nil
}

// Validate validates the list request
func (r *ListRequest) Validate() error {
	return validator.Validate(r)
}

