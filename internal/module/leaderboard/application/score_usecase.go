// Package application provides use cases for the leaderboard module.
package application

import (
	"context"

	"real-time-leaderboard/internal/shared/logger"
	"real-time-leaderboard/internal/shared/response"
)

// ScoreUseCase handles score use cases
type ScoreUseCase struct {
	backupRepo       LeaderboardBackupRepository
	leaderboardRepo  LeaderboardRepository
	broadcastService BroadcastService
	logger           *logger.Logger
}

// NewScoreUseCase creates a new score use case
func NewScoreUseCase(
	backupRepo LeaderboardBackupRepository,
	leaderboardRepo LeaderboardRepository,
	broadcastService BroadcastService,
	l *logger.Logger,
) *ScoreUseCase {
	return &ScoreUseCase{
		backupRepo:       backupRepo,
		leaderboardRepo:  leaderboardRepo,
		broadcastService: broadcastService,
		logger:           l,
	}
}

// SubmitScoreRequest represents a score submission request
type SubmitScoreRequest struct {
	Score int64 `json:"score" validate:"required,gte=0" example:"1000"`
}

// SubmitScore upserts the score for a user
func (uc *ScoreUseCase) SubmitScore(ctx context.Context, userID string, req SubmitScoreRequest) error {
	if err := uc.backupRepo.UpsertScore(ctx, userID, req.Score); err != nil {
		uc.logger.Errorf(ctx, "Failed to upsert score: %v", err)
		return response.NewInternalError("Failed to update score", err)
	}

	if err := uc.leaderboardRepo.UpdateScore(ctx, userID, req.Score); err != nil {
		uc.logger.Errorf(ctx, "Failed to update leaderboard: %v", err)
		return nil
	}

	if err := uc.broadcastService.PublishScoreUpdate(ctx); err != nil {
		uc.logger.Warnf(ctx, "Failed to publish score update notification: %v", err)
	}

	uc.logger.Infof(ctx, "Score updated: user=%s, score=%d", userID, req.Score)
	return nil
}
