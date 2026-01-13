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
	leaderboardRepo LeaderboardRepository
	backupRepo      LeaderboardBackupRepository
	userRepo        UserRepository
	logger          *logger.Logger
}

// NewLeaderboardUseCase creates a new leaderboard use case
func NewLeaderboardUseCase(
	leaderboardRepo LeaderboardRepository,
	backupRepo LeaderboardBackupRepository,
	userRepo UserRepository,
	l *logger.Logger,
) *LeaderboardUseCase {
	return &LeaderboardUseCase{
		leaderboardRepo: leaderboardRepo,
		backupRepo:      backupRepo,
		userRepo:        userRepo,
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

	leaderboard, err := uc.backupRepo.GetLeaderboard(ctx)
	if err != nil {
		return err
	}

	for _, entry := range leaderboard.Entries {
		if err := uc.leaderboardRepo.UpdateScore(ctx, entry.UserID, entry.Score); err != nil {
			uc.logger.Warnf(ctx, "Failed to sync score for user %s: %v", entry.UserID, err)
		}
	}

	if len(leaderboard.Entries) > 0 {
		uc.logger.Infof(ctx, "Synced %d leaderboard entries from PostgreSQL to Redis", len(leaderboard.Entries))
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

	// Enrich entries with usernames
	if err := uc.enrichEntriesWithUsernames(ctx, entries); err != nil {
		uc.logger.Warnf(ctx, "Failed to enrich entries with usernames: %v", err)
		// Continue even if username enrichment fails
	}

	return &domain.Leaderboard{
		Entries: entries,
		Total:   total,
	}, nil
}

// enrichEntriesWithUsernames enriches leaderboard entries with usernames
func (uc *LeaderboardUseCase) enrichEntriesWithUsernames(ctx context.Context, entries []domain.LeaderboardEntry) error {
	if len(entries) == 0 {
		return nil
	}

	// Extract user IDs
	userIDs := make([]string, 0, len(entries))
	for _, entry := range entries {
		userIDs = append(userIDs, entry.UserID)
	}

	// Fetch usernames in batch
	usernames, err := uc.userRepo.GetByIDs(ctx, userIDs)
	if err != nil {
		return err
	}

	// Enrich entries with usernames
	for i := range entries {
		if username, ok := usernames[entries[i].UserID]; ok {
			entries[i].Username = username
		} else {
			// User not found - log warning and use empty string
			uc.logger.Warnf(ctx, "Username not found for user ID: %s", entries[i].UserID)
			entries[i].Username = ""
		}
	}

	return nil
}
