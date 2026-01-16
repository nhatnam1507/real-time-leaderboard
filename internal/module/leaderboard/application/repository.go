package application

//go:generate mockgen -source=repository.go -destination=../mocks/repository_mock.go -package=mocks UserRepository,LeaderboardPersistenceRepository,LeaderboardCacheRepository

import (
	"context"

	"real-time-leaderboard/internal/module/leaderboard/domain"
)

// UserRepository defines the interface for user data operations in the leaderboard module
// This interface belongs to the leaderboard module application layer, not the auth module
type UserRepository interface {
	GetByIDs(ctx context.Context, userIDs []string) (map[string]string, error)
	// Returns map[userID]username for efficient batch fetching
}

// LeaderboardPersistenceRepository defines the interface for persistent leaderboard storage in PostgreSQL
// This stores the highest score per user as persistent storage
type LeaderboardPersistenceRepository interface {
	UpsertScore(ctx context.Context, userID string, score int64) error
	GetLeaderboard(ctx context.Context) ([]domain.LeaderboardEntry, error)
}

// LeaderboardCacheRepository defines the interface for leaderboard cache operations in Redis
type LeaderboardCacheRepository interface {
	UpdateScore(ctx context.Context, userID string, score int64) error
	GetTopPlayers(ctx context.Context, limit, offset int64) ([]domain.LeaderboardEntry, error)
	GetTotalPlayers(ctx context.Context) (int64, error)
	GetUserRank(ctx context.Context, userID string) (int64, error)
}
