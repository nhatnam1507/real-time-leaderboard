package domain

import "context"

// ScoreRepository defines the interface for score data operations
type ScoreRepository interface {
	Create(ctx context.Context, score *Score) error
	GetHighestByUserID(ctx context.Context, userID string) (*Score, error)
}

// LeaderboardRepository defines the interface for leaderboard operations
type LeaderboardRepository interface {
	UpdateScore(ctx context.Context, userID string, score int64) error
	GetTopPlayers(ctx context.Context, limit, offset int64) ([]LeaderboardEntry, error)
	GetTotalPlayers(ctx context.Context) (int64, error)
}
