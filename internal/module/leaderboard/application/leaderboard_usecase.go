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
	cacheRepo        LeaderboardCacheRepository
	persistenceRepo LeaderboardPersistenceRepository
	userRepo         UserRepository
	broadcastService BroadcastService
	logger           *logger.Logger
}

// NewLeaderboardUseCase creates a new leaderboard use case
func NewLeaderboardUseCase(
	cacheRepo LeaderboardCacheRepository,
	persistenceRepo LeaderboardPersistenceRepository,
	userRepo UserRepository,
	broadcastService BroadcastService,
	l *logger.Logger,
) *LeaderboardUseCase {
	return &LeaderboardUseCase{
		cacheRepo:        cacheRepo,
		persistenceRepo:  persistenceRepo,
		userRepo:         userRepo,
		broadcastService: broadcastService,
		logger:           l,
	}
}

// SyncFromPostgres syncs all leaderboard entries from PostgreSQL to Redis
// Called lazily when Redis is empty. UpdateScore doesn't publish, so no broadcasts triggered.
func (uc *LeaderboardUseCase) SyncFromPostgres(ctx context.Context) error {
	total, err := uc.cacheRepo.GetTotalPlayers(ctx)
	if err != nil || total > 0 {
		return err
	}

	entries, err := uc.persistenceRepo.GetLeaderboard(ctx)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if err := uc.cacheRepo.UpdateScore(ctx, entry.UserID, entry.Score); err != nil {
			uc.logger.Warnf(ctx, "Failed to sync score for user %s: %v", entry.UserID, err)
		}
	}

	if len(entries) > 0 {
		uc.logger.Infof(ctx, "Synced %d leaderboard entries from PostgreSQL to Redis", len(entries))
	}
	return nil
}

// GetFullLeaderboard retrieves the full leaderboard with username enrichment
func (uc *LeaderboardUseCase) GetFullLeaderboard(ctx context.Context) ([]domain.LeaderboardEntry, int64, error) {
	entries, err := uc.cacheRepo.GetTopPlayers(ctx, 1000, 0)
	if err != nil {
		uc.logger.Errorf(ctx, "Failed to get full leaderboard: %v", err)
		return nil, 0, response.NewInternalError("Failed to retrieve leaderboard", err)
	}

	total, err := uc.cacheRepo.GetTotalPlayers(ctx)
	if err != nil {
		uc.logger.Warnf(ctx, "Failed to get total players: %v", err)
		total = int64(len(entries))
	}

	if err := uc.enrichEntriesWithUsernames(ctx, entries); err != nil {
		uc.logger.Warnf(ctx, "Failed to enrich entries with usernames: %v", err)
	}

	return entries, total, nil
}

// GetLeaderboardPaginated retrieves a paginated leaderboard with username enrichment
func (uc *LeaderboardUseCase) GetLeaderboardPaginated(ctx context.Context, limit, offset int64) ([]domain.LeaderboardEntry, int64, error) {
	entries, err := uc.cacheRepo.GetTopPlayers(ctx, limit, offset)
	if err != nil {
		uc.logger.Errorf(ctx, "Failed to get paginated leaderboard: %v", err)
		return nil, 0, response.NewInternalError("Failed to retrieve leaderboard", err)
	}

	total, err := uc.cacheRepo.GetTotalPlayers(ctx)
	if err != nil {
		uc.logger.Warnf(ctx, "Failed to get total players: %v", err)
		total = int64(len(entries))
	}

	if err := uc.enrichEntriesWithUsernames(ctx, entries); err != nil {
		uc.logger.Warnf(ctx, "Failed to enrich entries with usernames: %v", err)
	}

	return entries, total, nil
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

// SubscribeToEntryUpdates subscribes to leaderboard entry delta update broadcasts for SSE handlers
func (uc *LeaderboardUseCase) SubscribeToEntryUpdates(ctx context.Context) (<-chan *domain.LeaderboardEntry, error) {
	return uc.broadcastService.SubscribeToEntryUpdates(ctx)
}
