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

// GetGlobalLeaderboard retrieves the global leaderboard
func (uc *LeaderboardUseCase) GetGlobalLeaderboard(ctx context.Context, listReq *request.ListRequest) (*domain.Leaderboard, error) {
	limit := int64(listReq.GetLimit())
	offset := int64(listReq.GetOffset())

	entries, err := uc.leaderboardRepo.GetTopPlayers(ctx, "global", limit, offset)
	if err != nil {
		uc.logger.Errorf(ctx, "Failed to get global leaderboard: %v", err)
		return nil, response.NewInternalError("Failed to retrieve leaderboard", err)
	}

	total, err := uc.leaderboardRepo.GetTotalPlayers(ctx, "global")
	if err != nil {
		uc.logger.Warnf(ctx, "Failed to get total players: %v", err)
		total = int64(len(entries))
	}

	return &domain.Leaderboard{
		GameID:  "global",
		Entries: entries,
		Total:   total,
	}, nil
}

// GetGameLeaderboard retrieves a game-specific leaderboard
func (uc *LeaderboardUseCase) GetGameLeaderboard(ctx context.Context, gameID string, listReq *request.ListRequest) (*domain.Leaderboard, error) {
	limit := int64(listReq.GetLimit())
	offset := int64(listReq.GetOffset())

	entries, err := uc.leaderboardRepo.GetTopPlayers(ctx, gameID, limit, offset)
	if err != nil {
		uc.logger.Errorf(ctx, "Failed to get game leaderboard: %v", err)
		return nil, response.NewInternalError("Failed to retrieve leaderboard", err)
	}

	total, err := uc.leaderboardRepo.GetTotalPlayers(ctx, gameID)
	if err != nil {
		uc.logger.Warnf(ctx, "Failed to get total players: %v", err)
		total = int64(len(entries))
	}

	return &domain.Leaderboard{
		GameID:  gameID,
		Entries: entries,
		Total:   total,
	}, nil
}

// GetUserRank retrieves a user's rank in a leaderboard
func (uc *LeaderboardUseCase) GetUserRank(ctx context.Context, gameID string, userID string) (*domain.LeaderboardEntry, error) {
	entry, err := uc.leaderboardRepo.GetUserRank(ctx, gameID, userID)
	if err != nil {
		uc.logger.Errorf(ctx, "Failed to get user rank: %v", err)
		return nil, response.NewInternalError("Failed to retrieve user rank", err)
	}

	if entry == nil {
		return nil, response.NewNotFoundError("User not found in leaderboard")
	}

	return entry, nil
}

// GetGlobalLeaderboardTopN retrieves the top N players from global leaderboard (no pagination)
// This is used for WebSocket broadcasts where pagination is not needed
func (uc *LeaderboardUseCase) GetGlobalLeaderboardTopN(ctx context.Context, limit int64) (*domain.Leaderboard, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	entries, err := uc.leaderboardRepo.GetTopPlayers(ctx, "global", limit, 0)
	if err != nil {
		uc.logger.Errorf(ctx, "Failed to get global leaderboard: %v", err)
		return nil, response.NewInternalError("Failed to retrieve leaderboard", err)
	}

	total, err := uc.leaderboardRepo.GetTotalPlayers(ctx, "global")
	if err != nil {
		uc.logger.Warnf(ctx, "Failed to get total players: %v", err)
		total = int64(len(entries))
	}

	return &domain.Leaderboard{
		GameID:  "global",
		Entries: entries,
		Total:   total,
	}, nil
}

// GetGameLeaderboardTopN retrieves the top N players from a game-specific leaderboard (no pagination)
// This is used for WebSocket broadcasts where pagination is not needed
func (uc *LeaderboardUseCase) GetGameLeaderboardTopN(ctx context.Context, gameID string, limit int64) (*domain.Leaderboard, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	entries, err := uc.leaderboardRepo.GetTopPlayers(ctx, gameID, limit, 0)
	if err != nil {
		uc.logger.Errorf(ctx, "Failed to get game leaderboard: %v", err)
		return nil, response.NewInternalError("Failed to retrieve leaderboard", err)
	}

	total, err := uc.leaderboardRepo.GetTotalPlayers(ctx, gameID)
	if err != nil {
		uc.logger.Warnf(ctx, "Failed to get total players: %v", err)
		total = int64(len(entries))
	}

	return &domain.Leaderboard{
		GameID:  gameID,
		Entries: entries,
		Total:   total,
	}, nil
}
