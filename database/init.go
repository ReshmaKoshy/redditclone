package database

import (
	"database/sql"
	"fmt"
	"sync"

	_ "modernc.org/sqlite"
)

type DatabaseManager struct {
	db *sql.DB
	mu sync.RWMutex
}

func InitDatabase(dbPath string) (*DatabaseManager, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Create tables
	_, err = db.Exec(`
		-- Users table
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			karma INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		-- Subreddits table
		CREATE TABLE IF NOT EXISTS subreddits (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		-- Subreddit Members table
		CREATE TABLE IF NOT EXISTS subreddit_members (
			subreddit_id INTEGER,
			user_id INTEGER,
			joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (subreddit_id, user_id),
			FOREIGN KEY (subreddit_id) REFERENCES subreddits(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		);

		-- Posts table
		CREATE TABLE IF NOT EXISTS posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			author_id INTEGER NOT NULL,
			subreddit_id INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (author_id) REFERENCES users(id),
			FOREIGN KEY (subreddit_id) REFERENCES subreddits(id)
		);

		-- Comments table (supports hierarchical comments)
		CREATE TABLE IF NOT EXISTS comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			content TEXT NOT NULL,
			author_id INTEGER NOT NULL,
			post_id INTEGER,
			parent_comment_id INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (author_id) REFERENCES users(id),
			FOREIGN KEY (post_id) REFERENCES posts(id),
			FOREIGN KEY (parent_comment_id) REFERENCES comments(id)
		);

		-- Votes table (for posts and comments)
		CREATE TABLE IF NOT EXISTS votes (
			user_id INTEGER NOT NULL,
			target_id INTEGER NOT NULL,
			target_type TEXT CHECK(target_type IN ('post', 'comment')) NOT NULL,
			vote_value INTEGER CHECK(vote_value IN (-1, 1)) NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, target_id, target_type, vote_value),
			FOREIGN KEY (user_id) REFERENCES users(id)
		);

		-- Direct Messages table
		CREATE TABLE IF NOT EXISTS direct_messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			from_user_id INTEGER NOT NULL,
			to_user_id INTEGER NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (from_user_id) REFERENCES users(id),
			FOREIGN KEY (to_user_id) REFERENCES users(id)
		);
	`)

	if err != nil {
		return nil, fmt.Errorf("failed to create tables: %v", err)
	}

	return &DatabaseManager{db: db}, nil
}
