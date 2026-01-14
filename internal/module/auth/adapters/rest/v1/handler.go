// Package v1 provides REST API v1 handlers for the auth module.
package v1

import (
	"real-time-leaderboard/internal/module/auth/application"
	"real-time-leaderboard/internal/shared/middleware"
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

// GetCurrentUser returns the current authenticated user's information
func (h *Handler) GetCurrentUser(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, response.NewUnauthorizedError("User ID not found in context"))
		return
	}

	user, err := h.authUseCase.GetCurrentUser(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, user, "User retrieved successfully")
}

// RegisterPublicRoutes registers public auth routes (no auth required)
func (h *Handler) RegisterPublicRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.RefreshToken)
	}
}

// RegisterProtectedRoutes registers protected auth routes (requires authentication)
func (h *Handler) RegisterProtectedRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.GET("/me", h.GetCurrentUser)
	}
}
