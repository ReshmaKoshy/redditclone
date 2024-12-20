// File: internal/repository/message_repository.go

package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"redditclone/internal/models"
	"redditclone/pkg/database"
)

// MessageRepository provides access to the messages storage.
type MessageRepository interface {
	SendMessage(message *models.Message) error
	GetMessageByID(id int) (*models.Message, error)
	GetMessagesForUser(userID int, limit, offset int) ([]*models.Message, error)
	GetReplies(parentID int, limit, offset int) ([]*models.Message, error)
	UpdateMessage(message *models.Message) error
	DeleteMessage(id int) error
}

type messageRepository struct {
	DB *sql.DB
}

// NewMessageRepository creates a new MessageRepository.
func NewMessageRepository(db *sql.DB) MessageRepository {
	return &messageRepository{DB: db}
}

// SendMessage inserts a new message into the database.
func (r *messageRepository) SendMessage(message *models.Message) error {
	database.DBMu.Lock()
	defer database.DBMu.Unlock()
	query := `
		INSERT INTO messages (sender_id, receiver_id, content, parent_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, created_at, updated_at
	`
	err := r.DB.QueryRow(query, message.SenderID, message.ReceiverID, message.Content, message.ParentID).
		Scan(&message.ID, &message.CreatedAt, &message.UpdatedAt)
	if err != nil {
		return fmt.Errorf("SendMessage: %v", err)
	}
	return nil
}

// GetMessageByID retrieves a message by its ID.
func (r *messageRepository) GetMessageByID(id int) (*models.Message, error) {
	database.DBMu.Lock()
    defer database.DBMu.Unlock()
	query := `
		SELECT id, sender_id, receiver_id, content, parent_id, created_at, updated_at
		FROM messages
		WHERE id = $1
	`
	message := &models.Message{}
	err := r.DB.QueryRow(query, id).Scan(
		&message.ID,
		&message.SenderID,
		&message.ReceiverID,
		&message.Content,
		&message.ParentID,
		&message.CreatedAt,
		&message.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("GetMessageByID: message not found")
		}
		return nil, fmt.Errorf("GetMessageByID: %v", err)
	}
	return message, nil
}

// GetMessagesForUser retrieves direct messages for a user with pagination.
func (r *messageRepository) GetMessagesForUser(userID int, limit, offset int) ([]*models.Message, error) {
	database.DBMu.Lock()
    defer database.DBMu.Unlock()
	query := `
		SELECT id, sender_id, receiver_id, content, parent_id, created_at, updated_at
		FROM messages
		WHERE receiver_id = $1 AND parent_id IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.DB.Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("GetMessagesForUser: %v", err)
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		message := &models.Message{}
		err := rows.Scan(
			&message.ID,
			&message.SenderID,
			&message.ReceiverID,
			&message.Content,
			&message.ParentID,
			&message.CreatedAt,
			&message.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("GetMessagesForUser: %v", err)
		}
		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetMessagesForUser: %v", err)
	}

	return messages, nil
}

// GetReplies retrieves replies to a specific message with pagination.
func (r *messageRepository) GetReplies(parentID int, limit, offset int) ([]*models.Message, error) {
	database.DBMu.Lock()
    defer database.DBMu.Unlock()
	query := `
		SELECT id, sender_id, receiver_id, content, parent_id, created_at, updated_at
		FROM messages
		WHERE parent_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.DB.Query(query, parentID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("GetReplies: %v", err)
	}
	defer rows.Close()

	var replies []*models.Message
	for rows.Next() {
		message := &models.Message{}
		err := rows.Scan(
			&message.ID,
			&message.SenderID,
			&message.ReceiverID,
			&message.Content,
			&message.ParentID,
			&message.CreatedAt,
			&message.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("GetReplies: %v", err)
		}
		replies = append(replies, message)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetReplies: %v", err)
	}

	return replies, nil
}

// UpdateMessage updates an existing message's content.
func (r *messageRepository) UpdateMessage(message *models.Message) error {
	database.DBMu.Lock()
    defer database.DBMu.Unlock()
	query := `
		UPDATE messages
		SET content = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
		RETURNING updated_at
	`
	err := r.DB.QueryRow(query, message.Content, message.ID).
		Scan(&message.UpdatedAt)
	if err != nil {
		return fmt.Errorf("UpdateMessage: %v", err)
	}
	return nil
}

// DeleteMessage removes a message from the database.
func (r *messageRepository) DeleteMessage(id int) error {
	database.DBMu.Lock()
    defer database.DBMu.Unlock()
	query := `
		DELETE FROM messages
		WHERE id = $1
	`
	result, err := r.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("DeleteMessage: %v", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("DeleteMessage: %v", err)
	}
	if rowsAffected == 0 {
		return errors.New("DeleteMessage: no message found to delete")
	}
	return nil
}
