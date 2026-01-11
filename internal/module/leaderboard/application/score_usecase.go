// Package application provides use cases for the leaderboard module.
package application

import (
	"context"
	"encoding/json"

	"real-time-leaderboard/internal/module/leaderboard/domain"
	"real-time-leaderboard/internal/shared/logger"
	"real-time-leaderboard/internal/shared/response"
)

// ScoreUseCase handles score use cases
type ScoreUseCase struct {
	scoreRepo       domain.ScoreRepository
	leaderboardRepo domain.LeaderboardRepository
	logger          *logger.Logger
}

// NewScoreUseCase creates a new score use case
func NewScoreUseCase(
	scoreRepo domain.ScoreRepository,
	leaderboardRepo domain.LeaderboardRepository,
	l *logger.Logger,
) *ScoreUseCase {
	return &ScoreUseCase{
		scoreRepo:       scoreRepo,
		leaderboardRepo: leaderboardRepo,
		logger:          l,
	}
}

// SubmitScoreRequest represents a score submission request
type SubmitScoreRequest struct {
	Score    int64           `json:"score" validate:"required,gte=0" example:"1000"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

// SubmitScore submits a new score
func (uc *ScoreUseCase) SubmitScore(ctx context.Context, userID string, req SubmitScoreRequest) (*domain.Score, error) {
	// Create score entity
	score := &domain.Score{
		UserID: userID,
		Score:  req.Score,
	}

	if req.Metadata != nil {
		score.Metadata = req.Metadata
	}

	// Save score to database
	if err := uc.scoreRepo.Create(ctx, score); err != nil {
		uc.logger.Errorf(ctx, "Failed to create score: %v", err)
		return nil, response.NewInternalError("Failed to submit score", err)
	}

	// Get highest score for user to update leaderboard
	highestScore, err := uc.scoreRepo.GetHighestByUserID(ctx, userID)
	if err != nil {
		uc.logger.Errorf(ctx, "Failed to get highest score: %v", err)
		return nil, response.NewInternalError("Failed to submit score", err)
	}

	if highestScore != nil {
		// Update leaderboard (publishes to Redis pub/sub)
		if err := uc.leaderboardRepo.UpdateScore(ctx, userID, highestScore.Score); err != nil {
			uc.logger.Errorf(ctx, "Failed to update leaderboard: %v", err)
			// Don't fail the request if leaderboard update fails
		}
	}

	uc.logger.Infof(ctx, "Score submitted: user=%s, score=%d", userID, req.Score)
	return score, nil
}
