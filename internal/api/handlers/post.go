// File: internal/api/handlers/post.go

package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"redditclone/internal/models"
	"redditclone/internal/service"

	"github.com/gorilla/mux"
)

// PostHandler handles post-related HTTP requests.
type PostHandler struct {
	PostService service.PostService
}

// NewPostHandler creates a new PostHandler with the given PostService.
func NewPostHandler(postService service.PostService) *PostHandler {
	return &PostHandler{PostService: postService}
}

// CreatePost handles the creation of a new post.
func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subredditIDStr, ok := vars["id"] // subreddit ID from URL
	if !ok {
		http.Error(w, "Subreddit ID is required", http.StatusBadRequest)
		return
	}
	subredditID, err := strconv.Atoi(subredditIDStr)
	if err != nil {
		http.Error(w, "Invalid Subreddit ID", http.StatusBadRequest)
		return
	}

	var post models.Post
	// Decode the JSON request body into the Post model
	err = json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// TODO: Extract the author's user ID from the authenticated context
	// For simplicity, assume it's provided in the request body
	if post.AuthorID == 0 {
		http.Error(w, "Author ID is required", http.StatusBadRequest)
		return
	}

	post.SubredditID = subredditID

	// Create the post via the service
	err = h.PostService.CreatePost(&post)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Respond with the created post
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

// GetPost retrieves a post by ID.
func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid Post ID", http.StatusBadRequest)
		return
	}

	// Retrieve the post via the service
	post, err := h.PostService.GetPostByID(postID)
	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Respond with the post
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

// UpdatePost updates a post's information.
func (h *PostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid Post ID", http.StatusBadRequest)
		return
	}

	var post models.Post
	// Decode the JSON request body into the Post model
	err = json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	post.ID = postID

	// Update the post via the service
	err = h.PostService.UpdatePost(&post)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Post updated successfully"})
}

// DeletePost removes a post from the system.
func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid Post ID", http.StatusBadRequest)
		return
	}

	// Delete the post via the service
	err = h.PostService.DeletePost(postID)
	if err != nil {
		http.Error(w, "Post not found or could not be deleted", http.StatusNotFound)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Post deleted successfully"})
}

// GetFeed retrieves a list of posts for the feed with pagination.
func (h *PostHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters for pagination
	limit, offset := parsePaginationParams(r)

	// Retrieve the feed posts via the service
	posts, err := h.PostService.GetFeedPosts(limit, offset)
	if err != nil {
		http.Error(w, "Failed to retrieve feed", http.StatusInternalServerError)
		return
	}

	// Respond with the feed posts
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

// Helper function to parse pagination query parameters
func parsePaginationParams(r *http.Request) (limit, offset int) {
	// Default values
	limit = 10
	offset = 0

	// Parse 'limit' query parameter
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 && val <= 100 {
			limit = val
		}
	}

	// Parse 'offset' query parameter
	if o := r.URL.Query().Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}

	return
}
