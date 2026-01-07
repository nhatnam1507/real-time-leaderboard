// Package rest provides REST API handlers for the auth module.
package rest

import (
	"real-time-leaderboard/internal/module/auth/application"
	"real-time-leaderboard/internal/shared/response"
	"real-time-leaderboard/internal/shared/validator"

	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for authentication
type Handler struct {
	authUseCase *application.AuthUseCase
}

// NewHandler creates a new auth HTTP handler
func NewHandler(authUseCase *application.AuthUseCase) *Handler {
	return &Handler{
		authUseCase: authUseCase,
	}
}

// Register handles user registration
// @Summary Register a new user
// @Description Register a new user with username, email, and password. Returns user information and JWT tokens.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body application.RegisterRequest true "Registration request" example({"username":"john_doe","email":"john@example.com","password":"password123"})
// @Success 201 {object} response.Response "User registered successfully"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 409 {object} response.Response "User already exists"
// @Router /api/v1/auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req application.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, validator.Validate(req))
		return
	}

	if err := validator.Validate(req); err != nil {
		response.Error(c, err)
		return
	}

	user, tokenPair, err := h.authUseCase.Register(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, gin.H{
		"user":  user,
		"token": tokenPair,
	}, "User registered successfully")
}

// Login handles user login
// @Summary Login user
// @Description Authenticate user with username and password, returns JWT access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body application.LoginRequest true "Login request" example({"username":"john_doe","password":"password123"})
// @Success 200 {object} response.Response "Login successful"
// @Failure 401 {object} response.Response "Invalid credentials"
// @Router /api/v1/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req application.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, validator.Validate(req))
		return
	}

	if err := validator.Validate(req); err != nil {
		response.Error(c, err)
		return
	}

	user, tokenPair, err := h.authUseCase.Login(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"user":  user,
		"token": tokenPair,
	}, "Login successful")
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Refresh access token using a valid refresh token. Returns new access and refresh tokens.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{refresh_token=string} true "Refresh token request" example({"refresh_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."})
// @Success 200 {object} response.Response "Token refreshed successfully"
// @Failure 401 {object} response.Response "Invalid or expired refresh token"
// @Router /api/v1/auth/refresh [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, validator.Validate(req))
		return
	}

	if err := validator.Validate(req); err != nil {
		response.Error(c, err)
		return
	}

	tokenPair, err := h.authUseCase.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"token": tokenPair}, "Token refreshed successfully")
}

// RegisterRoutes registers auth routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.RefreshToken)
	}
}
