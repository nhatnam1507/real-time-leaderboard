package application

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	"real-time-leaderboard/internal/module/auth/domain"
	"real-time-leaderboard/internal/module/auth/mocks"
	"real-time-leaderboard/internal/shared/logger"
	"real-time-leaderboard/internal/shared/response"
)

func TestAuthUseCase_Register_WhenValidRequest_ShouldCreateUserAndReturnTokens(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().
		GetByUsername(ctx, "alice").
		Return(nil, nil).
		Times(1)
	mockUserRepo.EXPECT().
		GetByEmail(ctx, "alice@example.com").
		Return(nil, nil).
		Times(1)
	mockUserRepo.EXPECT().
		Create(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, user *domain.User) error {
			user.ID = "user-123"
			return nil
		}).
		Times(1)

	mockJWT := mocks.NewMockJWTManager(ctrl)
	mockJWT.EXPECT().
		GenerateTokenPair("user-123").
		Return(&domain.TokenPair{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			ExpiresIn:    3600,
		}, nil).
		Times(1)

	logger := logger.New("info", false)
	uc := NewAuthUseCase(mockUserRepo, mockJWT, logger)

	req := RegisterRequest{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "secure123",
	}

	// ── Act ─────────────────────────────────────────────────────────────
	user, tokenPair, err := uc.Register(ctx, req)

	// ── Assert ──────────────────────────────────────────────────────────
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, "alice", user.Username)
	require.Equal(t, "alice@example.com", user.Email)
	require.NotEmpty(t, user.Password) // Password should be hashed
	require.NotNil(t, tokenPair)
	require.Equal(t, "access-token", tokenPair.AccessToken)
	require.Equal(t, "refresh-token", tokenPair.RefreshToken)
}

func TestAuthUseCase_Register_WhenUsernameExists_ShouldReturnConflictError(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().
		GetByUsername(ctx, "alice").
		Return(&domain.User{Username: "alice"}, nil).
		Times(1)

	mockJWT := mocks.NewMockJWTManager(ctrl)
	logger := logger.New("info", false)
	uc := NewAuthUseCase(mockUserRepo, mockJWT, logger)

	req := RegisterRequest{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "secure123",
	}

	// ── Act ─────────────────────────────────────────────────────────────
	user, tokenPair, err := uc.Register(ctx, req)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Error(t, err)
	require.Nil(t, user)
	require.Nil(t, tokenPair)

		var appErr *response.APIError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, response.CodeConflict, appErr.Code)
	require.Contains(t, appErr.Message, "Username already exists")
}

func TestAuthUseCase_Register_WhenEmailExists_ShouldReturnConflictError(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().
		GetByUsername(ctx, "alice").
		Return(nil, nil).
		Times(1)
	mockUserRepo.EXPECT().
		GetByEmail(ctx, "alice@example.com").
		Return(&domain.User{Email: "alice@example.com"}, nil).
		Times(1)

	mockJWT := mocks.NewMockJWTManager(ctrl)
	logger := logger.New("info", false)
	uc := NewAuthUseCase(mockUserRepo, mockJWT, logger)

	req := RegisterRequest{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "secure123",
	}

	// ── Act ─────────────────────────────────────────────────────────────
	user, tokenPair, err := uc.Register(ctx, req)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Error(t, err)
	require.Nil(t, user)
	require.Nil(t, tokenPair)

		var appErr *response.APIError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, response.CodeConflict, appErr.Code)
	require.Contains(t, appErr.Message, "Email already exists")
}

func TestAuthUseCase_Register_WhenGetByUsernameFails_ShouldReturnInternalError(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().
		GetByUsername(ctx, "alice").
		Return(nil, errors.New("database error")).
		Times(1)

	mockJWT := mocks.NewMockJWTManager(ctrl)
	logger := logger.New("info", false)
	uc := NewAuthUseCase(mockUserRepo, mockJWT, logger)

	req := RegisterRequest{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "secure123",
	}

	// ── Act ─────────────────────────────────────────────────────────────
	user, tokenPair, err := uc.Register(ctx, req)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Error(t, err)
	require.Nil(t, user)
	require.Nil(t, tokenPair)

		var appErr *response.APIError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, response.CodeInternal, appErr.Code)
}

func TestAuthUseCase_Register_WhenCreateFails_ShouldReturnInternalError(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().
		GetByUsername(ctx, "alice").
		Return(nil, nil).
		Times(1)
	mockUserRepo.EXPECT().
		GetByEmail(ctx, "alice@example.com").
		Return(nil, nil).
		Times(1)
	mockUserRepo.EXPECT().
		Create(ctx, gomock.Any()).
		Return(errors.New("database error")).
		Times(1)

	mockJWT := mocks.NewMockJWTManager(ctrl)
	logger := logger.New("info", false)
	uc := NewAuthUseCase(mockUserRepo, mockJWT, logger)

	req := RegisterRequest{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "secure123",
	}

	// ── Act ─────────────────────────────────────────────────────────────
	user, tokenPair, err := uc.Register(ctx, req)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Error(t, err)
	require.Nil(t, user)
	require.Nil(t, tokenPair)

		var appErr *response.APIError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, response.CodeInternal, appErr.Code)
}

func TestAuthUseCase_Register_WhenGenerateTokenFails_ShouldReturnInternalError(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().
		GetByUsername(ctx, "alice").
		Return(nil, nil).
		Times(1)
	mockUserRepo.EXPECT().
		GetByEmail(ctx, "alice@example.com").
		Return(nil, nil).
		Times(1)
	mockUserRepo.EXPECT().
		Create(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, user *domain.User) error {
			user.ID = "user-123"
			return nil
		}).
		Times(1)

	mockJWT := mocks.NewMockJWTManager(ctrl)
	mockJWT.EXPECT().
		GenerateTokenPair("user-123").
		Return(nil, errors.New("token generation error")).
		Times(1)

	logger := logger.New("info", false)
	uc := NewAuthUseCase(mockUserRepo, mockJWT, logger)

	req := RegisterRequest{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "secure123",
	}

	// ── Act ─────────────────────────────────────────────────────────────
	user, tokenPair, err := uc.Register(ctx, req)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Error(t, err)
	require.Nil(t, user)
	require.Nil(t, tokenPair)

		var appErr *response.APIError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, response.CodeInternal, appErr.Code)
}

func TestAuthUseCase_Login_WhenValidCredentials_ShouldReturnUserAndTokens(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("secure123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().
		GetByUsername(ctx, "alice").
		Return(&domain.User{
			ID:       "user-123",
			Username: "alice",
			Email:    "alice@example.com",
			Password: string(hashedPassword),
		}, nil).
		Times(1)

	mockJWT := mocks.NewMockJWTManager(ctrl)
	mockJWT.EXPECT().
		GenerateTokenPair("user-123").
		Return(&domain.TokenPair{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			ExpiresIn:    3600,
		}, nil).
		Times(1)

	logger := logger.New("info", false)
	uc := NewAuthUseCase(mockUserRepo, mockJWT, logger)

	req := LoginRequest{
		Username: "alice",
		Password: "secure123",
	}

	// ── Act ─────────────────────────────────────────────────────────────
	user, tokenPair, err := uc.Login(ctx, req)

	// ── Assert ──────────────────────────────────────────────────────────
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, "alice", user.Username)
	require.NotNil(t, tokenPair)
	require.Equal(t, "access-token", tokenPair.AccessToken)
}

func TestAuthUseCase_Login_WhenUserNotFound_ShouldReturnUnauthorizedError(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().
		GetByUsername(ctx, "unknown").
		Return(nil, nil).
		Times(1)

	mockJWT := mocks.NewMockJWTManager(ctrl)
	logger := logger.New("info", false)
	uc := NewAuthUseCase(mockUserRepo, mockJWT, logger)

	req := LoginRequest{
		Username: "unknown",
		Password: "secure123",
	}

	// ── Act ─────────────────────────────────────────────────────────────
	user, tokenPair, err := uc.Login(ctx, req)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Error(t, err)
	require.Nil(t, user)
	require.Nil(t, tokenPair)

		var appErr *response.APIError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, response.CodeUnauthorized, appErr.Code)
	require.Contains(t, appErr.Message, "Invalid credentials")
}

func TestAuthUseCase_Login_WhenWrongPassword_ShouldReturnUnauthorizedError(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	require.NoError(t, err)

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().
		GetByUsername(ctx, "alice").
		Return(&domain.User{
			ID:       "user-123",
			Username: "alice",
			Password: string(hashedPassword),
		}, nil).
		Times(1)

	mockJWT := mocks.NewMockJWTManager(ctrl)
	logger := logger.New("info", false)
	uc := NewAuthUseCase(mockUserRepo, mockJWT, logger)

	req := LoginRequest{
		Username: "alice",
		Password: "wrong-password",
	}

	// ── Act ─────────────────────────────────────────────────────────────
	user, tokenPair, err := uc.Login(ctx, req)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Error(t, err)
	require.Nil(t, user)
	require.Nil(t, tokenPair)

		var appErr *response.APIError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, response.CodeUnauthorized, appErr.Code)
	require.Contains(t, appErr.Message, "Invalid credentials")
}

func TestAuthUseCase_Login_WhenGetByUsernameFails_ShouldReturnInternalError(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().
		GetByUsername(ctx, "alice").
		Return(nil, errors.New("database error")).
		Times(1)

	mockJWT := mocks.NewMockJWTManager(ctrl)
	logger := logger.New("info", false)
	uc := NewAuthUseCase(mockUserRepo, mockJWT, logger)

	req := LoginRequest{
		Username: "alice",
		Password: "secure123",
	}

	// ── Act ─────────────────────────────────────────────────────────────
	user, tokenPair, err := uc.Login(ctx, req)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Error(t, err)
	require.Nil(t, user)
	require.Nil(t, tokenPair)

		var appErr *response.APIError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, response.CodeInternal, appErr.Code)
}

func TestAuthUseCase_ValidateToken_WhenValidToken_ShouldReturnUserID(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockJWT := mocks.NewMockJWTManager(ctrl)
	mockJWT.EXPECT().
		ValidateToken("valid-token").
		Return("user-123", nil).
		Times(1)

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().
		GetByID(ctx, "user-123").
		Return(&domain.User{ID: "user-123"}, nil).
		Times(1)

	logger := logger.New("info", false)
	uc := NewAuthUseCase(mockUserRepo, mockJWT, logger)

	// ── Act ─────────────────────────────────────────────────────────────
	userID, err := uc.ValidateToken(ctx, "valid-token")

	// ── Assert ──────────────────────────────────────────────────────────
	require.NoError(t, err)
	require.Equal(t, "user-123", userID)
}

func TestAuthUseCase_ValidateToken_WhenInvalidToken_ShouldReturnUnauthorizedError(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockJWT := mocks.NewMockJWTManager(ctrl)
	mockJWT.EXPECT().
		ValidateToken("invalid-token").
		Return("", errors.New("invalid token")).
		Times(1)

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	logger := logger.New("info", false)
	uc := NewAuthUseCase(mockUserRepo, mockJWT, logger)

	// ── Act ─────────────────────────────────────────────────────────────
	userID, err := uc.ValidateToken(ctx, "invalid-token")

	// ── Assert ──────────────────────────────────────────────────────────
	require.Error(t, err)
	require.Empty(t, userID)

		var appErr *response.APIError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, response.CodeUnauthorized, appErr.Code)
	require.Contains(t, appErr.Message, "Invalid or expired token")
}

func TestAuthUseCase_ValidateToken_WhenUserNotFound_ShouldReturnUnauthorizedError(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockJWT := mocks.NewMockJWTManager(ctrl)
	mockJWT.EXPECT().
		ValidateToken("valid-token").
		Return("user-123", nil).
		Times(1)

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().
		GetByID(ctx, "user-123").
		Return(nil, nil).
		Times(1)

	logger := logger.New("info", false)
	uc := NewAuthUseCase(mockUserRepo, mockJWT, logger)

	// ── Act ─────────────────────────────────────────────────────────────
	userID, err := uc.ValidateToken(ctx, "valid-token")

	// ── Assert ──────────────────────────────────────────────────────────
	require.Error(t, err)
	require.Empty(t, userID)

		var appErr *response.APIError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, response.CodeUnauthorized, appErr.Code)
	require.Contains(t, appErr.Message, "User not found")
}

func TestAuthUseCase_RefreshToken_WhenValidRefreshToken_ShouldReturnNewTokenPair(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockJWT := mocks.NewMockJWTManager(ctrl)
	mockJWT.EXPECT().
		ValidateToken("refresh-token").
		Return("user-123", nil).
		Times(1)
	mockJWT.EXPECT().
		GenerateTokenPair("user-123").
		Return(&domain.TokenPair{
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
			ExpiresIn:    3600,
		}, nil).
		Times(1)

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().
		GetByID(ctx, "user-123").
		Return(&domain.User{ID: "user-123"}, nil).
		Times(1)

	logger := logger.New("info", false)
	uc := NewAuthUseCase(mockUserRepo, mockJWT, logger)

	// ── Act ─────────────────────────────────────────────────────────────
	tokenPair, err := uc.RefreshToken(ctx, "refresh-token")

	// ── Assert ──────────────────────────────────────────────────────────
	require.NoError(t, err)
	require.NotNil(t, tokenPair)
	require.Equal(t, "new-access-token", tokenPair.AccessToken)
	require.Equal(t, "new-refresh-token", tokenPair.RefreshToken)
}

func TestAuthUseCase_RefreshToken_WhenInvalidRefreshToken_ShouldReturnUnauthorizedError(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockJWT := mocks.NewMockJWTManager(ctrl)
	mockJWT.EXPECT().
		ValidateToken("invalid-refresh-token").
		Return("", errors.New("invalid token")).
		Times(1)

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	logger := logger.New("info", false)
	uc := NewAuthUseCase(mockUserRepo, mockJWT, logger)

	// ── Act ─────────────────────────────────────────────────────────────
	tokenPair, err := uc.RefreshToken(ctx, "invalid-refresh-token")

	// ── Assert ──────────────────────────────────────────────────────────
	require.Error(t, err)
	require.Nil(t, tokenPair)

		var appErr *response.APIError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, response.CodeUnauthorized, appErr.Code)
	require.Contains(t, appErr.Message, "Invalid or expired refresh token")
}

func TestAuthUseCase_RefreshToken_WhenUserNotFound_ShouldReturnUnauthorizedError(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockJWT := mocks.NewMockJWTManager(ctrl)
	mockJWT.EXPECT().
		ValidateToken("refresh-token").
		Return("user-123", nil).
		Times(1)

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().
		GetByID(ctx, "user-123").
		Return(nil, nil).
		Times(1)

	logger := logger.New("info", false)
	uc := NewAuthUseCase(mockUserRepo, mockJWT, logger)

	// ── Act ─────────────────────────────────────────────────────────────
	tokenPair, err := uc.RefreshToken(ctx, "refresh-token")

	// ── Assert ──────────────────────────────────────────────────────────
	require.Error(t, err)
	require.Nil(t, tokenPair)

		var appErr *response.APIError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, response.CodeUnauthorized, appErr.Code)
	require.Contains(t, appErr.Message, "User not found")
}
