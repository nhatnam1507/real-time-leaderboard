package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"real-time-leaderboard/internal/module/report/domain"
)

// PostgresReportRepository implements ReportRepository using PostgreSQL for historical data
type PostgresReportRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresReportRepository creates a new PostgreSQL report repository
func NewPostgresReportRepository(pool *pgxpool.Pool) *PostgresReportRepository {
	return &PostgresReportRepository{pool: pool}
}

// GetTopPlayersByDateRange retrieves top players within a date range from PostgreSQL
func (r *PostgresReportRepository) GetTopPlayersByDateRange(ctx context.Context, gameID string, startDate, endDate time.Time, limit, offset int64) ([]domain.TopPlayer, error) {
	var query string
	var args []interface{}

	if gameID == "" || gameID == "global" {
		// Global leaderboard - sum scores across all games
		query = `
			SELECT user_id, SUM(score) as total_score, MAX(submitted_at) as last_updated
			FROM scores
			WHERE submitted_at >= $1 AND submitted_at <= $2
			GROUP BY user_id
			ORDER BY total_score DESC
			LIMIT $3 OFFSET $4
		`
		args = []interface{}{startDate, endDate, limit, offset}
	} else {
		// Game-specific leaderboard
		query = `
			SELECT user_id, MAX(score) as max_score, MAX(submitted_at) as last_updated
			FROM scores
			WHERE game_id = $1 AND submitted_at >= $2 AND submitted_at <= $3
			GROUP BY user_id
			ORDER BY max_score DESC
			LIMIT $4 OFFSET $5
		`
		args = []interface{}{gameID, startDate, endDate, limit, offset}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query top players: %w", err)
	}
	defer rows.Close()

	var players []domain.TopPlayer
	rank := offset + 1 // Rank accounts for offset

	for rows.Next() {
		var player domain.TopPlayer
		var lastUpdated time.Time

		if gameID == "" || gameID == "global" {
			var totalScore int64
			err := rows.Scan(&player.UserID, &totalScore, &lastUpdated)
			if err != nil {
				return nil, fmt.Errorf("failed to scan player: %w", err)
			}
			player.Score = totalScore
		} else {
			err := rows.Scan(&player.UserID, &player.Score, &lastUpdated)
			if err != nil {
				return nil, fmt.Errorf("failed to scan player: %w", err)
			}
		}

		player.Rank = rank
		player.GameID = gameID
		player.LastUpdated = lastUpdated
		players = append(players, player)
		rank++
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating players: %w", err)
	}

	return players, nil
}
