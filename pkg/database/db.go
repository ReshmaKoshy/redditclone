// File: pkg/database/db.go

package database

import (
	"database/sql"
    "sync"
	_ "modernc.org/sqlite"
)
  
var (
    DB   *sql.DB
    DBMu sync.Mutex
)

  func Connect() (*sql.DB, error) {
	// For an in-memory database:
	db, err := sql.Open("sqlite", "reddit_clone.db")
	if err != nil {
	  return nil, err
	}
  
	// By default, SQLite doesn't enforce foreign keys without PRAGMA
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
	  return nil, err
	}
    DB = db
	return DB, nil
  }
  

  func InitializeSchema(db *sql.DB) error {
    // Enable foreign key enforcement in SQLite
    if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
        return err
    }

    // Users table
    createUserTable := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT UNIQUE NOT NULL,
        email TEXT UNIQUE NOT NULL,
        password TEXT NOT NULL,
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
    );`
    if _, err := db.Exec(createUserTable); err != nil {
        return err
    }

    // Subreddits table
    createSubredditTable := `
    CREATE TABLE IF NOT EXISTS subreddits (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT UNIQUE NOT NULL,
        description TEXT,
        created_by INTEGER NOT NULL,
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY(created_by) REFERENCES users(id) ON DELETE CASCADE
    );`
    if _, err := db.Exec(createSubredditTable); err != nil {
        return err
    }

    // Posts table
    createPostTable := `
    CREATE TABLE IF NOT EXISTS posts (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL,
        content TEXT NOT NULL,
        author_id INTEGER NOT NULL,
        subreddit_id INTEGER NOT NULL,
        karma INTEGER NOT NULL DEFAULT 0,
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY(author_id) REFERENCES users(id) ON DELETE CASCADE,
        FOREIGN KEY(subreddit_id) REFERENCES subreddits(id) ON DELETE CASCADE
    );`
    if _, err := db.Exec(createPostTable); err != nil {
        return err
    }

    // Comments table
    createCommentTable := `
    CREATE TABLE IF NOT EXISTS comments (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        content TEXT NOT NULL,
        author_id INTEGER NOT NULL,
        post_id INTEGER NOT NULL,
        parent_id INTEGER,
        karma INTEGER NOT NULL DEFAULT 0,
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY(author_id) REFERENCES users(id) ON DELETE CASCADE,
        FOREIGN KEY(post_id) REFERENCES posts(id) ON DELETE CASCADE,
        FOREIGN KEY(parent_id) REFERENCES comments(id) ON DELETE CASCADE
    );`
    if _, err := db.Exec(createCommentTable); err != nil {
        return err
    }

    // Votes table
    createVoteTable := `
    CREATE TABLE IF NOT EXISTS votes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        post_id INTEGER,
        comment_id INTEGER,
        vote_type TEXT NOT NULL CHECK (vote_type IN ('upvote', 'downvote')),
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
        FOREIGN KEY(post_id) REFERENCES posts(id) ON DELETE CASCADE,
        FOREIGN KEY(comment_id) REFERENCES comments(id) ON DELETE CASCADE,
        UNIQUE(user_id, post_id, comment_id)
    );`
    if _, err := db.Exec(createVoteTable); err != nil {
        return err
    }

    // Messages table
    createMessageTable := `
    CREATE TABLE IF NOT EXISTS messages (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        sender_id INTEGER NOT NULL,
        receiver_id INTEGER NOT NULL,
        content TEXT NOT NULL,
        parent_id INTEGER,
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY(sender_id) REFERENCES users(id) ON DELETE CASCADE,
        FOREIGN KEY(receiver_id) REFERENCES users(id) ON DELETE CASCADE,
        FOREIGN KEY(parent_id) REFERENCES messages(id) ON DELETE CASCADE
    );`
    if _, err := db.Exec(createMessageTable); err != nil {
        return err
    }

    return nil
}
