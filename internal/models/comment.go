// File: internal/models/comment.go

package models

import "time"

// Comment represents a user's response to a post or another comment.
type Comment struct {
	ID         int       `json:"id"`                   // Unique identifier for the comment.
	Content    string    `json:"content"`              // Text content of the comment.
	AuthorID   int       `json:"author_id"`            // ID of the user who made the comment.
	PostID     int       `json:"post_id"`              // ID of the post the comment is associated with.
	ParentID   *int      `json:"parent_id,omitempty"`  // ID of the parent comment, if it's a reply.
	Karma      int       `json:"karma"`                // Net upvotes minus downvotes.
	CreatedAt  time.Time `json:"created_at"`           // Timestamp of comment creation.
	UpdatedAt  time.Time `json:"updated_at"`           // Timestamp of the last update to the comment.
}
