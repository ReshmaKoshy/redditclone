package database

import "time"

// Structs for database operations
type User struct {
	ID       string
	Username string
	Karma    int
}

type Post struct {
	ID             int
	Title          string
	Content        string
	AuthorID       int    `json:"author_id"`
	AuthorUsername string `json:"author_name"`
	SubredditID    int    `json:"subreddit_id"`
	SubredditName  string `json:"subreddit_name"`
	CreatedAt      time.Time
	VoteCount      struct {
		Upvotes   int `json:"upvotes"`
		Downvotes int `json:"downvotes"`
	} `json:"vote_count"`
}

type DirectMessage struct {
	ID           int
	FromUserID   int
	FromUsername string
	Content      string
	CreatedAt    time.Time
}

type TopUser struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	Karma        int    `json:"karma"`
	PostCount    int    `json:"post_count"`
	CommentCount int    `json:"comment_count"`
}
