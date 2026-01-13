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

// PostgresLeaderboardRepository implements LeaderboardBackupRepository using PostgreSQL
// Stores only the highest score per user as a backup/recovery mechanism for Redis
type PostgresLeaderboardRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresLeaderboardRepository creates a new PostgreSQL leaderboard backup repository
func NewPostgresLeaderboardRepository(pool *pgxpool.Pool) application.LeaderboardBackupRepository {
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

// GetLeaderboard retrieves the full leaderboard from PostgreSQL with usernames
// Used for syncing data to Redis when Redis is empty
func (r *PostgresLeaderboardRepository) GetLeaderboard(ctx context.Context) (*domain.Leaderboard, error) {
	query := `
		SELECT 
			l.user_id,
			u.username,
			l.score,
			ROW_NUMBER() OVER (ORDER BY l.score DESC) as rank
		FROM leaderboard l
		JOIN users u ON l.user_id = u.id
		ORDER BY l.score DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}
	defer rows.Close()

	var entries []domain.LeaderboardEntry
	for rows.Next() {
		var entry domain.LeaderboardEntry
		if err := rows.Scan(&entry.UserID, &entry.Username, &entry.Score, &entry.Rank); err != nil {
			return nil, fmt.Errorf("failed to scan leaderboard entry: %w", err)
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating leaderboard entries: %w", err)
	}

	return &domain.Leaderboard{
		Entries: entries,
		Total:   int64(len(entries)),
	}, nil
}
