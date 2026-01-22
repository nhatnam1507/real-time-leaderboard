package application

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"real-time-leaderboard/internal/module/leaderboard/mocks"
	"real-time-leaderboard/internal/module/leaderboard/domain"
	"real-time-leaderboard/internal/shared/logger"
	"real-time-leaderboard/internal/shared/response"
)

func TestLeaderboardUseCase_SyncFromPostgres_WhenRedisEmpty_ShouldSyncAllEntries(t *testing.T) {
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

	mockPersistenceRepo := mocks.NewMockLeaderboardPersistenceRepository(ctrl)
	mockPersistenceRepo.EXPECT().
		GetLeaderboard(ctx).
		Return([]domain.LeaderboardEntry{
			{UserID: "user-1", Score: 1000},
			{UserID: "user-2", Score: 500},
		}, nil).
		Times(1)

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockBroadcastService := mocks.NewMockBroadcastService(ctrl)

	logger := logger.New("info", false)
	uc := NewLeaderboardUseCase(mockCacheRepo, mockPersistenceRepo, mockUserRepo, mockBroadcastService, logger)

	// ── Act ─────────────────────────────────────────────────────────────
	err := uc.SyncFromPostgres(ctx)

	// ── Assert ──────────────────────────────────────────────────────────
	require.NoError(t, err)
}

func TestLeaderboardUseCase_SyncFromPostgres_WhenRedisNotEmpty_ShouldNotSync(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCacheRepo := mocks.NewMockLeaderboardCacheRepository(ctrl)
	mockCacheRepo.EXPECT().
		GetTotalPlayers(ctx).
		Return(int64(5), nil).
		Times(1)

	mockPersistenceRepo := mocks.NewMockLeaderboardPersistenceRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockBroadcastService := mocks.NewMockBroadcastService(ctrl)

	logger := logger.New("info", false)
	uc := NewLeaderboardUseCase(mockCacheRepo, mockPersistenceRepo, mockUserRepo, mockBroadcastService, logger)

	// ── Act ─────────────────────────────────────────────────────────────
	err := uc.SyncFromPostgres(ctx)

	// ── Assert ──────────────────────────────────────────────────────────
	require.NoError(t, err)
}

func TestLeaderboardUseCase_SyncFromPostgres_WhenGetTotalPlayersFails_ShouldReturnError(t *testing.T) {
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
	err := uc.SyncFromPostgres(ctx)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Error(t, err)
	require.Contains(t, err.Error(), "redis error")
}

func TestLeaderboardUseCase_SyncFromPostgres_WhenGetLeaderboardFails_ShouldReturnError(t *testing.T) {
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
	err := uc.SyncFromPostgres(ctx)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Error(t, err)
	require.Contains(t, err.Error(), "database error")
}

func TestLeaderboardUseCase_GetFullLeaderboard_WhenValidRequest_ShouldReturnEnrichedLeaderboard(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCacheRepo := mocks.NewMockLeaderboardCacheRepository(ctrl)
	mockCacheRepo.EXPECT().
		GetTopPlayers(ctx, int64(1000), int64(0)).
		Return([]domain.LeaderboardEntry{
			{UserID: "user-1", Score: 1000, Rank: 1},
			{UserID: "user-2", Score: 500, Rank: 2},
		}, nil).
		Times(1)
	mockCacheRepo.EXPECT().
		GetTotalPlayers(ctx).
		Return(int64(2), nil).
		Times(1)

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().
		GetByIDs(ctx, []string{"user-1", "user-2"}).
		Return(map[string]string{
			"user-1": "alice",
			"user-2": "bob",
		}, nil).
		Times(1)

	mockPersistenceRepo := mocks.NewMockLeaderboardPersistenceRepository(ctrl)
	mockBroadcastService := mocks.NewMockBroadcastService(ctrl)

	logger := logger.New("info", false)
	uc := NewLeaderboardUseCase(mockCacheRepo, mockPersistenceRepo, mockUserRepo, mockBroadcastService, logger)

	// ── Act ─────────────────────────────────────────────────────────────
	entries, total, err := uc.GetFullLeaderboard(ctx)

	// ── Assert ──────────────────────────────────────────────────────────
	require.NoError(t, err)
	require.NotNil(t, entries)
	require.Equal(t, int64(2), total)
	require.Len(t, entries, 2)
	require.Equal(t, "alice", entries[0].Username)
	require.Equal(t, "bob", entries[1].Username)
}

func TestLeaderboardUseCase_GetFullLeaderboard_WhenGetTopPlayersFails_ShouldReturnInternalError(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCacheRepo := mocks.NewMockLeaderboardCacheRepository(ctrl)
	mockCacheRepo.EXPECT().
		GetTopPlayers(ctx, int64(1000), int64(0)).
		Return(nil, errors.New("redis error")).
		Times(1)

	mockPersistenceRepo := mocks.NewMockLeaderboardPersistenceRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockBroadcastService := mocks.NewMockBroadcastService(ctrl)

	logger := logger.New("info", false)
	uc := NewLeaderboardUseCase(mockCacheRepo, mockPersistenceRepo, mockUserRepo, mockBroadcastService, logger)

	// ── Act ─────────────────────────────────────────────────────────────
	entries, total, err := uc.GetFullLeaderboard(ctx)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Error(t, err)
	require.Nil(t, entries)
	require.Equal(t, int64(0), total)

		var appErr *response.APIError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, response.CodeInternal, appErr.Code)
	require.Contains(t, appErr.Message, "Failed to retrieve leaderboard")
}

func TestLeaderboardUseCase_GetFullLeaderboard_WhenGetTotalPlayersFails_ShouldUseEntriesLength(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCacheRepo := mocks.NewMockLeaderboardCacheRepository(ctrl)
	mockCacheRepo.EXPECT().
		GetTopPlayers(ctx, int64(1000), int64(0)).
		Return([]domain.LeaderboardEntry{
			{UserID: "user-1", Score: 1000, Rank: 1},
		}, nil).
		Times(1)
	mockCacheRepo.EXPECT().
		GetTotalPlayers(ctx).
		Return(int64(0), errors.New("redis error")).
		Times(1)

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().
		GetByIDs(ctx, []string{"user-1"}).
		Return(map[string]string{"user-1": "alice"}, nil).
		Times(1)

	mockPersistenceRepo := mocks.NewMockLeaderboardPersistenceRepository(ctrl)
	mockBroadcastService := mocks.NewMockBroadcastService(ctrl)

	logger := logger.New("info", false)
	uc := NewLeaderboardUseCase(mockCacheRepo, mockPersistenceRepo, mockUserRepo, mockBroadcastService, logger)

	// ── Act ─────────────────────────────────────────────────────────────
	entries, total, err := uc.GetFullLeaderboard(ctx)

	// ── Assert ──────────────────────────────────────────────────────────
	require.NoError(t, err)
	require.NotNil(t, entries)
	require.Equal(t, int64(1), total) // Should use entries length as fallback
	require.Len(t, entries, 1)
}

func TestLeaderboardUseCase_GetFullLeaderboard_WhenEnrichmentFails_ShouldStillReturnLeaderboard(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCacheRepo := mocks.NewMockLeaderboardCacheRepository(ctrl)
	mockCacheRepo.EXPECT().
		GetTopPlayers(ctx, int64(1000), int64(0)).
		Return([]domain.LeaderboardEntry{
			{UserID: "user-1", Score: 1000, Rank: 1},
		}, nil).
		Times(1)
	mockCacheRepo.EXPECT().
		GetTotalPlayers(ctx).
		Return(int64(1), nil).
		Times(1)

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().
		GetByIDs(ctx, []string{"user-1"}).
		Return(nil, errors.New("database error")).
		Times(1)

	mockPersistenceRepo := mocks.NewMockLeaderboardPersistenceRepository(ctrl)
	mockBroadcastService := mocks.NewMockBroadcastService(ctrl)

	logger := logger.New("info", false)
	uc := NewLeaderboardUseCase(mockCacheRepo, mockPersistenceRepo, mockUserRepo, mockBroadcastService, logger)

	// ── Act ─────────────────────────────────────────────────────────────
	entries, total, err := uc.GetFullLeaderboard(ctx)

	// ── Assert ──────────────────────────────────────────────────────────
	require.NoError(t, err) // Enrichment failure is non-critical
	require.NotNil(t, entries)
	require.Equal(t, int64(1), total)
	require.Len(t, entries, 1)
	require.Empty(t, entries[0].Username) // Username not enriched
}

func TestLeaderboardUseCase_GetFullLeaderboard_WhenEmptyLeaderboard_ShouldReturnEmptyLeaderboard(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCacheRepo := mocks.NewMockLeaderboardCacheRepository(ctrl)
	mockCacheRepo.EXPECT().
		GetTopPlayers(ctx, int64(1000), int64(0)).
		Return([]domain.LeaderboardEntry{}, nil).
		Times(1)
	mockCacheRepo.EXPECT().
		GetTotalPlayers(ctx).
		Return(int64(0), nil).
		Times(1)

	mockPersistenceRepo := mocks.NewMockLeaderboardPersistenceRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockBroadcastService := mocks.NewMockBroadcastService(ctrl)

	logger := logger.New("info", false)
	uc := NewLeaderboardUseCase(mockCacheRepo, mockPersistenceRepo, mockUserRepo, mockBroadcastService, logger)

	// ── Act ─────────────────────────────────────────────────────────────
	entries, total, err := uc.GetFullLeaderboard(ctx)

	// ── Assert ──────────────────────────────────────────────────────────
	require.NoError(t, err)
	require.NotNil(t, entries)
	require.Equal(t, int64(0), total)
	require.Empty(t, entries)
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
