// Package domain provides domain entities for the leaderboard module.
package domain

import "time"

// Score represents a score entity
type Score struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Point     int64     `json:"point"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LeaderboardEntry represents a leaderboard entry
type LeaderboardEntry struct {
	UserID string `json:"user_id"`
	Score  int64  `json:"score"`
	Rank   int64  `json:"rank"`
}

// Leaderboard represents a leaderboard
type Leaderboard struct {
	Entries []LeaderboardEntry `json:"entries"`
	Total   int64              `json:"total"`
}
