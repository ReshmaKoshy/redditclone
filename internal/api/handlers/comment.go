// File: internal/api/handlers/comment.go

package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"redditclone/internal/models"
	"redditclone/internal/service"

	"github.com/gorilla/mux"
)

// CommentHandler handles comment-related HTTP requests.
type CommentHandler struct {
	CommentService service.CommentService
}

// NewCommentHandler creates a new CommentHandler with the given CommentService.
func NewCommentHandler(commentService service.CommentService) *CommentHandler {
	return &CommentHandler{CommentService: commentService}
}

// AddComment adds a new comment to a post.
func (h *CommentHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postIDStr, ok := vars["id"] // post ID from URL
	if !ok {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid Post ID", http.StatusBadRequest)
		return
	}

	var comment models.Comment
	// Decode the JSON request body into the Comment model
	err = json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// TODO: Extract the author's user ID from the authenticated context
	// For simplicity, assume it's provided in the request body
	if comment.AuthorID == 0 {
		http.Error(w, "Author ID is required", http.StatusBadRequest)
		return
	}

	comment.PostID = postID

	// Add the comment via the service
	err = h.CommentService.AddComment(&comment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Respond with the created comment
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

// ReplyToComment adds a reply to an existing comment.
func (h *CommentHandler) ReplyToComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	parentIDStr, ok := vars["id"] // parent comment ID from URL
	if !ok {
		http.Error(w, "Parent Comment ID is required", http.StatusBadRequest)
		return
	}
	parentID, err := strconv.Atoi(parentIDStr)
	if err != nil {
		http.Error(w, "Invalid Parent Comment ID", http.StatusBadRequest)
		return
	}

	var comment models.Comment
	// Decode the JSON request body into the Comment model
	err = json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// TODO: Extract the author's user ID from the authenticated context
	// For simplicity, assume it's provided in the request body
	if comment.AuthorID == 0 {
		http.Error(w, "Author ID is required", http.StatusBadRequest)
		return
	}

	comment.ParentID = &parentID

	// Reply to the comment via the service
	err = h.CommentService.ReplyToComment(&comment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Respond with the created reply
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

// GetComment retrieves a comment by ID.
func (h *CommentHandler) GetComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Comment ID is required", http.StatusBadRequest)
		return
	}
	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		http.Error(w, "Invalid Comment ID", http.StatusBadRequest)
		return
	}

	// Retrieve the comment via the service
	comment, err := h.CommentService.GetCommentByID(commentID)
	if err != nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	// Respond with the comment
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comment)
}

// UpdateComment updates a comment's content.
func (h *CommentHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Comment ID is required", http.StatusBadRequest)
		return
	}
	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		http.Error(w, "Invalid Comment ID", http.StatusBadRequest)
		return
	}

	var comment models.Comment
	// Decode the JSON request body into the Comment model
	err = json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	comment.ID = commentID

	// Update the comment via the service
	err = h.CommentService.UpdateComment(&comment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Comment updated successfully"})
}

// DeleteComment removes a comment from the system.
func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Comment ID is required", http.StatusBadRequest)
		return
	}
	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		http.Error(w, "Invalid Comment ID", http.StatusBadRequest)
		return
	}

	// Delete the comment via the service
	err = h.CommentService.DeleteComment(commentID)
	if err != nil {
		http.Error(w, "Comment not found or could not be deleted", http.StatusNotFound)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Comment deleted successfully"})
}

// GetCommentsByPost retrieves top-level comments for a specific post with pagination.
func (h *CommentHandler) GetCommentsByPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postIDStr, ok := vars["id"] // post ID from URL
	if !ok {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid Post ID", http.StatusBadRequest)
		return
	}

	// Parse pagination parameters
	limit, offset := parsePaginationParams(r)

	// Retrieve the comments via the service
	comments, err := h.CommentService.GetCommentsByPost(postID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to retrieve comments", http.StatusInternalServerError)
		return
	}

	// Respond with the comments
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

// GetReplies retrieves replies to a specific comment with pagination.
func (h *CommentHandler) GetReplies(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	parentIDStr, ok := vars["id"] // parent comment ID from URL
	if !ok {
		http.Error(w, "Parent Comment ID is required", http.StatusBadRequest)
		return
	}
	parentID, err := strconv.Atoi(parentIDStr)
	if err != nil {
		http.Error(w, "Invalid Parent Comment ID", http.StatusBadRequest)
		return
	}

	// Parse pagination parameters
	limit, offset := parsePaginationParams(r)

	// Retrieve the replies via the service
	replies, err := h.CommentService.GetReplies(parentID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to retrieve replies", http.StatusInternalServerError)
		return
	}

	// Respond with the replies
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(replies)
}
