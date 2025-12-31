package application

import (
	"context"
	"time"

	"real-time-leaderboard/internal/module/report/domain"
	"real-time-leaderboard/internal/shared/errors"
	"real-time-leaderboard/internal/shared/logger"
)

// ReportUseCase handles report use cases
type ReportUseCase struct {
	reportRepo domain.ReportRepository
	logger     *logger.Logger
}

// NewReportUseCase creates a new report use case
func NewReportUseCase(reportRepo domain.ReportRepository, l *logger.Logger) *ReportUseCase {
	return &ReportUseCase{
		reportRepo: reportRepo,
		logger:     l,
	}
}

// GetTopPlayersReportRequest represents a request for top players report
type GetTopPlayersReportRequest struct {
	GameID    string
	StartDate *time.Time
	EndDate   *time.Time
	Limit     int64
}

// GetTopPlayersReport generates a top players report
func (uc *ReportUseCase) GetTopPlayersReport(ctx context.Context, req GetTopPlayersReportRequest) (*domain.TopPlayersReport, error) {
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	var players []domain.TopPlayer
	var err error

	// If date range is provided, use PostgreSQL for historical data
	if req.StartDate != nil && req.EndDate != nil {
		players, err = uc.reportRepo.GetTopPlayersByDateRange(ctx, req.GameID, *req.StartDate, *req.EndDate, req.Limit)
	} else {
		// Otherwise, use Redis for current leaderboard
		players, err = uc.reportRepo.GetTopPlayers(ctx, req.GameID, req.Limit)
	}

	if err != nil {
		uc.logger.Errorf("Failed to get top players: %v", err)
		return nil, errors.NewInternalError("Failed to generate report", err)
	}

	total, err := uc.reportRepo.GetTotalPlayers(ctx, req.GameID)
	if err != nil {
		uc.logger.Warnf("Failed to get total players: %v", err)
		total = int64(len(players))
	}

	report := &domain.TopPlayersReport{
		GameID:    req.GameID,
		Players:   players,
		Total:     total,
	}

	if req.StartDate != nil {
		report.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		report.EndDate = *req.EndDate
	}

	return report, nil
}

