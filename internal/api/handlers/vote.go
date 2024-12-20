// File: internal/api/handlers/vote.go

package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"redditclone/internal/models"
	"redditclone/internal/service"

	"github.com/gorilla/mux"
)

// VoteHandler handles vote-related HTTP requests.
type VoteHandler struct {
	VoteService service.VoteService
}

// NewVoteHandler creates a new VoteHandler with the given VoteService.
func NewVoteHandler(voteService service.VoteService) *VoteHandler {
	return &VoteHandler{VoteService: voteService}
}

// CastVote handles casting a new vote on a post or comment.
func (h *VoteHandler) CastVote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	targetIDStr, ok := vars["id"] // post or comment ID from URL
	if !ok {
		http.Error(w, "Target ID is required", http.StatusBadRequest)
		return
	}
	_, err := strconv.Atoi(targetIDStr)
	if err != nil {
		http.Error(w, "Invalid Target ID", http.StatusBadRequest)
		return
	}

	var vote models.Vote
	// Decode the JSON request body into the Vote model

	err = json.NewDecoder(r.Body).Decode(&vote)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// TODO: Extract the user's ID from the authenticated context
	// For simplicity, assume it's provided in the request body
	if vote.UserID == 0 {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Determine if the vote is on a post or comment based on the URL or request body
	// For simplicity, assume the client specifies the target type
	// var payload struct {
	// 	VoteType  string `json:"vote_type"` // "upvote" or "downvote"
	// 	PostID    int    `json:"post_id,omitempty"`
	// 	CommentID int    `json:"comment_id,omitempty"`
	// }
	
	// err = json.NewDecoder(r.Body).Decode(&payload)
	// if err != nil {
	// 	http.Error(w, "Invalid request payload to payload", http.StatusBadRequest)
	// 	return
	// }

	// Set the vote type and target
	// if payload.VoteType == "upvote" || payload.VoteType == "downvote" {
	// 	vote.VoteType = payload.VoteType //models.VoteType(payload.VoteType)
	// } else {
	// 	http.Error(w, "Invalid vote type", http.StatusBadRequest)
	// 	return
	// }

	// if payload.PostID != 0 {
	// 	vote.PostID = &payload.PostID
	// } else if payload.CommentID != 0 {
	// 	vote.CommentID = &payload.CommentID
	// } else {
	// 	http.Error(w, "Either post_id or comment_id must be provided", http.StatusBadRequest)
	// 	return
	// }

	// Cast the vote via the service
	err = h.VoteService.CastVote(&vote)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Vote cast successfully"})
}

// ChangeVote handles changing an existing vote on a post or comment.
func (h *VoteHandler) ChangeVote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	targetIDStr, ok := vars["id"] // post or comment ID from URL
	if !ok {
		http.Error(w, "Target ID is required", http.StatusBadRequest)
		return
	}
	_, err := strconv.Atoi(targetIDStr)
	if err != nil {
		http.Error(w, "Invalid Target ID", http.StatusBadRequest)
		return
	}

	var vote models.Vote
	// Decode the JSON request body into the Vote model
	err = json.NewDecoder(r.Body).Decode(&vote)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// TODO: Extract the user's ID from the authenticated context
	// For simplicity, assume it's provided in the request body
	if vote.UserID == 0 {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Determine if the vote is on a post or comment based on the URL or request body
	var payload struct {
		VoteType  string `json:"vote_type"` // "upvote" or "downvote"
		PostID    int    `json:"post_id,omitempty"`
		CommentID int    `json:"comment_id,omitempty"`
	}
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Set the vote type and target
	if payload.VoteType == "upvote" || payload.VoteType == "downvote" {
		vote.VoteType = payload.VoteType //models.VoteType(payload.VoteType)
	} else {
		http.Error(w, "Invalid vote type", http.StatusBadRequest)
		return
	}

	if payload.PostID != 0 {
		vote.PostID = &payload.PostID
	} else if payload.CommentID != 0 {
		vote.CommentID = &payload.CommentID
	} else {
		http.Error(w, "Either post_id or comment_id must be provided", http.StatusBadRequest)
		return
	}

	// Change the vote via the service
	err = h.VoteService.ChangeVote(&vote)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Vote changed successfully"})
}

// RemoveVote handles removing an existing vote from a post or comment.
func (h *VoteHandler) RemoveVote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	targetIDStr, ok := vars["id"] // post or comment ID from URL
	if !ok {
		http.Error(w, "Target ID is required", http.StatusBadRequest)
		return
	}
	_, err := strconv.Atoi(targetIDStr)
	if err != nil {
		http.Error(w, "Invalid Target ID", http.StatusBadRequest)
		return
	}

	// TODO: Extract the user's ID from the authenticated context
	// For simplicity, assume it's provided in the request body
	var payload struct {
		UserID    int `json:"user_id"`
		PostID    int `json:"post_id,omitempty"`
		CommentID int `json:"comment_id,omitempty"`
	}
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil || payload.UserID == 0 {
		http.Error(w, "Invalid request payload: user_id is required", http.StatusBadRequest)
		return
	}

	// Determine if the vote is on a post or comment
	if payload.PostID != 0 {
		err = h.VoteService.RemoveVote(payload.UserID, payload.PostID, 0)
	} else if payload.CommentID != 0 {
		err = h.VoteService.RemoveVote(payload.UserID, 0, payload.CommentID)
	} else {
		http.Error(w, "Either post_id or comment_id must be provided", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Vote removed successfully"})
}
