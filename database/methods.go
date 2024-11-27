package database

import (
	"fmt"
)

// User Operations
func (dm *DatabaseManager) RegisterUser(username, password string) (int, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	query := `INSERT INTO users (username, password) VALUES (?, ?)`
	result, err := dm.db.Exec(query, username, password) // In real app, hash the password
	if err != nil {
		return 0, fmt.Errorf("failed to register user: %v", err)
	}

	id, err := result.LastInsertId()
	return int(id), err
}

func (dm *DatabaseManager) GetUserByUsername(username string) (*User, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	var user User
	query := `SELECT id, username, karma FROM users WHERE username = ?`
	err := dm.db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Karma)
	if err != nil {
		return nil, fmt.Errorf("user not found: %v", err)
	}

	return &user, nil
}

// Subreddit Operations
func (dm *DatabaseManager) CreateSubreddit(name, description string, creatorID int) (int, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	tx, err := dm.db.Begin()
	if err != nil {
		return 0, err
	}

	// Create subreddit
	result, err := tx.Exec(`INSERT INTO subreddits (name, description) VALUES (?, ?)`, name, description)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("failed to create subreddit: %v", err)
	}

	subredditID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	// Add creator as first member
	_, err = tx.Exec(`
		INSERT INTO subreddit_members (subreddit_id, user_id) 
		VALUES (?, ?)
	`, subredditID, creatorID)

	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("failed to add creator to subreddit: %v", err)
	}

	err = tx.Commit()
	return int(subredditID), err
}

func (dm *DatabaseManager) JoinSubreddit(userID, subredditID int) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	_, err := dm.db.Exec(`
		INSERT OR IGNORE INTO subreddit_members (subreddit_id, user_id) 
		VALUES (?, ?)
	`, subredditID, userID)

	return err
}

func (dm *DatabaseManager) LeaveSubreddit(userID, subredditID int) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	_, err := dm.db.Exec(`
		DELETE FROM subreddit_members 
		WHERE subreddit_id = ? AND user_id = ?
	`, subredditID, userID)

	return err
}

// Post Operations
func (dm *DatabaseManager) CreatePost(title, content string, authorID, subredditID int) (int, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	result, err := dm.db.Exec(`
		INSERT INTO posts (title, content, author_id, subreddit_id) 
		VALUES (?, ?, ?, ?)
	`, title, content, authorID, subredditID)

	if err != nil {
		return 0, fmt.Errorf("failed to create post: %v", err)
	}

	id, err := result.LastInsertId()
	return int(id), err
}

func (dm *DatabaseManager) GetFeed(userID int) ([]Post, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	query := `
		SELECT p.id, p.title, p.content, p.author_id, p.subreddit_id, p.created_at,
			   u.username AS author_username, s.name AS subreddit_name,
			(SELECT COUNT(*) FROM votes WHERE target_id = p.id AND target_type = 'post' AND vote_value = 1) AS upvotes,
            (SELECT COUNT(*) FROM votes WHERE target_id = p.id AND target_type = 'post' AND vote_value = -1) AS downvotes
		FROM posts p
		JOIN subreddit_members sm ON p.subreddit_id = sm.subreddit_id
		JOIN users u ON p.author_id = u.id
		JOIN subreddits s ON p.subreddit_id = s.id
		WHERE sm.user_id = ?
		ORDER BY p.created_at DESC
	`

	rows, err := dm.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(
			&post.ID, &post.Title, &post.Content, &post.AuthorID,
			&post.SubredditID, &post.CreatedAt,
			&post.AuthorUsername, &post.SubredditName,
			&post.VoteCount.Upvotes, &post.VoteCount.Downvotes,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

// Voting Operations
func (dm *DatabaseManager) Vote(userID, targetID int, targetType string, value int) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	tx, err := dm.db.Begin()
	if err != nil {
		return err
	}

	// Upsert vote
	_, err = tx.Exec(`
		INSERT INTO votes (user_id, target_id, target_type, vote_value) 
		VALUES (?, ?, ?, ?)
	`, userID, targetID, targetType, value)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to record vote: %v", err)
	}

	// Update karma based on vote type and target
	var updateQuery string
	if targetType == "post" {
		updateQuery = `
			UPDATE users 
			SET karma = karma + ? 
			WHERE id = (SELECT author_id FROM posts WHERE id = ?)
		`
	} else { // comment
		updateQuery = `
			UPDATE users 
			SET karma = karma + ? 
			WHERE id = (SELECT author_id FROM comments WHERE id = ?)
		`
	}

	_, err = tx.Exec(updateQuery, value, targetID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update karma: %v", err)
	}

	return tx.Commit()
}

// Comment Operations
func (dm *DatabaseManager) CreateComment(content string, authorID, postID int, parentCommentID *int) (int, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	query := `
		INSERT INTO comments (content, author_id, post_id, parent_comment_id) 
		VALUES (?, ?, ?, ?)
	`

	result, err := dm.db.Exec(query, content, authorID, postID, parentCommentID)
	if err != nil {
		return 0, fmt.Errorf("failed to create comment: %v", err)
	}

	id, err := result.LastInsertId()
	return int(id), err
}

// Direct Messaging Operations
func (dm *DatabaseManager) SendDirectMessage(fromUserID, toUserID int, content string) (int, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	result, err := dm.db.Exec(`
		INSERT INTO direct_messages (from_user_id, to_user_id, content) 
		VALUES (?, ?, ?)
	`, fromUserID, toUserID, content)

	if err != nil {
		return 0, fmt.Errorf("failed to send message: %v", err)
	}

	id, err := result.LastInsertId()
	return int(id), err
}

func (dm *DatabaseManager) GetDirectMessages(userID int) ([]DirectMessage, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	query := `
		SELECT 
			dm.id, 
			dm.from_user_id, 
			u.username AS from_username, 
			dm.content, 
			dm.created_at
		FROM direct_messages dm
		JOIN users u ON dm.from_user_id = u.id
		WHERE dm.to_user_id = ?
		ORDER BY dm.created_at DESC
	`

	rows, err := dm.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []DirectMessage
	for rows.Next() {
		var msg DirectMessage
		err := rows.Scan(
			&msg.ID,
			&msg.FromUserID,
			&msg.FromUsername,
			&msg.Content,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func (dm *DatabaseManager) GetTopUsers(limit int) ([]TopUser, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	query := `
        SELECT 
            u.id,
            u.username,
            u.karma,
            (SELECT COUNT(*) FROM posts WHERE author_id = u.id) as post_count,
            (SELECT COUNT(*) FROM comments WHERE author_id = u.id) as comment_count
        FROM users u
        ORDER BY u.karma DESC
        LIMIT ?
    `

	rows, err := dm.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []TopUser
	for rows.Next() {
		var user TopUser
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Karma,
			&user.PostCount,
			&user.CommentCount,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// Close closes the database connection
func (dm *DatabaseManager) Close() {
	if dm.db != nil {
		dm.db.Close()
	}
}

func (dm *DatabaseManager) ResetDatabase() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	tables := []string{
		"direct_messages",
		"votes",
		"comments",
		"posts",
		"subreddit_members",
		"subreddits",
		"users",
	}

	tx, err := dm.db.Begin()
	if err != nil {
		return err
	}

	// Delete all rows from tables
	for _, table := range tables {
		_, err = tx.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete from %s: %v", table, err)
		}
	}

	for _, table := range tables {
		_, err = tx.Exec(fmt.Sprintf("DELETE FROM sqlite_sequence WHERE name='%s'", table))
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to reset auto-increment for %s: %v", table, err)
		}
	}

	return tx.Commit()
}
