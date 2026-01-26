// Package application provides use cases for the leaderboard module.
package application

import (
	"context"
	"fmt"

	"real-time-leaderboard/internal/module/leaderboard/domain"
	"real-time-leaderboard/internal/shared/logger"
)

//go:generate mockgen -destination=../adapters/mocks/leaderboard_usecase_mock.go -package=mocks real-time-leaderboard/internal/module/leaderboard/application LeaderboardUseCase

// LeaderboardUseCase defines the interface for leaderboard operations
type LeaderboardUseCase interface {
	GetLeaderboard(ctx context.Context, limit, offset int64) ([]domain.LeaderboardEntry, int64, error)
	SubscribeToEntryUpdates(ctx context.Context) (<-chan *domain.LeaderboardEntry, error)
}

// leaderboardUseCase implements LeaderboardUseCase interface
type leaderboardUseCase struct {
	cacheRepo        LeaderboardCacheRepository
	persistenceRepo  LeaderboardPersistenceRepository
	userRepo         UserRepository
	broadcastService BroadcastService
	logger           *logger.Logger
}

// NewLeaderboardUseCase creates a new leaderboard use case
//
//nolint:revive // unexported-return: intentional design - accept interface, return struct
func NewLeaderboardUseCase(
	cacheRepo LeaderboardCacheRepository,
	persistenceRepo LeaderboardPersistenceRepository,
	userRepo UserRepository,
	broadcastService BroadcastService,
	l *logger.Logger,
) *leaderboardUseCase {
	return &leaderboardUseCase{
		cacheRepo:        cacheRepo,
		persistenceRepo:  persistenceRepo,
		userRepo:         userRepo,
		broadcastService: broadcastService,
		logger:           l,
	}
}

// GetLeaderboard retrieves a paginated leaderboard with username enrichment.
// Cache-aside strategy: tries cache first with limit/offset; on cache miss or empty cache loads paginated data from persistence, backfills cache with fetched entries, and returns results.
func (uc *leaderboardUseCase) GetLeaderboard(ctx context.Context, limit, offset int64) ([]domain.LeaderboardEntry, int64, error) {
	// Try cache first
	entries, total, err := uc.cacheRepo.GetLeaderboard(ctx, limit, offset)
	if err == nil && total > 0 {
		// Cache hit with data - enrich and return
		if err := uc.enrichEntriesWithUsernames(ctx, entries); err != nil {
			uc.logger.Warnf(ctx, "Failed to enrich entries with usernames: %v", err)
		}
		return entries, total, nil
	}

	// Cache miss, empty, or error - fallback to database
	if err != nil {
		uc.logger.Warnf(ctx, "Cache error, falling back to database: %v", err)
	} else {
		uc.logger.Warnf(ctx, "Cache empty, falling back to database")
	}

	entries, total, err = uc.persistenceRepo.GetLeaderboard(ctx, limit, offset)
	if err != nil {
		uc.logger.Errorf(ctx, "Failed to get leaderboard from persistence: %v", err)
		return nil, 0, fmt.Errorf("failed to retrieve leaderboard: %w", err)
	}

	// Backfill cache with fetched entries
	for _, e := range entries {
		if err := uc.cacheRepo.UpdateScore(ctx, e.UserID, e.Score); err != nil {
			uc.logger.Warnf(ctx, "Failed to backfill cache for user %s: %v", e.UserID, err)
		}
	}

	// Enrich with usernames
	if err := uc.enrichEntriesWithUsernames(ctx, entries); err != nil {
		uc.logger.Warnf(ctx, "Failed to enrich entries with usernames: %v", err)
	}

	return entries, total, nil
}

func (uc *leaderboardUseCase) enrichEntriesWithUsernames(ctx context.Context, entries []domain.LeaderboardEntry) error {
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
func (uc *leaderboardUseCase) SubscribeToEntryUpdates(ctx context.Context) (<-chan *domain.LeaderboardEntry, error) {
	return uc.broadcastService.SubscribeToEntryUpdates(ctx)
}
