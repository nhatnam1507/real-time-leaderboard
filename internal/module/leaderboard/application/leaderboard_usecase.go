// Package application provides use cases for the leaderboard module.
package application

import (
	"context"

	"real-time-leaderboard/internal/module/leaderboard/domain"
	"real-time-leaderboard/internal/shared/logger"
	"real-time-leaderboard/internal/shared/response"
)

// LeaderboardUseCase handles leaderboard use cases
type LeaderboardUseCase struct {
	leaderboardRepo domain.LeaderboardRepository
	backupRepo      domain.LeaderboardBackupRepository
	logger          *logger.Logger
}

// NewLeaderboardUseCase creates a new leaderboard use case
func NewLeaderboardUseCase(
	leaderboardRepo domain.LeaderboardRepository,
	backupRepo domain.LeaderboardBackupRepository,
	l *logger.Logger,
) *LeaderboardUseCase {
	return &LeaderboardUseCase{
		leaderboardRepo: leaderboardRepo,
		backupRepo:      backupRepo,
		logger:          l,
	}
}

// SyncFromPostgres syncs all leaderboard entries from PostgreSQL to Redis
// This is called lazily when Redis is detected to be empty
func (uc *LeaderboardUseCase) SyncFromPostgres(ctx context.Context) error {
	total, err := uc.leaderboardRepo.GetTotalPlayers(ctx)
	if err != nil || total > 0 {
		return err
	}

	scores, err := uc.backupRepo.GetAllScores(ctx)
	if err != nil {
		return err
	}

	for _, score := range scores {
		if err := uc.leaderboardRepo.UpdateScore(ctx, score.UserID, score.Point); err != nil {
			uc.logger.Warnf(ctx, "Failed to sync score for user %s: %v", score.UserID, err)
		}
	}

	if len(scores) > 0 {
		uc.logger.Infof(ctx, "Synced %d leaderboard entries from PostgreSQL to Redis", len(scores))
	}
	return nil
}

// GetFullLeaderboard retrieves the full leaderboard (for initial fetch)
func (uc *LeaderboardUseCase) GetFullLeaderboard(ctx context.Context) (*domain.Leaderboard, error) {
	// Fetch with large limit to get full leaderboard
	entries, err := uc.leaderboardRepo.GetTopPlayers(ctx, 1000, 0)
	if err != nil {
		uc.logger.Errorf(ctx, "Failed to get full leaderboard: %v", err)
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
