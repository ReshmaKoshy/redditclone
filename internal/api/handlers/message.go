// File: internal/api/handlers/message.go

package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"redditclone/internal/models"
	"redditclone/internal/service"

	"github.com/gorilla/mux"
)

// MessageHandler handles message-related HTTP requests.
type MessageHandler struct {
	MessageService service.MessageService
}

// NewMessageHandler creates a new MessageHandler with the given MessageService.
func NewMessageHandler(messageService service.MessageService) *MessageHandler {
	return &MessageHandler{MessageService: messageService}
}

// SendMessage handles sending a new direct message.
func (h *MessageHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	var message models.Message
	// Decode the JSON request body into the Message model
	err := json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Send the message via the service
	err = h.MessageService.SendMessage(&message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Respond with the created message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(message)
}

// ReplyToMessage handles replying to an existing message.
func (h *MessageHandler) ReplyToMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	parentIDStr, ok := vars["id"] // parent message ID from URL
	if !ok {
		http.Error(w, "Parent Message ID is required", http.StatusBadRequest)
		return
	}
	parentID, err := strconv.Atoi(parentIDStr)
	if err != nil {
		http.Error(w, "Invalid Parent Message ID", http.StatusBadRequest)
		return
	}

	var message models.Message
	// Decode the JSON request body into the Message model
	err = json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	message.ParentID = &parentID

	// Reply to the message via the service
	err = h.MessageService.ReplyToMessage(&message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Respond with the created reply
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(message)
}

// GetMessage retrieves a message by ID.
func (h *MessageHandler) GetMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	messageIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Message ID is required", http.StatusBadRequest)
		return
	}
	messageID, err := strconv.Atoi(messageIDStr)
	if err != nil {
		http.Error(w, "Invalid Message ID", http.StatusBadRequest)
		return
	}

	// Retrieve the message via the service
	message, err := h.MessageService.GetMessageByID(messageID)
	if err != nil {
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	// Respond with the message
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(message)
}

// UpdateMessage updates a message's content.
func (h *MessageHandler) UpdateMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	messageIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Message ID is required", http.StatusBadRequest)
		return
	}
	messageID, err := strconv.Atoi(messageIDStr)
	if err != nil {
		http.Error(w, "Invalid Message ID", http.StatusBadRequest)
		return
	}

	var message models.Message
	// Decode the JSON request body into the Message model
	err = json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	message.ID = messageID

	// Update the message via the service
	err = h.MessageService.UpdateMessage(&message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Message updated successfully"})
}

// DeleteMessage removes a message from the system.
func (h *MessageHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	messageIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Message ID is required", http.StatusBadRequest)
		return
	}
	messageID, err := strconv.Atoi(messageIDStr)
	if err != nil {
		http.Error(w, "Invalid Message ID", http.StatusBadRequest)
		return
	}

	// Delete the message via the service
	err = h.MessageService.DeleteMessage(messageID)
	if err != nil {
		http.Error(w, "Message not found or could not be deleted", http.StatusNotFound)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Message deleted successfully"})
}

// GetMessagesForUser retrieves direct messages received by a user with pagination.
func (h *MessageHandler) GetMessagesForUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr, ok := vars["id"] // user ID from URL
	if !ok {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid User ID", http.StatusBadRequest)
		return
	}

	// Parse pagination parameters
	limit, offset := parsePaginationParams(r)

	// Retrieve the messages via the service
	messages, err := h.MessageService.GetMessagesForUser(userID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to retrieve messages", http.StatusInternalServerError)
		return
	}

	// Respond with the messages
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

// GetReplies retrieves replies to a specific message with pagination.
func (h *MessageHandler) GetReplies(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	parentIDStr, ok := vars["id"] // parent message ID from URL
	if !ok {
		http.Error(w, "Parent Message ID is required", http.StatusBadRequest)
		return
	}
	parentID, err := strconv.Atoi(parentIDStr)
	if err != nil {
		http.Error(w, "Invalid Parent Message ID", http.StatusBadRequest)
		return
	}

	// Parse pagination parameters
	limit, offset := parsePaginationParams(r)

	// Retrieve the replies via the service
	replies, err := h.MessageService.GetReplies(parentID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to retrieve replies", http.StatusInternalServerError)
		return
	}

	// Respond with the replies
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(replies)
}
