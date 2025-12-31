package http

import (
	"github.com/gin-gonic/gin"
	"real-time-leaderboard/internal/module/auth/application"
	"real-time-leaderboard/internal/shared/response"
	"real-time-leaderboard/internal/shared/validator"
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
// @Description Register a new user with username, email, and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body application.RegisterRequest true "Registration request"
// @Success 201 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
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
// @Description Authenticate user and return access token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body application.LoginRequest true "Login request"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
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
// @Description Refresh access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param refresh_token body object{refresh_token=string} true "Refresh token"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
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

