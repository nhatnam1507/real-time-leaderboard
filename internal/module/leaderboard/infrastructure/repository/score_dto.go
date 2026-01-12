// Package repository provides repository implementations for the leaderboard module.
package repository

import "time"

// Score represents a score DTO for database operations
// This is an infrastructure concern and should not be exposed outside this package
type Score struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	Point     int64     `db:"point"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
