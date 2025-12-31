package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"real-time-leaderboard/internal/module/report/domain"
)

// RedisReportRepository implements ReportRepository using Redis
type RedisReportRepository struct {
	client *redis.Client
}

// NewRedisReportRepository creates a new Redis report repository
func NewRedisReportRepository(client *redis.Client) *RedisReportRepository {
	return &RedisReportRepository{client: client}
}

// getKey returns the Redis key for a leaderboard
func (r *RedisReportRepository) getKey(gameID string) string {
	if gameID == "" || gameID == "global" {
		return "leaderboard:global"
	}
	return fmt.Sprintf("leaderboard:%s", gameID)
}

// GetTopPlayers retrieves the top N players from a leaderboard
func (r *RedisReportRepository) GetTopPlayers(ctx context.Context, gameID string, limit int64) ([]domain.TopPlayer, error) {
	key := r.getKey(gameID)

	// Get top players (highest scores first)
	results, err := r.client.ZRevRangeWithScores(ctx, key, 0, limit-1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get top players: %w", err)
	}

	players := make([]domain.TopPlayer, 0, len(results))
	for rank, result := range results {
		userID, ok := result.Member.(string)
		if !ok {
			continue
		}

		players = append(players, domain.TopPlayer{
			UserID:      userID,
			Score:       int64(result.Score),
			Rank:        int64(rank + 1), // Rank is 1-indexed
			GameID:      gameID,
			LastUpdated: time.Now(),
		})
	}

	return players, nil
}

// GetTopPlayersByDateRange retrieves top players (Redis doesn't support date ranges directly,
// so we return current top players. For historical data, use PostgreSQL)
func (r *RedisReportRepository) GetTopPlayersByDateRange(ctx context.Context, gameID string, startDate, endDate time.Time, limit int64) ([]domain.TopPlayer, error) {
	// Redis sorted sets don't support date ranges directly
	// For now, return current top players
	// In a real implementation, you might query PostgreSQL for historical data
	return r.GetTopPlayers(ctx, gameID, limit)
}

// GetTotalPlayers retrieves the total number of players
func (r *RedisReportRepository) GetTotalPlayers(ctx context.Context, gameID string) (int64, error) {
	key := r.getKey(gameID)

	count, err := r.client.ZCard(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get total players: %w", err)
	}

	return count, nil
}

