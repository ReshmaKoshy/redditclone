// File: internal/repository/vote_repository.go

package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"redditclone/internal/models"
	"redditclone/pkg/database"
)

// VoteRepository provides access to the votes storage.
type VoteRepository interface {
	CreateVote(vote *models.Vote) error
	GetVoteByUserAndPost(userID, postID int) (*models.Vote, error)
	GetVoteByUserAndComment(userID, commentID int) (*models.Vote, error)
	UpdateVote(vote *models.Vote) error
	DeleteVote(vote *models.Vote) error
}

type voteRepository struct {
	DB *sql.DB
}

// NewVoteRepository creates a new VoteRepository.
func NewVoteRepository(db *sql.DB) VoteRepository {
	return &voteRepository{DB: db}
}

// CreateVote inserts a new vote into the database.
func (r *voteRepository) CreateVote(vote *models.Vote) error {
	database.DBMu.Lock()
	defer database.DBMu.Unlock()
	query := `
		INSERT INTO votes (user_id, post_id, comment_id, vote_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, created_at, updated_at
	`
	err := r.DB.QueryRow(query, vote.UserID, vote.PostID, vote.CommentID, vote.VoteType).
		Scan(&vote.ID, &vote.CreatedAt, &vote.UpdatedAt)
	if err != nil {
		return fmt.Errorf("CreateVote: %v", err)
	}
	return nil
}

// GetVoteByUserAndPost retrieves a vote by a user on a specific post.
func (r *voteRepository) GetVoteByUserAndPost(userID, postID int) (*models.Vote, error) {
	database.DBMu.Lock()
    defer database.DBMu.Unlock()
	query := `
		SELECT id, user_id, post_id, comment_id, vote_type, created_at, updated_at
		FROM votes
		WHERE user_id = $1 AND post_id = $2 AND comment_id IS NULL
	`
	vote := &models.Vote{}
	err := r.DB.QueryRow(query, userID, postID).Scan(
		&vote.ID,
		&vote.UserID,
		&vote.PostID,
		&vote.CommentID,
		&vote.VoteType,
		&vote.CreatedAt,
		&vote.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("GetVoteByUserAndPost: vote not found")
		}
		return nil, fmt.Errorf("GetVoteByUserAndPost: %v", err)
	}
	return vote, nil
}

// GetVoteByUserAndComment retrieves a vote by a user on a specific comment.
func (r *voteRepository) GetVoteByUserAndComment(userID, commentID int) (*models.Vote, error) {
	database.DBMu.Lock()
    defer database.DBMu.Unlock()
	query := `
		SELECT id, user_id, post_id, comment_id, vote_type, created_at, updated_at
		FROM votes
		WHERE user_id = $1 AND comment_id = $2 AND post_id IS NULL
	`
	vote := &models.Vote{}
	err := r.DB.QueryRow(query, userID, commentID).Scan(
		&vote.ID,
		&vote.UserID,
		&vote.PostID,
		&vote.CommentID,
		&vote.VoteType,
		&vote.CreatedAt,
		&vote.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("GetVoteByUserAndComment: vote not found")
		}
		return nil, fmt.Errorf("GetVoteByUserAndComment: %v", err)
	}
	return vote, nil
}

// UpdateVote updates an existing vote's type.
func (r *voteRepository) UpdateVote(vote *models.Vote) error {
	database.DBMu.Lock()
    defer database.DBMu.Unlock()
	query := `
		UPDATE votes
		SET vote_type = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
		RETURNING updated_at
	`
	err := r.DB.QueryRow(query, vote.VoteType, vote.ID).
		Scan(&vote.UpdatedAt)
	if err != nil {
		return fmt.Errorf("UpdateVote: %v", err)
	}
	return nil
}

// DeleteVote removes a vote from the database.
func (r *voteRepository) DeleteVote(vote *models.Vote) error {
	database.DBMu.Lock()
    defer database.DBMu.Unlock()
	query := `
		DELETE FROM votes
		WHERE id = $1
	`
	result, err := r.DB.Exec(query, vote.ID)
	if err != nil {
		return fmt.Errorf("DeleteVote: %v", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("DeleteVote: %v", err)
	}
	if rowsAffected == 0 {
		return errors.New("DeleteVote: no vote found to delete")
	}
	return nil
}
