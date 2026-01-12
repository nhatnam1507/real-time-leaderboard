package domain

import "context"

// UserRepository defines the interface for user data operations in the leaderboard module
// This interface belongs to the leaderboard module domain, not the auth module
type UserRepository interface {
	GetByIDs(ctx context.Context, userIDs []string) (map[string]string, error)
	// Returns map[userID]username for efficient batch fetching
}

// LeaderboardBackupRepository defines the interface for leaderboard backup operations in PostgreSQL
// This stores only the highest score per user as a backup/recovery mechanism for Redis
type LeaderboardBackupRepository interface {
	UpsertScore(ctx context.Context, userID string, score int64) error
	GetLeaderboard(ctx context.Context) (*Leaderboard, error)
}

// LeaderboardRepository defines the interface for leaderboard operations in Redis
type LeaderboardRepository interface {
	UpdateScore(ctx context.Context, userID string, score int64) error
	GetTopPlayers(ctx context.Context, limit, offset int64) ([]LeaderboardEntry, error)
	GetTotalPlayers(ctx context.Context) (int64, error)
}
