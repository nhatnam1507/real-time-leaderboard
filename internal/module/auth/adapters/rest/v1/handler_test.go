// Package v1 provides REST API v1 handlers for the auth module.
package v1

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"real-time-leaderboard/internal/module/auth/application"
	"real-time-leaderboard/internal/module/auth/domain"
	authmocks "real-time-leaderboard/internal/module/auth/adapters/mocks"
	"real-time-leaderboard/internal/shared/logger"
	"real-time-leaderboard/internal/shared/response"
)

func TestHandler_Register_WhenValidJSON_ShouldReturn201AndUserAndToken(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := authmocks.NewMockAuthUseCase(ctrl)
	req := application.RegisterRequest{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "secure123",
	}
	mockAuth.EXPECT().
		Register(gomock.Any(), req).
		Return(
			&domain.User{ID: "user-1", Username: "alice", Email: "alice@example.com"},
			&domain.TokenPair{AccessToken: "at", RefreshToken: "rt", ExpiresIn: 3600},
			nil,
		).
		Times(1)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{"username":"alice","email":"alice@example.com","password":"secure123"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h := NewHandler(mockAuth, logger.New("info", false))

	// ── Act ─────────────────────────────────────────────────────────────
	h.Register(c)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Equal(t, http.StatusCreated, w.Code)
	var body response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.True(t, body.Success)
	require.Equal(t, "User registered successfully", body.Message)
	require.NotNil(t, body.Data)
}

func TestHandler_Register_WhenMissingRequiredFields_ShouldReturn400(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := authmocks.NewMockAuthUseCase(ctrl)
	mockAuth.EXPECT().Register(gomock.Any(), gomock.Any()).Times(0)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h := NewHandler(mockAuth, logger.New("info", false))

	// ── Act ─────────────────────────────────────────────────────────────
	h.Register(c)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Equal(t, http.StatusBadRequest, w.Code)
	var body response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.False(t, body.Success)
	require.NotNil(t, body.Error)
	require.Equal(t, string(response.CodeValidation), body.Error.Code)
}

func TestHandler_Register_WhenUseCaseReturnsConflict_ShouldReturn409(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := authmocks.NewMockAuthUseCase(ctrl)
	req := application.RegisterRequest{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "secure123",
	}
	mockAuth.EXPECT().
		Register(gomock.Any(), req).
		Return(nil, nil, domain.ErrUserAlreadyExists).
		Times(1)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{"username":"alice","email":"alice@example.com","password":"secure123"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h := NewHandler(mockAuth, logger.New("info", false))

	// ── Act ─────────────────────────────────────────────────────────────
	h.Register(c)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Equal(t, http.StatusConflict, w.Code)
	var body response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.False(t, body.Success)
	require.NotNil(t, body.Error)
	require.Equal(t, string(response.CodeConflict), body.Error.Code)
}

func TestHandler_Login_WhenValidJSON_ShouldReturn200AndUserAndToken(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := authmocks.NewMockAuthUseCase(ctrl)
	req := application.LoginRequest{Username: "alice", Password: "secure123"}
	mockAuth.EXPECT().
		Login(gomock.Any(), req).
		Return(
			&domain.User{ID: "user-1", Username: "alice", Email: "alice@example.com"},
			&domain.TokenPair{AccessToken: "at", RefreshToken: "rt", ExpiresIn: 3600},
			nil,
		).
		Times(1)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"username":"alice","password":"secure123"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h := NewHandler(mockAuth, logger.New("info", false))

	// ── Act ─────────────────────────────────────────────────────────────
	h.Login(c)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Equal(t, http.StatusOK, w.Code)
	var body response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.True(t, body.Success)
	require.Equal(t, "Login successful", body.Message)
}

func TestHandler_Login_WhenUseCaseReturnsUnauthorized_ShouldReturn401(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := authmocks.NewMockAuthUseCase(ctrl)
	req := application.LoginRequest{Username: "alice", Password: "wrong"}
	mockAuth.EXPECT().
		Login(gomock.Any(), req).
		Return(nil, nil, domain.ErrInvalidCredentials).
		Times(1)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"username":"alice","password":"wrong"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h := NewHandler(mockAuth, logger.New("info", false))

	// ── Act ─────────────────────────────────────────────────────────────
	h.Login(c)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Equal(t, http.StatusUnauthorized, w.Code)
	var body response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.False(t, body.Success)
	require.Equal(t, string(response.CodeUnauthorized), body.Error.Code)
}

func TestHandler_RefreshToken_WhenValidJSON_ShouldReturn200AndToken(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := authmocks.NewMockAuthUseCase(ctrl)
	mockAuth.EXPECT().
		RefreshToken(gomock.Any(), "refresh-token").
		Return(&domain.TokenPair{AccessToken: "at2", RefreshToken: "rt2", ExpiresIn: 3600}, nil).
		Times(1)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBufferString(`{"refresh_token":"refresh-token"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h := NewHandler(mockAuth, logger.New("info", false))

	// ── Act ─────────────────────────────────────────────────────────────
	h.RefreshToken(c)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Equal(t, http.StatusOK, w.Code)
	var body response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.True(t, body.Success)
	require.Equal(t, "Token refreshed successfully", body.Message)
}

func TestHandler_RefreshToken_WhenMissingRefreshToken_ShouldReturn400(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := authmocks.NewMockAuthUseCase(ctrl)
	mockAuth.EXPECT().RefreshToken(gomock.Any(), gomock.Any()).Times(0)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBufferString(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h := NewHandler(mockAuth, logger.New("info", false))

	// ── Act ─────────────────────────────────────────────────────────────
	h.RefreshToken(c)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Equal(t, http.StatusBadRequest, w.Code)
	var body response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.False(t, body.Success)
	require.NotNil(t, body.Error)
}

func TestHandler_GetCurrentUser_WhenUserIDInContext_ShouldReturn200AndUser(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := authmocks.NewMockAuthUseCase(ctrl)
	mockAuth.EXPECT().
		GetCurrentUser(gomock.Any(), "user-123").
		Return(&domain.User{ID: "user-123", Username: "alice", Email: "alice@example.com"}, nil).
		Times(1)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	c.Set("user_id", "user-123")

	h := NewHandler(mockAuth, logger.New("info", false))

	// ── Act ─────────────────────────────────────────────────────────────
	h.GetCurrentUser(c)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Equal(t, http.StatusOK, w.Code)
	var body response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.True(t, body.Success)
	require.Equal(t, "User retrieved successfully", body.Message)
}

func TestHandler_GetCurrentUser_WhenUserIDNotInContext_ShouldReturn401(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := authmocks.NewMockAuthUseCase(ctrl)
	mockAuth.EXPECT().GetCurrentUser(gomock.Any(), gomock.Any()).Times(0)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	// do not set user_id

	h := NewHandler(mockAuth, logger.New("info", false))

	// ── Act ─────────────────────────────────────────────────────────────
	h.GetCurrentUser(c)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Equal(t, http.StatusUnauthorized, w.Code)
	var body response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.False(t, body.Success)
	require.Equal(t, string(response.CodeUnauthorized), body.Error.Code)
	require.Contains(t, body.Error.Message, "User ID not found")
}

func TestHandler_GetCurrentUser_WhenUseCaseReturnsError_ShouldReturnMappedError(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := authmocks.NewMockAuthUseCase(ctrl)
	mockAuth.EXPECT().
		GetCurrentUser(gomock.Any(), "user-123").
		Return(nil, domain.ErrUserNotFound).
		Times(1)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	c.Set("user_id", "user-123")

	h := NewHandler(mockAuth, logger.New("info", false))

	// ── Act ─────────────────────────────────────────────────────────────
	h.GetCurrentUser(c)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Equal(t, http.StatusNotFound, w.Code)
	var body response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.False(t, body.Success)
	require.Equal(t, string(response.CodeNotFound), body.Error.Code)
}
