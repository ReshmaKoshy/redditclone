// File: internal/repository/comment_repository.go

package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"redditclone/internal/models"
	"redditclone/pkg/database"
)

// CommentRepository provides access to the comments storage.
type CommentRepository interface {
	CreateComment(comment *models.Comment) error
	GetCommentByID(id int) (*models.Comment, error)
	GetCommentsByPost(postID int, limit, offset int) ([]*models.Comment, error)
	GetReplies(parentID int, limit, offset int) ([]*models.Comment, error)
	UpdateComment(comment *models.Comment) error
	DeleteComment(id int) error
}

type commentRepository struct {
	DB *sql.DB
}

// NewCommentRepository creates a new CommentRepository.
func NewCommentRepository(db *sql.DB) CommentRepository {
	return &commentRepository{DB: db}
}

// CreateComment inserts a new comment into the database.
func (r *commentRepository) CreateComment(comment *models.Comment) error {
	database.DBMu.Lock()
	defer database.DBMu.Unlock()
	query := `
		INSERT INTO comments (content, author_id, post_id, parent_id, karma, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, created_at, updated_at
	`
	err := r.DB.QueryRow(query, comment.Content, comment.AuthorID, comment.PostID, comment.ParentID).
		Scan(&comment.ID, &comment.CreatedAt, &comment.UpdatedAt)
	if err != nil {
		return fmt.Errorf("CreateComment: %v", err)
	}
	return nil
}

// GetCommentByID retrieves a comment by its ID.
func (r *commentRepository) GetCommentByID(id int) (*models.Comment, error) {
	database.DBMu.Lock()
    defer database.DBMu.Unlock()
	query := `
		SELECT id, content, author_id, post_id, parent_id, karma, created_at, updated_at
		FROM comments
		WHERE id = $1
	`
	comment := &models.Comment{}
	err := r.DB.QueryRow(query, id).Scan(
		&comment.ID,
		&comment.Content,
		&comment.AuthorID,
		&comment.PostID,
		&comment.ParentID,
		&comment.Karma,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("GetCommentByID: comment not found")
		}
		return nil, fmt.Errorf("GetCommentByID: %v", err)
	}
	return comment, nil
}

// GetCommentsByPost retrieves comments for a specific post with pagination.
func (r *commentRepository) GetCommentsByPost(postID int, limit, offset int) ([]*models.Comment, error) {
	database.DBMu.Lock()
    defer database.DBMu.Unlock()
	query := `
		SELECT id, content, author_id, post_id, parent_id, karma, created_at, updated_at
		FROM comments
		WHERE post_id = $1 AND parent_id IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.DB.Query(query, postID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("GetCommentsByPost: %v", err)
	}
	defer rows.Close()

	var comments []*models.Comment
	for rows.Next() {
		comment := &models.Comment{}
		err := rows.Scan(
			&comment.ID,
			&comment.Content,
			&comment.AuthorID,
			&comment.PostID,
			&comment.ParentID,
			&comment.Karma,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("GetCommentsByPost: %v", err)
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetCommentsByPost: %v", err)
	}

	return comments, nil
}

// GetReplies retrieves replies to a specific comment with pagination.
func (r *commentRepository) GetReplies(parentID int, limit, offset int) ([]*models.Comment, error) {
	database.DBMu.Lock()
    defer database.DBMu.Unlock()
	query := `
		SELECT id, content, author_id, post_id, parent_id, karma, created_at, updated_at
		FROM comments
		WHERE parent_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.DB.Query(query, parentID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("GetReplies: %v", err)
	}
	defer rows.Close()

	var replies []*models.Comment
	for rows.Next() {
		comment := &models.Comment{}
		err := rows.Scan(
			&comment.ID,
			&comment.Content,
			&comment.AuthorID,
			&comment.PostID,
			&comment.ParentID,
			&comment.Karma,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("GetReplies: %v", err)
		}
		replies = append(replies, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetReplies: %v", err)
	}

	return replies, nil
}

// UpdateComment updates an existing comment's content.
func (r *commentRepository) UpdateComment(comment *models.Comment) error {
	database.DBMu.Lock()
    defer database.DBMu.Unlock()
	query := `
		UPDATE comments
		SET content = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
		RETURNING updated_at
	`
	err := r.DB.QueryRow(query, comment.Content, comment.ID).
		Scan(&comment.UpdatedAt)
	if err != nil {
		return fmt.Errorf("UpdateComment: %v", err)
	}
	return nil
}

// DeleteComment removes a comment from the database.
func (r *commentRepository) DeleteComment(id int) error {
	database.DBMu.Lock()
    defer database.DBMu.Unlock()
	query := `
		DELETE FROM comments
		WHERE id = $1
	`
	result, err := r.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("DeleteComment: %v", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("DeleteComment: %v", err)
	}
	if rowsAffected == 0 {
		return errors.New("DeleteComment: no comment found to delete")
	}
	return nil
}
