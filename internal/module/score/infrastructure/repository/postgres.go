package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"real-time-leaderboard/internal/module/score/domain"
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
		INSERT INTO scores (id, user_id, game_id, score, submitted_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pool.Exec(ctx, query,
		score.ID,
		score.UserID,
		score.GameID,
		score.Score,
		score.SubmittedAt,
		metadataJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to create score: %w", err)
	}

	return nil
}

// GetByUserID retrieves scores by user ID
func (r *PostgresScoreRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.Score, error) {
	query := `
		SELECT id, user_id, game_id, score, submitted_at, metadata
		FROM scores
		WHERE user_id = $1
		ORDER BY submitted_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get scores by user id: %w", err)
	}
	defer rows.Close()

	var scores []*domain.Score
	for rows.Next() {
		var score domain.Score
		var metadataJSON []byte

		err := rows.Scan(
			&score.ID,
			&score.UserID,
			&score.GameID,
			&score.Score,
			&score.SubmittedAt,
			&metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan score: %w", err)
		}

		if len(metadataJSON) > 0 {
			score.Metadata = json.RawMessage(metadataJSON)
		}

		scores = append(scores, &score)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating scores: %w", err)
	}

	return scores, nil
}

// GetByUserIDAndGameID retrieves scores by user ID and game ID
func (r *PostgresScoreRepository) GetByUserIDAndGameID(ctx context.Context, userID, gameID string, limit, offset int) ([]*domain.Score, error) {
	query := `
		SELECT id, user_id, game_id, score, submitted_at, metadata
		FROM scores
		WHERE user_id = $1 AND game_id = $2
		ORDER BY submitted_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.pool.Query(ctx, query, userID, gameID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get scores by user id and game id: %w", err)
	}
	defer rows.Close()

	var scores []*domain.Score
	for rows.Next() {
		var score domain.Score
		var metadataJSON []byte

		err := rows.Scan(
			&score.ID,
			&score.UserID,
			&score.GameID,
			&score.Score,
			&score.SubmittedAt,
			&metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan score: %w", err)
		}

		if len(metadataJSON) > 0 {
			score.Metadata = json.RawMessage(metadataJSON)
		}

		scores = append(scores, &score)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating scores: %w", err)
	}

	return scores, nil
}

// GetHighestByUserID retrieves the highest score by user ID
func (r *PostgresScoreRepository) GetHighestByUserID(ctx context.Context, userID string) (*domain.Score, error) {
	query := `
		SELECT id, user_id, game_id, score, submitted_at, metadata
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
		&score.GameID,
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

// GetHighestByUserIDAndGameID retrieves the highest score by user ID and game ID
func (r *PostgresScoreRepository) GetHighestByUserIDAndGameID(ctx context.Context, userID, gameID string) (*domain.Score, error) {
	query := `
		SELECT id, user_id, game_id, score, submitted_at, metadata
		FROM scores
		WHERE user_id = $1 AND game_id = $2
		ORDER BY score DESC
		LIMIT 1
	`

	var score domain.Score
	var metadataJSON []byte

	err := r.pool.QueryRow(ctx, query, userID, gameID).Scan(
		&score.ID,
		&score.UserID,
		&score.GameID,
		&score.Score,
		&score.SubmittedAt,
		&metadataJSON,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get highest score by user id and game id: %w", err)
	}

	if len(metadataJSON) > 0 {
		score.Metadata = json.RawMessage(metadataJSON)
	}

	return &score, nil
}

