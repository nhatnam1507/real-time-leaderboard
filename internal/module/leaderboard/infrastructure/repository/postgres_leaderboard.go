// Package repository provides repository implementations for the leaderboard module.
package repository

import (
	"context"
	"fmt"
	"time"

	"real-time-leaderboard/internal/module/leaderboard/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresLeaderboardRepository implements LeaderboardBackupRepository using PostgreSQL
// Stores only the highest score per user as a backup/recovery mechanism for Redis
type PostgresLeaderboardRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresLeaderboardRepository creates a new PostgreSQL leaderboard backup repository
func NewPostgresLeaderboardRepository(pool *pgxpool.Pool) *PostgresLeaderboardRepository {
	return &PostgresLeaderboardRepository{pool: pool}
}

// UpsertScore upserts the score for a user
// If user doesn't exist, creates a new record with the given point
// If user exists, updates the score
func (r *PostgresLeaderboardRepository) UpsertScore(ctx context.Context, userID string, point int64) (*domain.Score, error) {
	now := time.Now()

	query := `
		INSERT INTO leaderboard (id, user_id, point, created_at, updated_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4)
		ON CONFLICT (user_id) 
		DO UPDATE SET 
			point = $2,
			updated_at = $4
		RETURNING id, user_id, point, created_at, updated_at
	`

	var score domain.Score
	err := r.pool.QueryRow(ctx, query, userID, point, now, now).Scan(
		&score.ID,
		&score.UserID,
		&score.Point,
		&score.CreatedAt,
		&score.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to upsert score: %w", err)
	}

	return &score, nil
}

// GetAllScores retrieves all leaderboard entries from PostgreSQL
// Used for syncing data to Redis when Redis is empty
func (r *PostgresLeaderboardRepository) GetAllScores(ctx context.Context) ([]domain.Score, error) {
	query := `
		SELECT id, user_id, point, created_at, updated_at
		FROM leaderboard
		ORDER BY point DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all scores: %w", err)
	}
	defer rows.Close()

	var scores []domain.Score
	for rows.Next() {
		var score domain.Score
		if err := rows.Scan(&score.ID, &score.UserID, &score.Point, &score.CreatedAt, &score.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan score: %w", err)
		}
		scores = append(scores, score)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating scores: %w", err)
	}

	return scores, nil
}
