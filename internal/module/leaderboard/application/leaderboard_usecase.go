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
	leaderboardRepo  LeaderboardRepository
	backupRepo       LeaderboardBackupRepository
	userRepo         UserRepository
	broadcastService BroadcastService
	logger           *logger.Logger
}

// NewLeaderboardUseCase creates a new leaderboard use case
func NewLeaderboardUseCase(
	leaderboardRepo LeaderboardRepository,
	backupRepo LeaderboardBackupRepository,
	userRepo UserRepository,
	broadcastService BroadcastService,
	l *logger.Logger,
) *LeaderboardUseCase {
	return &LeaderboardUseCase{
		leaderboardRepo:  leaderboardRepo,
		backupRepo:       backupRepo,
		userRepo:         userRepo,
		broadcastService: broadcastService,
		logger:           l,
	}
}

// SyncFromPostgres syncs all leaderboard entries from PostgreSQL to Redis
// Called lazily when Redis is empty. UpdateScore doesn't publish, so no broadcasts triggered.
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

// GetFullLeaderboard retrieves the full leaderboard with username enrichment
func (uc *LeaderboardUseCase) GetFullLeaderboard(ctx context.Context) (*domain.Leaderboard, error) {
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

	if err := uc.enrichEntriesWithUsernames(ctx, entries); err != nil {
		uc.logger.Warnf(ctx, "Failed to enrich entries with usernames: %v", err)
	}

	return &domain.Leaderboard{
		Entries: entries,
		Total:   total,
	}, nil
}

func (uc *LeaderboardUseCase) enrichEntriesWithUsernames(ctx context.Context, entries []domain.LeaderboardEntry) error {
	if len(entries) == 0 {
		return nil
	}

	userIDs := make([]string, 0, len(entries))
	for _, entry := range entries {
		userIDs = append(userIDs, entry.UserID)
	}

	usernames, err := uc.userRepo.GetByIDs(ctx, userIDs)
	if err != nil {
		return err
	}

	for i := range entries {
		if username, ok := usernames[entries[i].UserID]; ok {
			entries[i].Username = username
		} else {
			uc.logger.Warnf(ctx, "Username not found for user ID: %s", entries[i].UserID)
			entries[i].Username = ""
		}
	}

	return nil
}

// StartBroadcasting listens to score updates and broadcasts enriched leaderboard
// Runs in a goroutine. Subscribes to notifications, fetches/enriches leaderboard, broadcasts.
func (uc *LeaderboardUseCase) StartBroadcasting(ctx context.Context) error {
	scoreUpdateCh, err := uc.broadcastService.SubscribeToScoreUpdates(ctx)
	if err != nil {
		uc.logger.Errorf(ctx, "Failed to subscribe to score updates: %v", err)
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case _, ok := <-scoreUpdateCh:
			if !ok {
				return nil
			}

			leaderboard, err := uc.GetFullLeaderboard(ctx)
			if err != nil {
				uc.logger.Errorf(ctx, "Failed to get full leaderboard: %v", err)
				continue
			}

			if err := uc.broadcastService.BroadcastLeaderboard(ctx, leaderboard); err != nil {
				uc.logger.Errorf(ctx, "Failed to broadcast leaderboard: %v", err)
			}
		}
	}
}

// SubscribeToLeaderboardUpdates subscribes to leaderboard update broadcasts for SSE handlers
func (uc *LeaderboardUseCase) SubscribeToLeaderboardUpdates(ctx context.Context) (<-chan *domain.Leaderboard, error) {
	return uc.broadcastService.SubscribeToLeaderboardUpdates(ctx)
}
