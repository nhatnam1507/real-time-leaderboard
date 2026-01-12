// Package application provides use cases for the leaderboard module.
package application

import (
	"context"

	"real-time-leaderboard/internal/module/leaderboard/domain"
	"real-time-leaderboard/internal/shared/logger"
	"real-time-leaderboard/internal/shared/response"
)

// ScoreUseCase handles score use cases
type ScoreUseCase struct {
	backupRepo      domain.LeaderboardBackupRepository
	leaderboardRepo domain.LeaderboardRepository
	logger          *logger.Logger
}

// NewScoreUseCase creates a new score use case
func NewScoreUseCase(
	backupRepo domain.LeaderboardBackupRepository,
	leaderboardRepo domain.LeaderboardRepository,
	l *logger.Logger,
) *ScoreUseCase {
	return &ScoreUseCase{
		backupRepo:      backupRepo,
		leaderboardRepo: leaderboardRepo,
		logger:          l,
	}
}

// SubmitScoreRequest represents a score submission request
type SubmitScoreRequest struct {
	Point int64 `json:"point" validate:"required,gte=0" example:"1000"`
}

// SubmitScore upserts the score for a user
// Creates record if it doesn't exist, updates if it exists
func (uc *ScoreUseCase) SubmitScore(ctx context.Context, userID string, req SubmitScoreRequest) error {
	// Upsert score in PostgreSQL (creates if not exists, updates if exists)
	// This serves as backup/recovery mechanism for Redis
	if err := uc.backupRepo.UpsertScore(ctx, userID, req.Point); err != nil {
		uc.logger.Errorf(ctx, "Failed to upsert score: %v", err)
		return response.NewInternalError("Failed to update score", err)
	}

	// Update Redis leaderboard with the new score (publishes to Redis pub/sub)
	// This is the source of truth for real-time leaderboard queries
	if err := uc.leaderboardRepo.UpdateScore(ctx, userID, req.Point); err != nil {
		uc.logger.Errorf(ctx, "Failed to update leaderboard: %v", err)
		// Don't fail the request if leaderboard update fails - PostgreSQL backup is still updated
	}

	uc.logger.Infof(ctx, "Score updated: user=%s, point=%d", userID, req.Point)
	return nil
}
