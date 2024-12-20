// File: internal/repository/user_repository.go

package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"redditclone/internal/models"
	"redditclone/pkg/database"
)

// UserRepository provides access to the users storage.
type UserRepository interface {
	CreateUser(user *models.User) error
	GetUserByID(id int) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(id int) error
}

type userRepository struct {
	DB *sql.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{DB: db}
}

// CreateUser inserts a new user into the database.
func (r *userRepository) CreateUser(user *models.User) error {
	database.DBMu.Lock()
	defer database.DBMu.Unlock()
	query := `
		INSERT INTO users (username, email, password, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, created_at, updated_at
	`
	err := r.DB.QueryRow(query, user.Username, user.Email, user.Password).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("CreateUser: %v", err)
	}
	return nil
}

// GetUserByID retrieves a user by their ID.
func (r *userRepository) GetUserByID(id int) (*models.User, error) {
	database.DBMu.Lock()
    defer database.DBMu.Unlock()
	query := `
		SELECT id, username, email, password, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	user := &models.User{}
	err := r.DB.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("GetUserByID: user not found")
		}
		return nil, fmt.Errorf("GetUserByID: %v", err)
	}
	return user, nil
}

// GetUserByUsername retrieves a user by their username.
func (r *userRepository) GetUserByUsername(username string) (*models.User, error) {
	database.DBMu.Lock()
    defer database.DBMu.Unlock()
	query := `
		SELECT id, username, email, password, created_at, updated_at
		FROM users
		WHERE username = $1
	`
	user := &models.User{}
	err := r.DB.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("GetUserByUsername: user not found")
		}
		return nil, fmt.Errorf("GetUserByUsername: %v", err)
	}
	return user, nil
}

// UpdateUser updates an existing user's information.
func (r *userRepository) UpdateUser(user *models.User) error {
	database.DBMu.Lock()
    defer database.DBMu.Unlock()
	query := `
		UPDATE users
		SET username = $1, email = $2, password = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
		RETURNING updated_at
	`
	err := r.DB.QueryRow(query, user.Username, user.Email, user.Password, user.ID).
		Scan(&user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("UpdateUser: %v", err)
	}
	return nil
}

// DeleteUser removes a user from the database.
func (r *userRepository) DeleteUser(id int) error {
	database.DBMu.Lock()
    defer database.DBMu.Unlock()
	query := `
		DELETE FROM users
		WHERE id = $1
	`
	result, err := r.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("DeleteUser: %v", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("DeleteUser: %v", err)
	}
	if rowsAffected == 0 {
		return errors.New("DeleteUser: no user found to delete")
	}
	return nil
}
