// Package repository provides repository implementations for the leaderboard module.
package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"real-time-leaderboard/internal/module/leaderboard/domain"
)

// PostgresScoreRepository implements ScoreRepository using PostgreSQL
type PostgresScoreRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresScoreRepository creates a new PostgreSQL score repository
func NewPostgresScoreRepository(pool *pgxpool.Pool) *PostgresScoreRepository {
	return &PostgresScoreRepository{pool: pool}
}

// Create creates a new score
func (r *PostgresScoreRepository) Create(ctx context.Context, score *domain.Score) error {
	if score.ID == "" {
		score.ID = uuid.New().String()
	}
	if score.SubmittedAt.IsZero() {
		score.SubmittedAt = time.Now()
	}

	var metadataJSON []byte
	if score.Metadata != nil {
		metadataJSON = score.Metadata
	}

	query := `
		INSERT INTO scores (id, user_id, score, submitted_at, metadata)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.pool.Exec(ctx, query,
		score.ID,
		score.UserID,
		score.Score,
		score.SubmittedAt,
		metadataJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to create score: %w", err)
	}

	return nil
}

// GetHighestByUserID retrieves the highest score by user ID
func (r *PostgresScoreRepository) GetHighestByUserID(ctx context.Context, userID string) (*domain.Score, error) {
	query := `
		SELECT id, user_id, score, submitted_at, metadata
		FROM scores
		WHERE user_id = $1
		ORDER BY score DESC
		LIMIT 1
	`

	var score domain.Score
	var metadataJSON []byte

	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&score.ID,
		&score.UserID,
		&score.Score,
		&score.SubmittedAt,
		&metadataJSON,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get highest score by user id: %w", err)
	}

	if len(metadataJSON) > 0 {
		score.Metadata = json.RawMessage(metadataJSON)
	}

	return &score, nil
}
