// File: internal/models/message.go

package models

import "time"

// Message represents a direct message between users.
type Message struct {
	ID         int       `json:"id"`                  // Unique identifier for the message.
	SenderID   int       `json:"sender_id"`           // ID of the user sending the message.
	ReceiverID int       `json:"receiver_id"`         // ID of the user receiving the message.
	Content    string    `json:"content"`             // Text content of the message.
	ParentID   *int      `json:"parent_id,omitempty"` // ID of the parent message, if it's a reply.
	CreatedAt  time.Time `json:"created_at"`          // Timestamp of when the message was sent.
	UpdatedAt  time.Time `json:"updated_at"`          // Timestamp of the last update to the message.
}
