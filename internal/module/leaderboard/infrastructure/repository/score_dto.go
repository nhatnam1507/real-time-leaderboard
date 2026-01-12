// Package repository provides repository implementations for the leaderboard module.
package repository

import "time"

// Score represents a score DTO for database operations
// This is an infrastructure concern and should not be exposed outside this package
type Score struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	Score     int64     `db:"score"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
