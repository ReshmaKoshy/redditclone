// File: internal/api/handlers/subreddit.go

package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"redditclone/internal/models"
	"redditclone/internal/service"

	"github.com/gorilla/mux"
)

// SubredditHandler handles subreddit-related HTTP requests.
type SubredditHandler struct {
	SubredditService service.SubredditService
}

// NewSubredditHandler creates a new SubredditHandler with the given SubredditService.
func NewSubredditHandler(subredditService service.SubredditService) *SubredditHandler {
	return &SubredditHandler{SubredditService: subredditService}
}

// CreateSubreddit handles the creation of a new subreddit.
func (h *SubredditHandler) CreateSubreddit(w http.ResponseWriter, r *http.Request) {
	var subreddit models.Subreddit
	// Decode the JSON request body into the Subreddit model
	err := json.NewDecoder(r.Body).Decode(&subreddit)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// TODO: Extract the creator's user ID from the authenticated context
	// For simplicity, assume it's provided in the request body
	// In a real application, retrieve it from the authentication token
	if subreddit.CreatedBy == 0 {
		http.Error(w, "Creator ID is required", http.StatusBadRequest)
		return
	}

	// Create the subreddit via the service
	err = h.SubredditService.CreateSubreddit(&subreddit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(subreddit)
}

// GetSubreddit retrieves a subreddit by ID.
func (h *SubredditHandler) GetSubreddit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Subreddit ID is required", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Subreddit ID", http.StatusBadRequest)
		return
	}

	// Retrieve the subreddit via the service
	subreddit, err := h.SubredditService.GetSubredditByID(id)
	if err != nil {
		http.Error(w, "Subreddit not found", http.StatusNotFound)
		return
	}

	// Respond with the subreddit
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subreddit)
}

// UpdateSubreddit updates a subreddit's information.
func (h *SubredditHandler) UpdateSubreddit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Subreddit ID is required", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Subreddit ID", http.StatusBadRequest)
		return
	}

	var subreddit models.Subreddit
	// Decode the JSON request body into the Subreddit model
	err = json.NewDecoder(r.Body).Decode(&subreddit)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	subreddit.ID = id

	// Update the subreddit via the service
	err = h.SubredditService.UpdateSubreddit(&subreddit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Subreddit updated successfully"})
}

// DeleteSubreddit removes a subreddit from the system.
func (h *SubredditHandler) DeleteSubreddit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Subreddit ID is required", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Subreddit ID", http.StatusBadRequest)
		return
	}

	// Delete the subreddit via the service
	err = h.SubredditService.DeleteSubreddit(id)
	if err != nil {
		http.Error(w, "Subreddit not found or could not be deleted", http.StatusNotFound)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Subreddit deleted successfully"})
}

// JoinSubreddit allows a user to join a subreddit.
func (h *SubredditHandler) JoinSubreddit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"] // subreddit ID
	if !ok {
		http.Error(w, "Subreddit ID is required", http.StatusBadRequest)
		return
	}
	subredditID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Subreddit ID", http.StatusBadRequest)
		return
	}

	// TODO: Extract the user's ID from the authenticated context
	// For simplicity, assume it's provided in the request body
	var payload struct {
		UserID int `json:"user_id"`
	}
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil || payload.UserID == 0 {
		http.Error(w, "Invalid request payload: user_id is required", http.StatusBadRequest)
		return
	}

	// Join the subreddit via the service
	err = h.SubredditService.JoinSubreddit(payload.UserID, subredditID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Joined subreddit successfully"})
}

// LeaveSubreddit allows a user to leave a subreddit.
func (h *SubredditHandler) LeaveSubreddit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"] // subreddit ID
	if !ok {
		http.Error(w, "Subreddit ID is required", http.StatusBadRequest)
		return
	}
	subredditID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Subreddit ID", http.StatusBadRequest)
		return
	}

	// TODO: Extract the user's ID from the authenticated context
	// For simplicity, assume it's provided in the request body
	var payload struct {
		UserID int `json:"user_id"`
	}
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil || payload.UserID == 0 {
		http.Error(w, "Invalid request payload: user_id is required", http.StatusBadRequest)
		return
	}

	// Leave the subreddit via the service
	err = h.SubredditService.LeaveSubreddit(payload.UserID, subredditID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Left subreddit successfully"})
}
