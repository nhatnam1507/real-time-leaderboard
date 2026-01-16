// Package repository provides repository implementations for the leaderboard module.
package repository

import (
	"context"
	"fmt"

	"real-time-leaderboard/internal/module/leaderboard/application"
	"real-time-leaderboard/internal/module/leaderboard/domain"

	"github.com/redis/go-redis/v9"
)


// RedisLeaderboardRepository implements LeaderboardCacheRepository using Redis sorted sets
type RedisLeaderboardRepository struct {
	client *redis.Client
}

// NewRedisLeaderboardRepository creates a new Redis leaderboard cache repository
func NewRedisLeaderboardRepository(client *redis.Client) application.LeaderboardCacheRepository {
	return &RedisLeaderboardRepository{client: client}
}

// UpdateScore updates the score in the leaderboard (does not publish notifications)
func (r *RedisLeaderboardRepository) UpdateScore(ctx context.Context, userID string, score int64) error {
	err := r.client.ZAdd(ctx, domain.RedisLeaderboardKey, redis.Z{
		Score:  float64(score),
		Member: userID,
	}).Err()
	if err != nil {
		return fmt.Errorf("failed to update score in leaderboard: %w", err)
	}

	return nil
}

// GetTopPlayers retrieves the top N players from the leaderboard with offset support
func (r *RedisLeaderboardRepository) GetTopPlayers(ctx context.Context, limit, offset int64) ([]domain.LeaderboardEntry, error) {
	start := offset
	stop := offset + limit - 1

	results, err := r.client.ZRevRangeWithScores(ctx, domain.RedisLeaderboardKey, start, stop).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get top players: %w", err)
	}

	entries := make([]domain.LeaderboardEntry, 0, len(results))
	for i, result := range results {
		userID, ok := result.Member.(string)
		if !ok {
			continue
		}

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
	count, err := r.client.ZCard(ctx, domain.RedisLeaderboardKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get total players: %w", err)
	}

	return count, nil
}

// GetUserRank retrieves the rank of a user in the leaderboard (1-indexed)
func (r *RedisLeaderboardRepository) GetUserRank(ctx context.Context, userID string) (int64, error) {
	rank, err := r.client.ZRevRank(ctx, domain.RedisLeaderboardKey, userID).Result()
	if err != nil {
		if err == redis.Nil {
			// User not found in leaderboard
			return 0, fmt.Errorf("user not found in leaderboard")
		}
		return 0, fmt.Errorf("failed to get user rank: %w", err)
	}

	// ZRevRank returns 0-based rank, convert to 1-based
	return rank + 1, nil
}
