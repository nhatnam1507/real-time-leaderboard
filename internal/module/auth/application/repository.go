// Package application provides use cases for the auth module.
package application

//go:generate mockgen -source=repository.go -destination=../mocks/repository_mock.go -package=mocks UserRepository

import (
	"context"

	"real-time-leaderboard/internal/module/auth/domain"
)

// UserRepository defines the interface for user data operations in the auth module
// This interface belongs to the auth module application layer
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
}
