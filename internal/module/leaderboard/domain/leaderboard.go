// Package domain provides domain entities for the leaderboard module.
package domain

// LeaderboardEntry represents a leaderboard entry
type LeaderboardEntry struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Score    int64  `json:"score"`
	Rank     int64  `json:"rank"`
}
