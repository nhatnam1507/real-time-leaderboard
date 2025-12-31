package domain

import (
	"encoding/json"
	"time"
)

// Score represents a score entity
type Score struct {
	ID          string          `json:"id"`
	UserID      string          `json:"user_id"`
	GameID      string          `json:"game_id"`
	Score       int64           `json:"score"`
	SubmittedAt time.Time      `json:"submitted_at"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
}

