// Package request provides common request structures for API endpoints.
package request

// Pagination represents pagination parameters for list endpoints
type Pagination struct {
	Offset int64 `json:"offset" form:"offset" validate:"min=0"`
	Limit  int64 `json:"limit" form:"limit" validate:"min=1,max=100"`
}

const (
	// DefaultLimit is the default number of items per page
	DefaultLimit int64 = 10
	// DefaultOffset is the default offset value
	DefaultOffset int64 = 0
	// MaxLimit is the maximum number of items per page
	MaxLimit int64 = 100
	// MinLimit is the minimum number of items per page
	MinLimit int64 = 1
)

// Normalize applies default values and enforces bounds for pagination parameters
func (p *Pagination) Normalize() *Pagination {
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
func (p *Pagination) GetOffset() int64 {
	normalized := p.Normalize()
	return normalized.Offset
}

// GetLimit returns the normalized limit value
func (p *Pagination) GetLimit() int64 {
	normalized := p.Normalize()
	return normalized.Limit
}
