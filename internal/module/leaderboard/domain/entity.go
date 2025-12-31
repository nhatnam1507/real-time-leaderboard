package domain

// LeaderboardEntry represents a leaderboard entry
type LeaderboardEntry struct {
	UserID string `json:"user_id"`
	Score  int64  `json:"score"`
	Rank   int64  `json:"rank"`
}

// Leaderboard represents a leaderboard
type Leaderboard struct {
	GameID  string             `json:"game_id,omitempty"`
	Entries []LeaderboardEntry `json:"entries"`
	Total   int64              `json:"total"`
}

