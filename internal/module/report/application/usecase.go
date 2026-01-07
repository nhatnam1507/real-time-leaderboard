// Package application provides use cases for the report module.
package application

import (
	"context"
	"time"

	"real-time-leaderboard/internal/module/report/domain"
	"real-time-leaderboard/internal/shared/request"
	"real-time-leaderboard/internal/shared/response"
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
	request.ListRequest
	GameID    string
	StartDate *time.Time
	EndDate   *time.Time
}

// GetTopPlayersReport generates a top players report
func (uc *ReportUseCase) GetTopPlayersReport(ctx context.Context, req GetTopPlayersReportRequest) (*domain.TopPlayersReport, error) {
	// Extract pagination values from embedded ListRequest
	limit := int64(req.GetLimit())
	offset := int64(req.GetOffset())

	var players []domain.TopPlayer
	var err error

	// If date range is provided, use PostgreSQL for historical data
	if req.StartDate != nil && req.EndDate != nil {
		players, err = uc.reportRepo.GetTopPlayersByDateRange(ctx, req.GameID, *req.StartDate, *req.EndDate, limit, offset)
	} else {
		// Otherwise, use Redis for current leaderboard
		players, err = uc.reportRepo.GetTopPlayers(ctx, req.GameID, limit, offset)
	}

	if err != nil {
		uc.logger.Errorf(ctx, "Failed to get top players: %v", err)
		return nil, response.NewInternalError("Failed to generate report", err)
	}

	total, err := uc.reportRepo.GetTotalPlayers(ctx, req.GameID)
	if err != nil {
		uc.logger.Warnf(ctx, "Failed to get total players: %v", err)
		total = int64(len(players))
	}

	report := &domain.TopPlayersReport{
		GameID:  req.GameID,
		Players: players,
		Total:   total,
	}

	if req.StartDate != nil {
		report.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		report.EndDate = *req.EndDate
	}

	return report, nil
}
