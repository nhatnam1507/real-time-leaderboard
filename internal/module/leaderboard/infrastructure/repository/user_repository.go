// Package repository provides repository implementations for the leaderboard module.
package repository

import (
	"context"
	"fmt"

	"real-time-leaderboard/internal/module/leaderboard/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresUserRepository implements UserRepository using PostgreSQL
// This is owned by the leaderboard module, not the auth module
type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new PostgreSQL user repository for the leaderboard module
func NewUserRepository(pool *pgxpool.Pool) domain.UserRepository {
	return &PostgresUserRepository{pool: pool}
}

// GetByIDs retrieves usernames for multiple user IDs in a single query
func (r *PostgresUserRepository) GetByIDs(ctx context.Context, userIDs []string) (map[string]string, error) {
	if len(userIDs) == 0 {
		return make(map[string]string), nil
	}

	query := `SELECT id, username FROM users WHERE id = ANY($1)`

	rows, err := r.pool.Query(ctx, query, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get users by IDs: %w", err)
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var id, username string
		if err := rows.Scan(&id, &username); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		result[id] = username
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return result, nil
}
