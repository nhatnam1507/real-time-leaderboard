// Package application provides use cases for the auth module.
package application

//go:generate mockgen -destination=../infrastructure/mocks/jwt_manager_mock.go -package=mocks real-time-leaderboard/internal/module/auth/application JWTManager

import (
	"context"
	"fmt"

	"real-time-leaderboard/internal/module/auth/domain"
	"real-time-leaderboard/internal/shared/logger"

	"golang.org/x/crypto/bcrypt"
)

//go:generate mockgen -destination=../adapters/mocks/auth_usecase_mock.go -package=mocks real-time-leaderboard/internal/module/auth/application AuthUseCase

// AuthUseCase defines the interface for authentication operations
type AuthUseCase interface {
	Register(ctx context.Context, req RegisterRequest) (*domain.User, *domain.TokenPair, error)
	Login(ctx context.Context, req LoginRequest) (*domain.User, *domain.TokenPair, error)
	ValidateToken(ctx context.Context, token string) (string, error)
	RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error)
	GetCurrentUser(ctx context.Context, userID string) (*domain.User, error)
}

// authUseCase implements AuthUseCase interface
type authUseCase struct {
	userRepo UserRepository
	jwtMgr   JWTManager
	logger   *logger.Logger
}

// JWTManager interface for JWT operations
type JWTManager interface {
	GenerateTokenPair(userID string) (*domain.TokenPair, error)
	ValidateToken(token string) (string, error)
}

// NewAuthUseCase creates a new auth use case
//
//nolint:revive // unexported-return: intentional design - accept interface, return struct
func NewAuthUseCase(userRepo UserRepository, jwtMgr JWTManager, l *logger.Logger) *authUseCase {
	return &authUseCase{
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
func (uc *authUseCase) Register(ctx context.Context, req RegisterRequest) (*domain.User, *domain.TokenPair, error) {
	// Check if username already exists
	existingUser, err := uc.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		uc.logger.Errorf(ctx, "Failed to check username existence: %v", err)
		return nil, nil, fmt.Errorf("failed to check username existence: %w", err)
	}
	if existingUser != nil {
		return nil, nil, fmt.Errorf("%w: username", domain.ErrUserAlreadyExists)
	}

	// Check if email already exists
	existingUser, err = uc.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		uc.logger.Errorf(ctx, "Failed to check email existence: %v", err)
		return nil, nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if existingUser != nil {
		return nil, nil, fmt.Errorf("%w: email", domain.ErrUserAlreadyExists)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		uc.logger.Errorf(ctx, "Failed to hash password: %v", err)
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &domain.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		uc.logger.Errorf(ctx, "Failed to create user: %v", err)
		return nil, nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	tokenPair, err := uc.jwtMgr.GenerateTokenPair(user.ID)
	if err != nil {
		uc.logger.Errorf(ctx, "Failed to generate tokens: %v", err)
		return nil, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	uc.logger.Infof(ctx, "User registered: %s", user.ID)
	return user, tokenPair, nil
}

// Login authenticates a user
func (uc *authUseCase) Login(ctx context.Context, req LoginRequest) (*domain.User, *domain.TokenPair, error) {
	// Get user by username
	user, err := uc.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		uc.logger.Errorf(ctx, "Failed to get user: %v", err)
		return nil, nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, nil, domain.ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, nil, domain.ErrInvalidCredentials
	}

	// Generate tokens
	tokenPair, err := uc.jwtMgr.GenerateTokenPair(user.ID)
	if err != nil {
		uc.logger.Errorf(ctx, "Failed to generate tokens: %v", err)
		return nil, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	uc.logger.Infof(ctx, "User logged in: %s", user.ID)
	return user, tokenPair, nil
}

// ValidateToken validates a JWT token and returns the user ID
func (uc *authUseCase) ValidateToken(ctx context.Context, token string) (string, error) {
	userID, err := uc.jwtMgr.ValidateToken(token)
	if err != nil {
		return "", fmt.Errorf("%w: %v", domain.ErrInvalidToken, err)
	}

	// Optionally verify user still exists
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		uc.logger.Errorf(ctx, "Failed to get user: %v", err)
		return "", fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return "", domain.ErrUserNotFound
	}

	return userID, nil
}

// RefreshToken refreshes an access token using a refresh token
func (uc *authUseCase) RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	userID, err := uc.jwtMgr.ValidateToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("%w: refresh token: %v", domain.ErrInvalidToken, err)
	}

	// Verify user exists
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		uc.logger.Errorf(ctx, "Failed to get user: %v", err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	// Generate new token pair
	tokenPair, err := uc.jwtMgr.GenerateTokenPair(userID)
	if err != nil {
		uc.logger.Errorf(ctx, "Failed to generate tokens: %v", err)
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return tokenPair, nil
}

// GetCurrentUser retrieves the current user by ID
func (uc *authUseCase) GetCurrentUser(ctx context.Context, userID string) (*domain.User, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		uc.logger.Errorf(ctx, "Failed to get user: %v", err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	return user, nil
}
