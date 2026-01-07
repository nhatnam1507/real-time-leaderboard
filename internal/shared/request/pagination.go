// Package request provides common request structures for API endpoints.
package request

// PaginationRequest represents pagination parameters for list endpoints
type PaginationRequest struct {
	Offset int `json:"offset" form:"offset" validate:"min=0"`
	Limit  int `json:"limit" form:"limit" validate:"min=1,max=100"`
}

const (
	// DefaultLimit is the default number of items per page
	DefaultLimit = 10
	// DefaultOffset is the default offset value
	DefaultOffset = 0
	// MaxLimit is the maximum number of items per page
	MaxLimit = 100
	// MinLimit is the minimum number of items per page
	MinLimit = 1
)

// Normalize applies default values and enforces bounds for pagination parameters
func (p *PaginationRequest) Normalize() *PaginationRequest {
	if p.Limit <= 0 {
		p.Limit = DefaultLimit
	}
	if p.Limit > MaxLimit {
		p.Limit = MaxLimit
	}
	if p.Limit < MinLimit {
		p.Limit = MinLimit
	}
	if p.Offset < 0 {
		p.Offset = DefaultOffset
	}
	return p
}

// GetOffset returns the normalized offset value
func (p *PaginationRequest) GetOffset() int {
	normalized := p.Normalize()
	return normalized.Offset
}

// GetLimit returns the normalized limit value
func (p *PaginationRequest) GetLimit() int {
	normalized := p.Normalize()
	return normalized.Limit
}

