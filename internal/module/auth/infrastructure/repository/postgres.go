// Package repository provides repository implementations for the auth module.
package repository

import (
	"context"
	"fmt"
	"time"

	"real-time-leaderboard/internal/module/auth/application"
	"real-time-leaderboard/internal/module/auth/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresUserRepository implements UserRepository using PostgreSQL
type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresUserRepository creates a new PostgreSQL user repository
func NewPostgresUserRepository(pool *pgxpool.Pool) application.UserRepository {
	return &PostgresUserRepository{pool: pool}
}

// Create creates a new user
func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	dto := &User{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Password: user.Password,
	}

	if dto.ID == "" {
		dto.ID = uuid.New().String()
	}
	now := time.Now()
	// Timestamps are infrastructure concerns, handled in DTO only
	dto.CreatedAt = now
	dto.UpdatedAt = now

	query := `
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pool.Exec(ctx, query,
		dto.ID,
		dto.Username,
		dto.Email,
		dto.Password,
		dto.CreatedAt,
		dto.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Update domain entity with generated ID only (timestamps stay in infrastructure)
	user.ID = dto.ID

	return nil
}

// GetByID retrieves a user by ID
func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var dto User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&dto.ID,
		&dto.Username,
		&dto.Email,
		&dto.Password,
		&dto.CreatedAt,
		&dto.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &domain.User{
		ID:       dto.ID,
		Username: dto.Username,
		Email:    dto.Email,
		Password: dto.Password,
		// Timestamps are infrastructure concerns, not part of domain entity
	}, nil
}

// GetByUsername retrieves a user by username
func (r *PostgresUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	var dto User
	err := r.pool.QueryRow(ctx, query, username).Scan(
		&dto.ID,
		&dto.Username,
		&dto.Email,
		&dto.Password,
		&dto.CreatedAt,
		&dto.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &domain.User{
		ID:       dto.ID,
		Username: dto.Username,
		Email:    dto.Email,
		Password: dto.Password,
		// Timestamps are infrastructure concerns, not part of domain entity
	}, nil
}

// GetByEmail retrieves a user by email
func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var dto User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&dto.ID,
		&dto.Username,
		&dto.Email,
		&dto.Password,
		&dto.CreatedAt,
		&dto.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &domain.User{
		ID:       dto.ID,
		Username: dto.Username,
		Email:    dto.Email,
		Password: dto.Password,
		// Timestamps are infrastructure concerns, not part of domain entity
	}, nil
}

// Update updates a user
func (r *PostgresUserRepository) Update(ctx context.Context, user *domain.User) error {
	dto := &User{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Password: user.Password,
		// Timestamps are infrastructure concerns, handled in DTO only
		UpdatedAt: time.Now(),
	}

	query := `
		UPDATE users
		SET username = $2, email = $3, password_hash = $4, updated_at = $5
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query,
		dto.ID,
		dto.Username,
		dto.Email,
		dto.Password,
		dto.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Domain entity doesn't need timestamp updates (infrastructure concern)

	return nil
}

// Delete deletes a user
func (r *PostgresUserRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`

	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}
