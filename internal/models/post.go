// File: internal/models/post.go

package models

import "time"

// Post represents a user's submission within a subreddit.
type Post struct {
	ID          int       `json:"id"`                    // Unique identifier for the post.
	Title       string    `json:"title"`                 // Title of the post.
	Content     string    `json:"content"`               // Text content of the post.
	AuthorID    int       `json:"author_id"`             // ID of the user who created the post.
	SubredditID int       `json:"subreddit_id"`          // ID of the subreddit where the post was made.
	Karma       int       `json:"karma"`                 // Net upvotes minus downvotes.
	CreatedAt   time.Time `json:"created_at"`            // Timestamp of post creation.
	UpdatedAt   time.Time `json:"updated_at"`            // Timestamp of the last update to the post.
}
