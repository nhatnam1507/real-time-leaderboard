package domain

import "time"

// TopPlayer represents a top player in a report
type TopPlayer struct {
	UserID      string    `json:"user_id"`
	Score       int64     `json:"score"`
	Rank        int64     `json:"rank"`
	GameID      string    `json:"game_id,omitempty"`
	LastUpdated time.Time `json:"last_updated,omitempty"`
}

// TopPlayersReport represents a top players report
type TopPlayersReport struct {
	GameID    string      `json:"game_id,omitempty"`
	StartDate time.Time   `json:"start_date,omitempty"`
	EndDate   time.Time   `json:"end_date,omitempty"`
	Players   []TopPlayer `json:"players"`
	Total     int64       `json:"total"`
}
