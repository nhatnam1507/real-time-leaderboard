// Package repository provides repository implementations for the auth module.
package repository

import "time"

// User represents a user DTO for database operations
// This is an infrastructure concern and should not be exposed outside this package
type User struct {
	ID        string    `db:"id"`
	Username  string    `db:"username"`
	Email     string    `db:"email"`
	Password  string    `db:"password_hash"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
