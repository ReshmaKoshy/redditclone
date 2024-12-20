// File: internal/repository/post_repository.go

package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"redditclone/internal/models"
	"redditclone/pkg/database"
)

// PostRepository provides access to the posts storage.
type PostRepository interface {
	CreatePost(post *models.Post) error
	GetPostByID(id int) (*models.Post, error)
	GetPostsBySubreddit(subredditID int, limit, offset int) ([]*models.Post, error)
	GetFeedPosts(limit, offset int) ([]*models.Post, error)
	UpdatePost(post *models.Post) error
	DeletePost(id int) error
}

type postRepository struct {
	DB *sql.DB
}

// NewPostRepository creates a new PostRepository.
func NewPostRepository(db *sql.DB) PostRepository {
	return &postRepository{DB: db}
}

// CreatePost inserts a new post into the database.
func (r *postRepository) CreatePost(post *models.Post) error {
	database.DBMu.Lock()
	defer database.DBMu.Unlock()
	query := `
		INSERT INTO posts (title, content, author_id, subreddit_id, karma, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, created_at, updated_at
	`
	err := r.DB.QueryRow(query, post.Title, post.Content, post.AuthorID, post.SubredditID).
		Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return fmt.Errorf("CreatePost: %v", err)
	}
	return nil
}

// GetPostByID retrieves a post by its ID.
func (r *postRepository) GetPostByID(id int) (*models.Post, error) {
	database.DBMu.Lock()
    defer database.DBMu.Unlock()
	query := `
		SELECT id, title, content, author_id, subreddit_id, karma, created_at, updated_at
		FROM posts
		WHERE id = $1
	`
	post := &models.Post{}
	err := r.DB.QueryRow(query, id).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.AuthorID,
		&post.SubredditID,
		&post.Karma,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("GetPostByID: post not found")
		}
		return nil, fmt.Errorf("GetPostByID: %v", err)
	}
	return post, nil
}

// GetPostsBySubreddit retrieves posts from a specific subreddit with pagination.
func (r *postRepository) GetPostsBySubreddit(subredditID int, limit, offset int) ([]*models.Post, error) {
	database.DBMu.Lock()
    defer database.DBMu.Unlock()
	query := `
		SELECT id, title, content, author_id, subreddit_id, karma, created_at, updated_at
		FROM posts
		WHERE subreddit_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.DB.Query(query, subredditID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("GetPostsBySubreddit: %v", err)
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post := &models.Post{}
		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.AuthorID,
			&post.SubredditID,
			&post.Karma,
			&post.CreatedAt,
			&post.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("GetPostsBySubreddit: %v", err)
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetPostsBySubreddit: %v", err)
	}

	return posts, nil
}

// GetFeedPosts retrieves posts for the feed with pagination.
func (r *postRepository) GetFeedPosts(limit, offset int) ([]*models.Post, error) {
	database.DBMu.Lock()
    defer database.DBMu.Unlock()
	query := `
		SELECT id, title, content, author_id, subreddit_id, karma, created_at, updated_at
		FROM posts
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.DB.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("GetFeedPosts: %v", err)
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post := &models.Post{}
		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.AuthorID,
			&post.SubredditID,
			&post.Karma,
			&post.CreatedAt,
			&post.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("GetFeedPosts: %v", err)
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetFeedPosts: %v", err)
	}

	return posts, nil
}

// UpdatePost updates an existing post's information.
func (r *postRepository) UpdatePost(post *models.Post) error {
	database.DBMu.Lock()
    defer database.DBMu.Unlock()
	query := `
		UPDATE posts
		SET title = $1, content = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
		RETURNING updated_at
	`
	err := r.DB.QueryRow(query, post.Title, post.Content, post.ID).
		Scan(&post.UpdatedAt)
	if err != nil {
		return fmt.Errorf("UpdatePost: %v", err)
	}
	return nil
}

// DeletePost removes a post from the database.
func (r *postRepository) DeletePost(id int) error {
	database.DBMu.Lock()
    defer database.DBMu.Unlock()
	query := `
		DELETE FROM posts
		WHERE id = $1
	`
	result, err := r.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("DeletePost: %v", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("DeletePost: %v", err)
	}
	if rowsAffected == 0 {
		return errors.New("DeletePost: no post found to delete")
	}
	return nil
}
