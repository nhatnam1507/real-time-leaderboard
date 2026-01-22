// Package domain provides domain entities for the auth module.
package domain

// User represents a user entity (pure business concept)
type User struct {
	ID       string `json:"id"`        // User identifier (used for business logic like JWT tokens)
	Username string `json:"username"`   // User's username
	Email    string `json:"email"`      // User's email address
	Password string `json:"-"`          // Never serialize password
}
