package domain

import "context"

// LeaderboardBackupRepository defines the interface for leaderboard backup operations in PostgreSQL
// This stores only the highest score per user as a backup/recovery mechanism for Redis
type LeaderboardBackupRepository interface {
	UpsertScore(ctx context.Context, userID string, point int64) (*Score, error)
}

// LeaderboardRepository defines the interface for leaderboard operations in Redis
type LeaderboardRepository interface {
	UpdateScore(ctx context.Context, userID string, score int64) error
	GetTopPlayers(ctx context.Context, limit, offset int64) ([]LeaderboardEntry, error)
	GetTotalPlayers(ctx context.Context) (int64, error)
}
