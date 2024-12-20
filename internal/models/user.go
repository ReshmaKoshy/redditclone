// File: internal/models/user.go

package models

import "time"

// User represents a registered user in the system.
type User struct {
	ID        int       `json:"id"`                  // Unique identifier for the user.
	Username  string    `json:"username"`            // Unique username chosen by the user.
	Email     string    `json:"email"`               // User's email address.
	Password  string    `json:"password"`                   // Hashed password. Omitted from JSON responses for security.
	CreatedAt time.Time `json:"created_at"`          // Timestamp of when the user was created.
	UpdatedAt time.Time `json:"updated_at"`          // Timestamp of the last update to the user's information.
}
