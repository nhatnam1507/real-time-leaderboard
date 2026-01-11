// Package repository provides repository implementations for the leaderboard module.
package repository

import (
	"context"
	"fmt"

	"real-time-leaderboard/internal/module/leaderboard/domain"

	"github.com/redis/go-redis/v9"
)

const (
	leaderboardKey = "leaderboard:global"
	updateChannel  = "leaderboard:updates"
)

// RedisLeaderboardRepository implements LeaderboardRepository using Redis sorted sets
type RedisLeaderboardRepository struct {
	client *redis.Client
}

// NewRedisLeaderboardRepository creates a new Redis leaderboard repository
func NewRedisLeaderboardRepository(client *redis.Client) *RedisLeaderboardRepository {
	return &RedisLeaderboardRepository{client: client}
}

// UpdateScore updates the score in the leaderboard
func (r *RedisLeaderboardRepository) UpdateScore(ctx context.Context, userID string, score int64) error {
	err := r.client.ZAdd(ctx, leaderboardKey, redis.Z{
		Score:  float64(score),
		Member: userID,
	}).Err()
	if err != nil {
		return fmt.Errorf("failed to update score in leaderboard: %w", err)
	}

	// Publish update notification to Redis pub/sub
	r.client.Publish(ctx, updateChannel, "updated")

	return nil
}

// GetTopPlayers retrieves the top N players from the leaderboard with offset support
func (r *RedisLeaderboardRepository) GetTopPlayers(ctx context.Context, limit, offset int64) ([]domain.LeaderboardEntry, error) {
	// Calculate Redis range: start = offset, stop = offset + limit - 1
	start := offset
	stop := offset + limit - 1

	// Get top players (highest scores first) with offset
	results, err := r.client.ZRevRangeWithScores(ctx, leaderboardKey, start, stop).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get top players: %w", err)
	}

	entries := make([]domain.LeaderboardEntry, 0, len(results))
	for i, result := range results {
		userID, ok := result.Member.(string)
		if !ok {
			continue
		}

		// Rank is 1-indexed and accounts for offset
		rank := offset + int64(i) + 1

		entries = append(entries, domain.LeaderboardEntry{
			UserID: userID,
			Score:  int64(result.Score),
			Rank:   rank,
		})
	}

	return entries, nil
}

// GetTotalPlayers retrieves the total number of players in the leaderboard
func (r *RedisLeaderboardRepository) GetTotalPlayers(ctx context.Context) (int64, error) {
	count, err := r.client.ZCard(ctx, leaderboardKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get total players: %w", err)
	}

	return count, nil
}
