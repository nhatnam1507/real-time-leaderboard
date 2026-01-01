package application

import (
	"context"

	"golang.org/x/crypto/bcrypt"
	"real-time-leaderboard/internal/module/auth/domain"
	"real-time-leaderboard/internal/shared/errors"
	"real-time-leaderboard/internal/shared/logger"
)

// AuthUseCase handles authentication use cases
type AuthUseCase struct {
	userRepo domain.UserRepository
	jwtMgr   JWTManager
	logger   *logger.Logger
}

// JWTManager interface for JWT operations
type JWTManager interface {
	GenerateTokenPair(userID string) (*domain.TokenPair, error)
	ValidateToken(token string) (string, error)
}

// NewAuthUseCase creates a new auth use case
func NewAuthUseCase(userRepo domain.UserRepository, jwtMgr JWTManager, l *logger.Logger) *AuthUseCase {
	return &AuthUseCase{
		userRepo: userRepo,
		jwtMgr:   jwtMgr,
		logger:   l,
	}
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// Register registers a new user
func (uc *AuthUseCase) Register(ctx context.Context, req RegisterRequest) (*domain.User, *domain.TokenPair, error) {
	// Check if username already exists
	existingUser, err := uc.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		uc.logger.Errorf("Failed to check username existence: %v", err)
		return nil, nil, errors.NewInternalError("Registration failed", err)
	}
	if existingUser != nil {
		return nil, nil, errors.NewConflictError("Username already exists")
	}

	// Check if email already exists
	existingUser, err = uc.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		uc.logger.Errorf("Failed to check email existence: %v", err)
		return nil, nil, errors.NewInternalError("Registration failed", err)
	}
	if existingUser != nil {
		return nil, nil, errors.NewConflictError("Email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		uc.logger.Errorf("Failed to hash password: %v", err)
		return nil, nil, errors.NewInternalError("Registration failed", err)
	}

	// Create user
	user := &domain.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		uc.logger.Errorf("Failed to create user: %v", err)
		return nil, nil, errors.NewInternalError("Registration failed", err)
	}

	// Generate tokens
	tokenPair, err := uc.jwtMgr.GenerateTokenPair(user.ID)
	if err != nil {
		uc.logger.Errorf("Failed to generate tokens: %v", err)
		return nil, nil, errors.NewInternalError("Registration failed", err)
	}

	uc.logger.Infof("User registered: %s", user.ID)
	return user, tokenPair, nil
}

// Login authenticates a user
func (uc *AuthUseCase) Login(ctx context.Context, req LoginRequest) (*domain.User, *domain.TokenPair, error) {
	// Get user by username
	user, err := uc.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		uc.logger.Errorf("Failed to get user: %v", err)
		return nil, nil, errors.NewInternalError("Login failed", err)
	}
	if user == nil {
		return nil, nil, errors.NewUnauthorizedError("Invalid credentials")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, nil, errors.NewUnauthorizedError("Invalid credentials")
	}

	// Generate tokens
	tokenPair, err := uc.jwtMgr.GenerateTokenPair(user.ID)
	if err != nil {
		uc.logger.Errorf("Failed to generate tokens: %v", err)
		return nil, nil, errors.NewInternalError("Login failed", err)
	}

	uc.logger.Infof("User logged in: %s", user.ID)
	return user, tokenPair, nil
}

// ValidateToken validates a JWT token and returns the user ID
func (uc *AuthUseCase) ValidateToken(ctx context.Context, token string) (string, error) {
	userID, err := uc.jwtMgr.ValidateToken(token)
	if err != nil {
		return "", errors.NewUnauthorizedError("Invalid or expired token")
	}

	// Optionally verify user still exists
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		uc.logger.Errorf("Failed to get user: %v", err)
		return "", errors.NewInternalError("Token validation failed", err)
	}
	if user == nil {
		return "", errors.NewUnauthorizedError("User not found")
	}

	return userID, nil
}

// RefreshToken refreshes an access token using a refresh token
func (uc *AuthUseCase) RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	userID, err := uc.jwtMgr.ValidateToken(refreshToken)
	if err != nil {
		return nil, errors.NewUnauthorizedError("Invalid or expired refresh token")
	}

	// Verify user exists
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		uc.logger.Errorf("Failed to get user: %v", err)
		return nil, errors.NewInternalError("Token refresh failed", err)
	}
	if user == nil {
		return nil, errors.NewUnauthorizedError("User not found")
	}

	// Generate new token pair
	tokenPair, err := uc.jwtMgr.GenerateTokenPair(userID)
	if err != nil {
		uc.logger.Errorf("Failed to generate tokens: %v", err)
		return nil, errors.NewInternalError("Token refresh failed", err)
	}

	return tokenPair, nil
}
