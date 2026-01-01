package domain

import (
	"context"
	"time"
)

// ReportRepository defines the interface for report operations
type ReportRepository interface {
	GetTopPlayers(ctx context.Context, gameID string, limit int64) ([]TopPlayer, error)
	GetTopPlayersByDateRange(ctx context.Context, gameID string, startDate, endDate time.Time, limit int64) ([]TopPlayer, error)
	GetTotalPlayers(ctx context.Context, gameID string) (int64, error)
}
