package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"real-time-leaderboard/internal/module/auth/domain"
)

// PostgresUserRepository implements UserRepository using PostgreSQL
type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresUserRepository creates a new PostgreSQL user repository
func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

// Create creates a new user
func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	query := `
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pool.Exec(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.Password,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *PostgresUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// Update updates a user
func (r *PostgresUserRepository) Update(ctx context.Context, user *domain.User) error {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users
		SET username = $2, email = $3, password_hash = $4, updated_at = $5
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.Password,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

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
