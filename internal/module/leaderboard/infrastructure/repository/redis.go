package repository

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"real-time-leaderboard/internal/module/leaderboard/domain"
)

// RedisLeaderboardRepository implements LeaderboardRepository using Redis sorted sets
type RedisLeaderboardRepository struct {
	client *redis.Client
}

// NewRedisLeaderboardRepository creates a new Redis leaderboard repository
func NewRedisLeaderboardRepository(client *redis.Client) *RedisLeaderboardRepository {
	return &RedisLeaderboardRepository{client: client}
}

// getKey returns the Redis key for a leaderboard
func (r *RedisLeaderboardRepository) getKey(gameID string) string {
	if gameID == "" || gameID == "global" {
		return "leaderboard:global"
	}
	return fmt.Sprintf("leaderboard:%s", gameID)
}

// GetTopPlayers retrieves the top N players from a leaderboard
func (r *RedisLeaderboardRepository) GetTopPlayers(ctx context.Context, gameID string, limit int64) ([]domain.LeaderboardEntry, error) {
	key := r.getKey(gameID)

	// Get top players (highest scores first)
	results, err := r.client.ZRevRangeWithScores(ctx, key, 0, limit-1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get top players: %w", err)
	}

	entries := make([]domain.LeaderboardEntry, 0, len(results))
	for rank, result := range results {
		userID, ok := result.Member.(string)
		if !ok {
			continue
		}

		entries = append(entries, domain.LeaderboardEntry{
			UserID: userID,
			Score:  int64(result.Score),
			Rank:   int64(rank + 1), // Rank is 1-indexed
		})
	}

	return entries, nil
}

// GetUserRank retrieves a user's rank in a leaderboard
func (r *RedisLeaderboardRepository) GetUserRank(ctx context.Context, gameID string, userID string) (*domain.LeaderboardEntry, error) {
	key := r.getKey(gameID)

	// Get user's rank (0-indexed, so we add 1)
	rank, err := r.client.ZRevRank(ctx, key, userID).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // User not found in leaderboard
		}
		return nil, fmt.Errorf("failed to get user rank: %w", err)
	}

	// Get user's score
	score, err := r.client.ZScore(ctx, key, userID).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user score: %w", err)
	}

	return &domain.LeaderboardEntry{
		UserID: userID,
		Score:  int64(score),
		Rank:   rank + 1, // Rank is 1-indexed
	}, nil
}

// GetUserScore retrieves a user's score in a leaderboard
func (r *RedisLeaderboardRepository) GetUserScore(ctx context.Context, gameID string, userID string) (int64, error) {
	key := r.getKey(gameID)

	score, err := r.client.ZScore(ctx, key, userID).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get user score: %w", err)
	}

	return int64(score), nil
}

// GetTotalPlayers retrieves the total number of players in a leaderboard
func (r *RedisLeaderboardRepository) GetTotalPlayers(ctx context.Context, gameID string) (int64, error) {
	key := r.getKey(gameID)

	count, err := r.client.ZCard(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get total players: %w", err)
	}

	return count, nil
}

