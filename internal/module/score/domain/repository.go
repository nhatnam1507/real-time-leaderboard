package domain

import "context"

// ScoreRepository defines the interface for score data operations
type ScoreRepository interface {
	Create(ctx context.Context, score *Score) error
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*Score, error)
	GetByUserIDAndGameID(ctx context.Context, userID, gameID string, limit, offset int) ([]*Score, error)
	GetHighestByUserID(ctx context.Context, userID string) (*Score, error)
	GetHighestByUserIDAndGameID(ctx context.Context, userID, gameID string) (*Score, error)
}

// LeaderboardRepository defines the interface for leaderboard operations
type LeaderboardRepository interface {
	UpdateScore(ctx context.Context, gameID string, userID string, score int64) error
	UpdateGlobalScore(ctx context.Context, userID string, score int64) error
}
