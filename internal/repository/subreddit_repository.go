// File: internal/repository/subreddit_repository.go

package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"redditclone/internal/models"
	"redditclone/pkg/database"
)

// SubredditRepository provides access to the subreddits storage.
type SubredditRepository interface {
	CreateSubreddit(subreddit *models.Subreddit) error
	GetSubredditByID(id int) (*models.Subreddit, error)
	GetSubredditByName(name string) (*models.Subreddit, error)
	UpdateSubreddit(subreddit *models.Subreddit) error
	DeleteSubreddit(id int) error
}

type subredditRepository struct {
	DB *sql.DB
}

// NewSubredditRepository creates a new SubredditRepository.
func NewSubredditRepository(db *sql.DB) SubredditRepository {
	return &subredditRepository{DB: db}
}

// CreateSubreddit inserts a new subreddit into the database.
func (r *subredditRepository) CreateSubreddit(subreddit *models.Subreddit) error {
	database.DBMu.Lock()
	defer database.DBMu.Unlock()
	query := `
		INSERT INTO subreddits (name, description, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, created_at, updated_at
	`
	err := r.DB.QueryRow(query, subreddit.Name, subreddit.Description, subreddit.CreatedBy).
		Scan(&subreddit.ID, &subreddit.CreatedAt, &subreddit.UpdatedAt)
	if err != nil {
		return fmt.Errorf("CreateSubreddit: %v", err)
	}
	return nil
}

// GetSubredditByID retrieves a subreddit by its ID.
func (r *subredditRepository) GetSubredditByID(id int) (*models.Subreddit, error) {
	database.DBMu.Lock()
    defer database.DBMu.Unlock()
	query := `
		SELECT id, name, description, created_by, created_at, updated_at
		FROM subreddits
		WHERE id = $1
	`
	subreddit := &models.Subreddit{}
	err := r.DB.QueryRow(query, id).Scan(
		&subreddit.ID,
		&subreddit.Name,
		&subreddit.Description,
		&subreddit.CreatedBy,
		&subreddit.CreatedAt,
		&subreddit.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("GetSubredditByID: subreddit not found")
		}
		return nil, fmt.Errorf("GetSubredditByID: %v", err)
	}
	return subreddit, nil
}

// GetSubredditByName retrieves a subreddit by its unique name.
func (r *subredditRepository) GetSubredditByName(name string) (*models.Subreddit, error) {
	database.DBMu.Lock()
    defer database.DBMu.Unlock()
	query := `
		SELECT id, name, description, created_by, created_at, updated_at
		FROM subreddits
		WHERE name = $1
	`
	subreddit := &models.Subreddit{}
	err := r.DB.QueryRow(query, name).Scan(
		&subreddit.ID,
		&subreddit.Name,
		&subreddit.Description,
		&subreddit.CreatedBy,
		&subreddit.CreatedAt,
		&subreddit.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("GetSubredditByName: subreddit not found")
		}
		return nil, fmt.Errorf("GetSubredditByName: %v", err)
	}
	return subreddit, nil
}

// UpdateSubreddit updates an existing subreddit's information.
func (r *subredditRepository) UpdateSubreddit(subreddit *models.Subreddit) error {
	database.DBMu.Lock()
    defer database.DBMu.Unlock()
	query := `
		UPDATE subreddits
		SET name = $1, description = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
		RETURNING updated_at
	`
	err := r.DB.QueryRow(query, subreddit.Name, subreddit.Description, subreddit.ID).
		Scan(&subreddit.UpdatedAt)
	if err != nil {
		return fmt.Errorf("UpdateSubreddit: %v", err)
	}
	return nil
}

// DeleteSubreddit removes a subreddit from the database.
func (r *subredditRepository) DeleteSubreddit(id int) error {
	database.DBMu.Lock()
    defer database.DBMu.Unlock()
	query := `
		DELETE FROM subreddits
		WHERE id = $1
	`
	result, err := r.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("DeleteSubreddit: %v", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("DeleteSubreddit: %v", err)
	}
	if rowsAffected == 0 {
		return errors.New("DeleteSubreddit: no subreddit found to delete")
	}
	return nil
}
