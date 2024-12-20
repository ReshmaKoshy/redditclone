// File: internal/models/vote.go

package models

import "time"

// VoteType defines the type of vote: upvote or downvote.
// type VoteType string

// const (
// 	Upvote   VoteType = "upvote"
// 	Downvote VoteType = "downvote"
// )

// Vote represents a user's vote on a post or comment.
type Vote struct {
	ID          int       `json:"id"`                     // Unique identifier for the vote.
	UserID      int       `json:"user_id"`                // ID of the user who cast the vote.
	PostID      *int      `json:"post_id,omitempty"`      // ID of the post being voted on (if applicable).
	CommentID   *int      `json:"comment_id,omitempty"`   // ID of the comment being voted on (if applicable).
	VoteType    string  `json:"vote_type"`              // Type of vote: upvote or downvote.
	CreatedAt   time.Time `json:"created_at"`             // Timestamp of when the vote was cast.
	UpdatedAt   time.Time `json:"updated_at"`             // Timestamp of the last update to the vote.
}
