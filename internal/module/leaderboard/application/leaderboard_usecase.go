// Package application provides use cases for the leaderboard module.
package application

import (
	"context"

	"real-time-leaderboard/internal/module/leaderboard/domain"
	"real-time-leaderboard/internal/shared/logger"
	"real-time-leaderboard/internal/shared/request"
	"real-time-leaderboard/internal/shared/response"
)

// LeaderboardUseCase handles leaderboard use cases
type LeaderboardUseCase struct {
	leaderboardRepo domain.LeaderboardRepository
	logger          *logger.Logger
}

// NewLeaderboardUseCase creates a new leaderboard use case
func NewLeaderboardUseCase(leaderboardRepo domain.LeaderboardRepository, l *logger.Logger) *LeaderboardUseCase {
	return &LeaderboardUseCase{
		leaderboardRepo: leaderboardRepo,
		logger:          l,
	}
}

// GetLeaderboard retrieves the leaderboard
func (uc *LeaderboardUseCase) GetLeaderboard(ctx context.Context, listReq *request.ListRequest) (*domain.Leaderboard, error) {
	limit := int64(listReq.GetLimit())
	offset := int64(listReq.GetOffset())

	entries, err := uc.leaderboardRepo.GetTopPlayers(ctx, limit, offset)
	if err != nil {
		uc.logger.Errorf(ctx, "Failed to get leaderboard: %v", err)
		return nil, response.NewInternalError("Failed to retrieve leaderboard", err)
	}

	total, err := uc.leaderboardRepo.GetTotalPlayers(ctx)
	if err != nil {
		uc.logger.Warnf(ctx, "Failed to get total players: %v", err)
		total = int64(len(entries))
	}

	return &domain.Leaderboard{
		Entries: entries,
		Total:   total,
	}, nil
}
