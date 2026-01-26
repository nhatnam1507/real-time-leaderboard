// Package repository provides repository implementations for the leaderboard module.
package repository

import (
	"context"
	"fmt"
	"time"

	"real-time-leaderboard/internal/module/leaderboard/application"
	"real-time-leaderboard/internal/module/leaderboard/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresLeaderboardRepository implements LeaderboardPersistenceRepository using PostgreSQL
// Stores the highest score per user as persistent storage
type PostgresLeaderboardRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresLeaderboardRepository creates a new PostgreSQL leaderboard persistence repository
func NewPostgresLeaderboardRepository(pool *pgxpool.Pool) application.LeaderboardPersistenceRepository {
	return &PostgresLeaderboardRepository{pool: pool}
}

// UpsertScore upserts the score for a user
// If user doesn't exist, creates a new record with the given score
// If user exists, updates the score
func (r *PostgresLeaderboardRepository) UpsertScore(ctx context.Context, userID string, score int64) error {
	now := time.Now()

	query := `
		INSERT INTO leaderboard (id, user_id, score, created_at, updated_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4)
		ON CONFLICT (user_id) 
		DO UPDATE SET 
			score = $2,
			updated_at = $4
	`

	_, err := r.pool.Exec(ctx, query, userID, score, now, now)
	if err != nil {
		return fmt.Errorf("failed to upsert score: %w", err)
	}

	return nil
}

// GetLeaderboard retrieves a paginated leaderboard from PostgreSQL with usernames and total count
func (r *PostgresLeaderboardRepository) GetLeaderboard(ctx context.Context, limit, offset int64) ([]domain.LeaderboardEntry, int64, error) {
	query := `
		SELECT 
			l.user_id,
			u.username,
			l.score,
			ROW_NUMBER() OVER (ORDER BY l.score DESC) as rank,
			COUNT(*) OVER() as total
		FROM leaderboard l
		JOIN users u ON l.user_id = u.id
		ORDER BY l.score DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get leaderboard: %w", err)
	}
	defer rows.Close()

	var entries []domain.LeaderboardEntry
	var total int64
	for rows.Next() {
		var entry domain.LeaderboardEntry
		if err := rows.Scan(&entry.UserID, &entry.Username, &entry.Score, &entry.Rank, &total); err != nil {
			return nil, 0, fmt.Errorf("failed to scan leaderboard entry: %w", err)
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating leaderboard entries: %w", err)
	}

	return entries, total, nil
}
