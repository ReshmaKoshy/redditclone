// File: internal/models/subreddit.go

package models

import "time"

// Subreddit represents a community where users can post and interact.
type Subreddit struct {
	ID          int       `json:"id"`                   // Unique identifier for the subreddit.
	Name        string    `json:"name"`                 // Unique name of the subreddit (e.g., "golang").
	Description string    `json:"description"`          // Brief description of the subreddit's purpose.
	CreatedBy   int       `json:"created_by"`           // ID of the user who created the subreddit.
	CreatedAt   time.Time `json:"created_at"`           // Timestamp of subreddit creation.
	UpdatedAt   time.Time `json:"updated_at"`           // Timestamp of the last update to the subreddit.
}
