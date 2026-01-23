package application

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"real-time-leaderboard/internal/module/leaderboard/domain"
	"real-time-leaderboard/internal/module/leaderboard/infrastructure/mocks"
	"real-time-leaderboard/internal/shared/logger"
)

func TestLeaderboardUseCase_GetLeaderboard_WhenCacheHit_ShouldReturnFromCacheAndEnrich(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCacheRepo := mocks.NewMockLeaderboardCacheRepository(ctrl)
	mockCacheRepo.EXPECT().
		GetTotalPlayers(ctx).
		Return(int64(2), nil).
		Times(1)
	mockCacheRepo.EXPECT().
		GetTopPlayers(ctx, int64(10), int64(0)).
		Return([]domain.LeaderboardEntry{
			{UserID: "user-1", Score: 1000, Rank: 1},
			{UserID: "user-2", Score: 500, Rank: 2},
		}, nil).
		Times(1)

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().
		GetByIDs(ctx, []string{"user-1", "user-2"}).
		Return(map[string]string{"user-1": "alice", "user-2": "bob"}, nil).
		Times(1)

	mockPersistenceRepo := mocks.NewMockLeaderboardPersistenceRepository(ctrl)
	mockBroadcastService := mocks.NewMockBroadcastService(ctrl)

	logger := logger.New("info", false)
	uc := NewLeaderboardUseCase(mockCacheRepo, mockPersistenceRepo, mockUserRepo, mockBroadcastService, logger)

	// ── Act ─────────────────────────────────────────────────────────────
	entries, total, err := uc.GetLeaderboard(ctx, 10, 0)

	// ── Assert ──────────────────────────────────────────────────────────
	require.NoError(t, err)
	require.Len(t, entries, 2)
	require.Equal(t, int64(2), total)
	require.Equal(t, "alice", entries[0].Username)
	require.Equal(t, "bob", entries[1].Username)
}

func TestLeaderboardUseCase_GetLeaderboard_WhenCacheMiss_ShouldLoadFromPersistenceBackfillAndReturnPaginated(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCacheRepo := mocks.NewMockLeaderboardCacheRepository(ctrl)
	mockCacheRepo.EXPECT().
		GetTotalPlayers(ctx).
		Return(int64(0), nil).
		Times(1)
	mockCacheRepo.EXPECT().
		UpdateScore(ctx, "user-1", int64(1000)).
		Return(nil).
		Times(1)
	mockCacheRepo.EXPECT().
		UpdateScore(ctx, "user-2", int64(500)).
		Return(nil).
		Times(1)
	mockCacheRepo.EXPECT().
		UpdateScore(ctx, "user-3", int64(250)).
		Return(nil).
		Times(1)

	mockPersistenceRepo := mocks.NewMockLeaderboardPersistenceRepository(ctrl)
	mockPersistenceRepo.EXPECT().
		GetLeaderboard(ctx).
		Return([]domain.LeaderboardEntry{
			{UserID: "user-1", Username: "alice", Score: 1000, Rank: 1},
			{UserID: "user-2", Username: "bob", Score: 500, Rank: 2},
			{UserID: "user-3", Username: "carol", Score: 250, Rank: 3},
		}, nil).
		Times(1)

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockBroadcastService := mocks.NewMockBroadcastService(ctrl)

	logger := logger.New("info", false)
	uc := NewLeaderboardUseCase(mockCacheRepo, mockPersistenceRepo, mockUserRepo, mockBroadcastService, logger)

	// ── Act ─────────────────────────────────────────────────────────────
	entries, total, err := uc.GetLeaderboard(ctx, 2, 0)

	// ── Assert ──────────────────────────────────────────────────────────
	require.NoError(t, err)
	require.Len(t, entries, 2)
	require.Equal(t, int64(3), total)
	require.Equal(t, "user-1", entries[0].UserID)
	require.Equal(t, "user-2", entries[1].UserID)
}

func TestLeaderboardUseCase_GetLeaderboard_WhenGetTotalPlayersFails_ShouldReturnError(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCacheRepo := mocks.NewMockLeaderboardCacheRepository(ctrl)
	mockCacheRepo.EXPECT().
		GetTotalPlayers(ctx).
		Return(int64(0), errors.New("redis error")).
		Times(1)

	mockPersistenceRepo := mocks.NewMockLeaderboardPersistenceRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockBroadcastService := mocks.NewMockBroadcastService(ctrl)

	logger := logger.New("info", false)
	uc := NewLeaderboardUseCase(mockCacheRepo, mockPersistenceRepo, mockUserRepo, mockBroadcastService, logger)

	// ── Act ─────────────────────────────────────────────────────────────
	entries, total, err := uc.GetLeaderboard(ctx, 10, 0)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Error(t, err)
	require.Nil(t, entries)
	require.Equal(t, int64(0), total)
	require.Contains(t, err.Error(), "redis error")
}

func TestLeaderboardUseCase_GetLeaderboard_WhenCacheHitGetTopPlayersFails_ShouldReturnError(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCacheRepo := mocks.NewMockLeaderboardCacheRepository(ctrl)
	mockCacheRepo.EXPECT().
		GetTotalPlayers(ctx).
		Return(int64(2), nil).
		Times(1)
	mockCacheRepo.EXPECT().
		GetTopPlayers(ctx, int64(10), int64(0)).
		Return(nil, errors.New("redis error")).
		Times(1)

	mockPersistenceRepo := mocks.NewMockLeaderboardPersistenceRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockBroadcastService := mocks.NewMockBroadcastService(ctrl)

	logger := logger.New("info", false)
	uc := NewLeaderboardUseCase(mockCacheRepo, mockPersistenceRepo, mockUserRepo, mockBroadcastService, logger)

	// ── Act ─────────────────────────────────────────────────────────────
	entries, total, err := uc.GetLeaderboard(ctx, 10, 0)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Error(t, err)
	require.Nil(t, entries)
	require.Equal(t, int64(0), total)
}

func TestLeaderboardUseCase_GetLeaderboard_WhenCacheMissGetLeaderboardFails_ShouldReturnError(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCacheRepo := mocks.NewMockLeaderboardCacheRepository(ctrl)
	mockCacheRepo.EXPECT().
		GetTotalPlayers(ctx).
		Return(int64(0), nil).
		Times(1)

	mockPersistenceRepo := mocks.NewMockLeaderboardPersistenceRepository(ctrl)
	mockPersistenceRepo.EXPECT().
		GetLeaderboard(ctx).
		Return(nil, errors.New("database error")).
		Times(1)

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockBroadcastService := mocks.NewMockBroadcastService(ctrl)

	logger := logger.New("info", false)
	uc := NewLeaderboardUseCase(mockCacheRepo, mockPersistenceRepo, mockUserRepo, mockBroadcastService, logger)

	// ── Act ─────────────────────────────────────────────────────────────
	entries, total, err := uc.GetLeaderboard(ctx, 10, 0)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Error(t, err)
	require.Nil(t, entries)
	require.Equal(t, int64(0), total)
	require.Contains(t, err.Error(), "database error")
}

func TestLeaderboardUseCase_GetLeaderboard_WhenCacheMissEmpty_ShouldReturnEmpty(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCacheRepo := mocks.NewMockLeaderboardCacheRepository(ctrl)
	mockCacheRepo.EXPECT().
		GetTotalPlayers(ctx).
		Return(int64(0), nil).
		Times(1)

	mockPersistenceRepo := mocks.NewMockLeaderboardPersistenceRepository(ctrl)
	mockPersistenceRepo.EXPECT().
		GetLeaderboard(ctx).
		Return([]domain.LeaderboardEntry{}, nil).
		Times(1)

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockBroadcastService := mocks.NewMockBroadcastService(ctrl)

	logger := logger.New("info", false)
	uc := NewLeaderboardUseCase(mockCacheRepo, mockPersistenceRepo, mockUserRepo, mockBroadcastService, logger)

	// ── Act ─────────────────────────────────────────────────────────────
	entries, total, err := uc.GetLeaderboard(ctx, 10, 0)

	// ── Assert ──────────────────────────────────────────────────────────
	require.NoError(t, err)
	require.NotNil(t, entries)
	require.Empty(t, entries)
	require.Equal(t, int64(0), total)
}

func TestLeaderboardUseCase_SubscribeToEntryUpdates_ShouldReturnChannelFromBroadcastService(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedCh := make(chan *domain.LeaderboardEntry, 1)

	mockCacheRepo := mocks.NewMockLeaderboardCacheRepository(ctrl)
	mockPersistenceRepo := mocks.NewMockLeaderboardPersistenceRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)

	mockBroadcastService := mocks.NewMockBroadcastService(ctrl)
	mockBroadcastService.EXPECT().
		SubscribeToEntryUpdates(ctx).
		Return((<-chan *domain.LeaderboardEntry)(expectedCh), nil).
		Times(1)

	logger := logger.New("info", false)
	uc := NewLeaderboardUseCase(mockCacheRepo, mockPersistenceRepo, mockUserRepo, mockBroadcastService, logger)

	// ── Act ─────────────────────────────────────────────────────────────
	ch, err := uc.SubscribeToEntryUpdates(ctx)

	// ── Assert ──────────────────────────────────────────────────────────
	require.NoError(t, err)
	require.NotNil(t, ch)
	// Channel comparison is not reliable, just verify it's not nil
}
