package application

import (
	"context"
	"encoding/json"

	"real-time-leaderboard/internal/module/score/domain"
	"real-time-leaderboard/internal/shared/errors"
	"real-time-leaderboard/internal/shared/logger"
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
	GameID   string          `json:"game_id" validate:"required"`
	Score    int64           `json:"score" validate:"required,gte=0"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

// SubmitScore submits a new score
func (uc *ScoreUseCase) SubmitScore(ctx context.Context, userID string, req SubmitScoreRequest) (*domain.Score, error) {
	// Create score entity
	score := &domain.Score{
		UserID: userID,
		GameID: req.GameID,
		Score:  req.Score,
	}

	if req.Metadata != nil {
		score.Metadata = req.Metadata
	}

	// Save score to database
	if err := uc.scoreRepo.Create(ctx, score); err != nil {
		uc.logger.Errorf("Failed to create score: %v", err)
		return nil, errors.NewInternalError("Failed to submit score", err)
	}

	// Update leaderboard in Redis (use highest score for leaderboard)
	highestScore, err := uc.scoreRepo.GetHighestByUserIDAndGameID(ctx, userID, req.GameID)
	if err != nil {
		uc.logger.Errorf("Failed to get highest score: %v", err)
		return nil, errors.NewInternalError("Failed to submit score", err)
	}

	if highestScore != nil {
		// Update game-specific leaderboard
		if err := uc.leaderboardRepo.UpdateScore(ctx, req.GameID, userID, highestScore.Score); err != nil {
			uc.logger.Errorf("Failed to update game leaderboard: %v", err)
			// Don't fail the request if leaderboard update fails
		}

		// Update global leaderboard (sum of all game scores or highest)
		// For simplicity, we'll use the highest score across all games
		globalHighest, err := uc.scoreRepo.GetHighestByUserID(ctx, userID)
		if err == nil && globalHighest != nil {
			if err := uc.leaderboardRepo.UpdateGlobalScore(ctx, userID, globalHighest.Score); err != nil {
				uc.logger.Errorf("Failed to update global leaderboard: %v", err)
				// Don't fail the request if leaderboard update fails
			}
		}
	}

	uc.logger.Infof("Score submitted: user=%s, game=%s, score=%d", userID, req.GameID, req.Score)
	return score, nil
}

// GetUserScores retrieves scores for a user
func (uc *ScoreUseCase) GetUserScores(ctx context.Context, userID string, gameID string, limit, offset int) ([]*domain.Score, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	var scores []*domain.Score
	var err error

	if gameID != "" {
		scores, err = uc.scoreRepo.GetByUserIDAndGameID(ctx, userID, gameID, limit, offset)
	} else {
		scores, err = uc.scoreRepo.GetByUserID(ctx, userID, limit, offset)
	}

	if err != nil {
		uc.logger.Errorf("Failed to get user scores: %v", err)
		return nil, errors.NewInternalError("Failed to retrieve scores", err)
	}

	return scores, nil
}
