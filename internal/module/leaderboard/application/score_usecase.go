// Package application provides use cases for the leaderboard module.
package application

import (
	"context"
	"fmt"

	"real-time-leaderboard/internal/module/leaderboard/domain"
	"real-time-leaderboard/internal/shared/logger"
)

//go:generate mockgen -source=score_usecase.go -destination=../mocks/score_usecase_mock.go -package=mocks ScoreUseCase

// ScoreUseCase defines the interface for score operations
type ScoreUseCase interface {
	SubmitScore(ctx context.Context, userID string, req SubmitScoreRequest) error
}

// scoreUseCase implements ScoreUseCase interface
type scoreUseCase struct {
	persistenceRepo LeaderboardPersistenceRepository
	cacheRepo       LeaderboardCacheRepository
	userRepo        UserRepository
	broadcastService BroadcastService
	logger          *logger.Logger
}

// NewScoreUseCase creates a new score use case
func NewScoreUseCase(
	persistenceRepo LeaderboardPersistenceRepository,
	cacheRepo LeaderboardCacheRepository,
	userRepo UserRepository,
	broadcastService BroadcastService,
	l *logger.Logger,
) *scoreUseCase {
	return &scoreUseCase{
		persistenceRepo: persistenceRepo,
		cacheRepo:       cacheRepo,
		userRepo:        userRepo,
		broadcastService: broadcastService,
		logger:          l,
	}
}

// SubmitScoreRequest represents a score submission request
type SubmitScoreRequest struct {
	Score int64 `json:"score" validate:"required,gte=0" example:"1000"`
}

// SubmitScore upserts the score for a user in both persistence and cache systems.
// It publishes a leaderboard entry delta update after successful updates.
func (uc *scoreUseCase) SubmitScore(ctx context.Context, userID string, req SubmitScoreRequest) error {
	if err := uc.persistenceRepo.UpsertScore(ctx, userID, req.Score); err != nil {
		uc.logger.Errorf(ctx, "Failed to upsert score: %v", err)
		return fmt.Errorf("failed to update score: %w", err)
	}

	if err := uc.cacheRepo.UpdateScore(ctx, userID, req.Score); err != nil {
		uc.logger.Errorf(ctx, "Failed to update cache: %v", err)
		return nil
	}

	// Get user's rank after update
	rank, err := uc.cacheRepo.GetUserRank(ctx, userID)
	if err != nil {
		uc.logger.Warnf(ctx, "Failed to get user rank: %v", err)
		uc.logger.Infof(ctx, "Score updated: user=%s, score=%d", userID, req.Score)
		// Continue without broadcasting if rank fetch fails
		return nil
	}

	// Only broadcast if entry is within the broadcast threshold
	// This optimizes network traffic by skipping updates for very low-ranked entries
	if rank > domain.MaxBroadcastRank {
		uc.logger.Infof(ctx, "Score updated: user=%s, score=%d, rank=%d (outside broadcast range, skipping)", userID, req.Score, rank)
		return nil
	}

	// Get username
	usernames, err := uc.userRepo.GetByIDs(ctx, []string{userID})
	if err != nil {
		uc.logger.Warnf(ctx, "Failed to get username: %v", err)
	}

	username := ""
	if usernames != nil {
		if u, ok := usernames[userID]; ok {
			username = u
		}
	}

	// Create entry update
	entry := domain.LeaderboardEntry{
		UserID:   userID,
		Username: username,
		Score:    req.Score,
		Rank:     rank,
	}

	// Broadcast entry update
	if err := uc.broadcastService.BroadcastEntryUpdate(ctx, &entry); err != nil {
		uc.logger.Warnf(ctx, "Failed to broadcast entry update: %v", err)
	}

	uc.logger.Infof(ctx, "Score updated: user=%s, score=%d, rank=%d", userID, req.Score, rank)
	return nil
}
