// Package repository provides repository implementations for the report module.
package repository

import (
	"context"
	"time"

	"real-time-leaderboard/internal/module/report/domain"
)

// CompositeReportRepository combines Redis and PostgreSQL repositories
type CompositeReportRepository struct {
	redisRepo    *RedisReportRepository
	postgresRepo *PostgresReportRepository
}

// NewCompositeReportRepository creates a new composite report repository
func NewCompositeReportRepository(redisRepo *RedisReportRepository, postgresRepo *PostgresReportRepository) *CompositeReportRepository {
	return &CompositeReportRepository{
		redisRepo:    redisRepo,
		postgresRepo: postgresRepo,
	}
}

// GetTopPlayers retrieves top players from Redis (current leaderboard)
func (r *CompositeReportRepository) GetTopPlayers(ctx context.Context, gameID string, limit, offset int64) ([]domain.TopPlayer, error) {
	return r.redisRepo.GetTopPlayers(ctx, gameID, limit, offset)
}

// GetTopPlayersByDateRange retrieves top players by date range
// Uses PostgreSQL for historical data, Redis for current data
func (r *CompositeReportRepository) GetTopPlayersByDateRange(ctx context.Context, gameID string, startDate, endDate time.Time, limit, offset int64) ([]domain.TopPlayer, error) {
	// Use PostgreSQL for historical data
	return r.postgresRepo.GetTopPlayersByDateRange(ctx, gameID, startDate, endDate, limit, offset)
}

// GetTotalPlayers retrieves total players from Redis
func (r *CompositeReportRepository) GetTotalPlayers(ctx context.Context, gameID string) (int64, error) {
	return r.redisRepo.GetTotalPlayers(ctx, gameID)
}
