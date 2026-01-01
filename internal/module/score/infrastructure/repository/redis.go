package repository

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// RedisLeaderboardRepository implements LeaderboardRepository using Redis sorted sets
type RedisLeaderboardRepository struct {
	client *redis.Client
}

// NewRedisLeaderboardRepository creates a new Redis leaderboard repository
func NewRedisLeaderboardRepository(client *redis.Client) *RedisLeaderboardRepository {
	return &RedisLeaderboardRepository{client: client}
}

// UpdateScore updates the score in a game-specific leaderboard
func (r *RedisLeaderboardRepository) UpdateScore(ctx context.Context, gameID string, userID string, score int64) error {
	key := fmt.Sprintf("leaderboard:%s", gameID)
	err := r.client.ZAdd(ctx, key, redis.Z{
		Score:  float64(score),
		Member: userID,
	}).Err()
	if err != nil {
		return fmt.Errorf("failed to update score in leaderboard: %w", err)
	}
	return nil
}

// UpdateGlobalScore updates the score in the global leaderboard
func (r *RedisLeaderboardRepository) UpdateGlobalScore(ctx context.Context, userID string, score int64) error {
	key := "leaderboard:global"
	err := r.client.ZAdd(ctx, key, redis.Z{
		Score:  float64(score),
		Member: userID,
	}).Err()
	if err != nil {
		return fmt.Errorf("failed to update global score: %w", err)
	}
	return nil
}
