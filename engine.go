package main

import (
	"database/sql"
	"log"
	"net/http"
	"redditclone/database" // Replace with your actual module name
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type DatabaseManager struct {
	db *sql.DB
	mu sync.RWMutex
}

// Structs for database operations
type User struct {
	ID       string
	Username string
	Karma    int
}

type Post struct {
	ID             int
	Title          string
	Content        string
	AuthorID       int    `json:"author_id"`
	AuthorUsername string `json:"author_name"`
	SubredditID    int    `json:"subreddit_id"`
	SubredditName  string `json:"subreddit_name"`
	CreatedAt      time.Time
	VoteCount      struct {
		Upvotes   int `json:"upvotes"`
		Downvotes int `json:"downvotes"`
	} `json:"vote_count"`
}

type DirectMessage struct {
	ID           int
	FromUserID   int
	FromUsername string
	Content      string
	CreatedAt    time.Time
}

// Request/Response structs
type RegisterUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type CreateSubredditRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type CreatePostRequest struct {
	Title       string `json:"title" binding:"required"`
	Content     string `json:"content" binding:"required"`
	SubredditID int    `json:"subreddit_id" binding:"required"`
}

type CreateCommentRequest struct {
	Content         string `json:"content" binding:"required"`
	PostID          int    `json:"post_id" binding:"required"`
	ParentCommentID *int   `json:"parent_comment_id"`
}

type VoteRequest struct {
	TargetID   int    `json:"target_id" binding:"required"`
	TargetType string `json:"target_type" binding:"required,oneof=post comment"`
	Value      int    `json:"value" binding:"required,oneof=-1 1"`
}

type SendMessageRequest struct {
	ToUserID int    `json:"to_user_id" binding:"required"`
	Content  string `json:"content" binding:"required"`
}

// API handler struct
type APIHandler struct {
	db *database.DatabaseManager
}

type PostWithDetails struct {
	Post
	Votes     int       `json:"votes"`
	UserVote  *int      `json:"user_vote"` // -1, 1, or null if no vote
	Comments  []Comment `json:"comments"`
	VoteCount struct {
		Upvotes   int `json:"upvotes"`
		Downvotes int `json:"downvotes"`
	} `json:"vote_count"`
}

type Comment struct {
	ID              int       `json:"id"`
	Content         string    `json:"content"`
	AuthorID        int       `json:"author_id"`
	AuthorUsername  string    `json:"author_username"`
	PostID          int       `json:"post_id"`
	ParentCommentID *int      `json:"parent_comment_id"`
	CreatedAt       time.Time `json:"created_at"`
	Votes           int       `json:"votes"`
	UserVote        *int      `json:"user_vote"` // -1, 1, or null if no vote
}

type TopUser struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	Karma        int    `json:"karma"`
	PostCount    int    `json:"post_count"`
	CommentCount int    `json:"comment_count"`
}

func NewAPIHandler(dbPath string) (*APIHandler, error) {
	dbManager, err := database.InitDatabase(dbPath)
	if err != nil {
		return nil, err
	}
	return &APIHandler{db: dbManager}, nil
}

// Middleware
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// In a real application, implement proper authentication
		// For now, we'll use a simple user_id header
		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID required"})
			c.Abort()
			return
		}
		c.Set("user_id", userID)
		c.Next()
	}
}

func (dm *DatabaseManager) GetTopUsers(limit int) ([]TopUser, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	query := `
        SELECT 
            u.id,
            u.username,
            u.karma,
            (SELECT COUNT(*) FROM posts WHERE author_id = u.id) as post_count,
            (SELECT COUNT(*) FROM comments WHERE author_id = u.id) as comment_count
        FROM users u
        ORDER BY u.karma DESC
        LIMIT ?
    `

	rows, err := dm.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []TopUser
	for rows.Next() {
		var user TopUser
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Karma,
			&user.PostCount,
			&user.CommentCount,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// Route handlers
func (h *APIHandler) registerUser(c *gin.Context) {
	var req RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := h.db.RegisterUser(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user_id":  userID,
		"username": req.Username,
	})
}

func (h *APIHandler) getUserByUsername(c *gin.Context) {
	username := c.Param("username")
	user, err := h.db.GetUserByUsername(username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *APIHandler) createSubreddit(c *gin.Context) {
	var req CreateSubredditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := strconv.Atoi(c.GetString("user_id"))
	subredditID, err := h.db.CreateSubreddit(req.Name, req.Description, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"subreddit_id": subredditID,
		"name":         req.Name,
	})
}

func (h *APIHandler) joinSubreddit(c *gin.Context) {
	subredditID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subreddit ID"})
		return
	}

	userID, _ := strconv.Atoi(c.GetString("user_id"))
	err = h.db.JoinSubreddit(userID, subredditID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully joined subreddit"})
}

func (h *APIHandler) leaveSubreddit(c *gin.Context) {
	subredditID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subreddit ID"})
		return
	}

	userID, _ := strconv.Atoi(c.GetString("user_id"))
	err = h.db.LeaveSubreddit(userID, subredditID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully left subreddit"})
}

func (h *APIHandler) createPost(c *gin.Context) {
	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := strconv.Atoi(c.GetString("user_id"))
	postID, err := h.db.CreatePost(req.Title, req.Content, userID, req.SubredditID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"post_id": postID,
		"title":   req.Title,
	})
}

func (h *APIHandler) getFeed(c *gin.Context) {
	userID, _ := strconv.Atoi(c.GetString("user_id"))
	posts, err := h.db.GetFeed(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, posts)
}

func (h *APIHandler) vote(c *gin.Context) {
	var req VoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := strconv.Atoi(c.GetString("user_id"))
	err := h.db.Vote(userID, req.TargetID, req.TargetType, req.Value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vote recorded successfully"})
}

func (h *APIHandler) createComment(c *gin.Context) {
	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := strconv.Atoi(c.GetString("user_id"))
	commentID, err := h.db.CreateComment(req.Content, userID, req.PostID, req.ParentCommentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"comment_id": commentID,
		"content":    req.Content,
	})
}

func (h *APIHandler) sendDirectMessage(c *gin.Context) {
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := strconv.Atoi(c.GetString("user_id"))
	messageID, err := h.db.SendDirectMessage(userID, req.ToUserID, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message_id": messageID,
		"content":    req.Content,
	})
}

func (h *APIHandler) resetDatabase(c *gin.Context) {

	err := h.db.ResetDatabase()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Database reset successfully"})
}

func (h *APIHandler) getDirectMessages(c *gin.Context) {
	userID, _ := strconv.Atoi(c.GetString("user_id"))
	messages, err := h.db.GetDirectMessages(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}
func (h *APIHandler) getTopUsers(c *gin.Context) {
	limit := 10 // Default limit
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	users, err := h.db.GetTopUsers(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

func main() {
	handler, err := NewAPIHandler("reddit_clone.db")
	if err != nil {
		log.Fatalf("Failed to initialize API handler: %v", err)
	}
	defer handler.db.Close()

	r := gin.Default()

	// Public routes
	r.POST("/register", handler.registerUser)
	r.GET("/users/:username", handler.getUserByUsername)

	// Protected routes
	authorized := r.Group("/")
	authorized.Use(authMiddleware())
	{
		// Subreddit routes
		authorized.POST("/subreddits", handler.createSubreddit)
		authorized.POST("/subreddits/:id/join", handler.joinSubreddit)
		authorized.POST("/subreddits/:id/leave", handler.leaveSubreddit)

		// Post routes
		authorized.POST("/posts", handler.createPost)
		authorized.GET("/feed", handler.getFeed)

		// Voting routes
		authorized.POST("/vote", handler.vote)

		// Comment routes
		authorized.POST("/comments", handler.createComment)

		// Direct message routes
		authorized.POST("/message", handler.sendDirectMessage)
		authorized.GET("/message", handler.getDirectMessages)
		authorized.GET("/users/top", handler.getTopUsers)

	}

	r.Run(":8080")
}
