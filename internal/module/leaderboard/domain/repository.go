package domain

import "context"

// LeaderboardRepository defines the interface for leaderboard operations
type LeaderboardRepository interface {
	GetTopPlayers(ctx context.Context, gameID string, limit int64) ([]LeaderboardEntry, error)
	GetUserRank(ctx context.Context, gameID string, userID string) (*LeaderboardEntry, error)
	GetUserScore(ctx context.Context, gameID string, userID string) (int64, error)
	GetTotalPlayers(ctx context.Context, gameID string) (int64, error)
}
