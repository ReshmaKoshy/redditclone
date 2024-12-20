// File: internal/service/message_service.go

package service

import (
	"errors"

	"redditclone/internal/models"
	"redditclone/internal/repository"
)

// MessageService defines the methods for message-related business logic.
type MessageService interface {
	SendMessage(message *models.Message) error
	ReplyToMessage(message *models.Message) error
	GetMessageByID(id int) (*models.Message, error)
	GetMessagesForUser(userID int, limit, offset int) ([]*models.Message, error)
	GetReplies(parentID int, limit, offset int) ([]*models.Message, error)
	UpdateMessage(message *models.Message) error
	DeleteMessage(id int) error
}

type messageService struct {
	MessageRepo repository.MessageRepository
	UserRepo    repository.UserRepository
	// Add additional repositories if necessary
}

// NewMessageService creates a new MessageService.
func NewMessageService(messageRepo repository.MessageRepository, userRepo repository.UserRepository) MessageService {
	return &messageService{
		MessageRepo: messageRepo,
		UserRepo:    userRepo,
	}
}

// SendMessage handles sending a new direct message.
func (s *messageService) SendMessage(message *models.Message) error {
	// Validate input
	if message.Content == "" {
		return errors.New("SendMessage: content is required")
	}
	if message.SenderID == message.ReceiverID {
		return errors.New("SendMessage: sender and receiver cannot be the same")
	}

	// Check if sender exists
	_, err := s.UserRepo.GetUserByID(message.SenderID)
	if err != nil {
		return errors.New("SendMessage: sender does not exist")
	}

	// Check if receiver exists
	receiver, err := s.UserRepo.GetUserByID(message.ReceiverID)
	if err != nil {
		return errors.New("SendMessage: receiver does not exist")
	}
	_=receiver
	// Create the message via the repository
	err = s.MessageRepo.SendMessage(message)
	if err != nil {
		return err
	}

	return nil
}

// ReplyToMessage handles replying to an existing message.
func (s *messageService) ReplyToMessage(message *models.Message) error {
	// Validate input
	if message.Content == "" {
		return errors.New("ReplyToMessage: content is required")
	}
	if message.SenderID == message.ReceiverID {
		return errors.New("ReplyToMessage: sender and receiver cannot be the same")
	}

	// Check if sender exists
	_, err := s.UserRepo.GetUserByID(message.SenderID)
	if err != nil {
		return errors.New("ReplyToMessage: sender does not exist")
	}

	// Check if parent message exists
	parentMessage, err := s.MessageRepo.GetMessageByID(*message.ParentID)
	if err != nil {
		return errors.New("ReplyToMessage: parent message does not exist")
	}

	// The reply should be sent to the original sender of the parent message
	message.ReceiverID = parentMessage.SenderID

	// Create the reply via the repository
	err = s.MessageRepo.SendMessage(message)
	if err != nil {
		return err
	}

	return nil
}

// GetMessageByID retrieves a message by its ID.
func (s *messageService) GetMessageByID(id int) (*models.Message, error) {
	message, err := s.MessageRepo.GetMessageByID(id)
	if err != nil {
		return nil, err
	}
	return message, nil
}

// GetMessagesForUser retrieves direct messages received by a user with pagination.
func (s *messageService) GetMessagesForUser(userID int, limit, offset int) ([]*models.Message, error) {
	messages, err := s.MessageRepo.GetMessagesForUser(userID, limit, offset)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

// GetReplies retrieves replies to a specific message with pagination.
func (s *messageService) GetReplies(parentID int, limit, offset int) ([]*models.Message, error) {
	// Check if parent message exists
	_, err := s.MessageRepo.GetMessageByID(parentID)
	if err != nil {
		return nil, errors.New("GetReplies: parent message does not exist")
	}

	replies, err := s.MessageRepo.GetReplies(parentID, limit, offset)
	if err != nil {
		return nil, err
	}
	return replies, nil
}

// UpdateMessage updates a message's content.
func (s *messageService) UpdateMessage(message *models.Message) error {
	// Validate input
	if message.Content == "" {
		return errors.New("UpdateMessage: content is required")
	}

	// Check if message exists
	existingMessage, err := s.MessageRepo.GetMessageByID(message.ID)
	if err != nil {
		return errors.New("UpdateMessage: message does not exist")
	}

	// Optionally, check if the user is authorized to update the message (e.g., the sender)

	// Update fields
	existingMessage.Content = message.Content

	// Update the message via the repository
	err = s.MessageRepo.UpdateMessage(existingMessage)
	if err != nil {
		return err
	}

	return nil
}

// DeleteMessage removes a message from the system.
func (s *messageService) DeleteMessage(id int) error {
	// Optionally, check if the user is authorized to delete the message

	err := s.MessageRepo.DeleteMessage(id)
	if err != nil {
		return err
	}
	return nil
}
