// File: pkg/router/router.go

package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"redditclone/internal/api/handlers"
	"redditclone/internal/service"
)

// NewRouter initializes the HTTP router with all routes and handlers.
func NewRouter(
	userService service.UserService,
	subredditService service.SubredditService,
	postService service.PostService,
	commentService service.CommentService,
	voteService service.VoteService,
	messageService service.MessageService,
) http.Handler {
	r := mux.NewRouter()

	// Initialize handlers with the services
	userHandler := handlers.NewUserHandler(userService)
	subredditHandler := handlers.NewSubredditHandler(subredditService)
	postHandler := handlers.NewPostHandler(postService)
	commentHandler := handlers.NewCommentHandler(commentService)
	voteHandler := handlers.NewVoteHandler(voteService)
	messageHandler := handlers.NewMessageHandler(messageService)

	// Define API routes and associate them with handlers.

	// User routes
	r.HandleFunc("/register", userHandler.Register).Methods("POST")
	r.HandleFunc("/login", userHandler.Login).Methods("POST")
	r.HandleFunc("/users/{id}", userHandler.GetProfile).Methods("GET")
	r.HandleFunc("/users/{id}", userHandler.UpdateProfile).Methods("PUT")
	r.HandleFunc("/users/{id}", userHandler.DeleteUser).Methods("DELETE")

	// Subreddit routes
	r.HandleFunc("/subreddits", subredditHandler.CreateSubreddit).Methods("POST")
	r.HandleFunc("/subreddits/{id}", subredditHandler.GetSubreddit).Methods("GET")
	r.HandleFunc("/subreddits/{id}", subredditHandler.UpdateSubreddit).Methods("PUT")
	r.HandleFunc("/subreddits/{id}", subredditHandler.DeleteSubreddit).Methods("DELETE")
	r.HandleFunc("/subreddits/{id}/join", subredditHandler.JoinSubreddit).Methods("POST")
	r.HandleFunc("/subreddits/{id}/leave", subredditHandler.LeaveSubreddit).Methods("POST")

	// Post routes
	r.HandleFunc("/subreddits/{id}/posts", postHandler.CreatePost).Methods("POST")
	r.HandleFunc("/posts/{id}", postHandler.GetPost).Methods("GET")
	r.HandleFunc("/posts/{id}", postHandler.UpdatePost).Methods("PUT")
	r.HandleFunc("/posts/{id}", postHandler.DeletePost).Methods("DELETE")
	r.HandleFunc("/feed", postHandler.GetFeed).Methods("GET")

	// Comment routes
	r.HandleFunc("/posts/{id}/comments", commentHandler.AddComment).Methods("POST")
	r.HandleFunc("/comments/{id}/reply", commentHandler.ReplyToComment).Methods("POST")
	r.HandleFunc("/comments/{id}", commentHandler.GetComment).Methods("GET")
	r.HandleFunc("/comments/{id}", commentHandler.UpdateComment).Methods("PUT")
	r.HandleFunc("/comments/{id}", commentHandler.DeleteComment).Methods("DELETE")
	r.HandleFunc("/posts/{id}/comments", commentHandler.GetCommentsByPost).Methods("GET")
	r.HandleFunc("/comments/{id}/replies", commentHandler.GetReplies).Methods("GET")

	// Vote routes
	r.HandleFunc("/posts/{id}/vote", voteHandler.CastVote).Methods("POST")
	r.HandleFunc("/posts/{id}/vote", voteHandler.ChangeVote).Methods("PUT")
	r.HandleFunc("/posts/{id}/vote", voteHandler.RemoveVote).Methods("DELETE")
	r.HandleFunc("/comments/{id}/vote", voteHandler.CastVote).Methods("POST")
	r.HandleFunc("/comments/{id}/vote", voteHandler.ChangeVote).Methods("PUT")
	r.HandleFunc("/comments/{id}/vote", voteHandler.RemoveVote).Methods("DELETE")

	// Message routes
	r.HandleFunc("/messages", messageHandler.SendMessage).Methods("POST")
	r.HandleFunc("/messages/{id}/reply", messageHandler.ReplyToMessage).Methods("POST")
	r.HandleFunc("/messages/{id}", messageHandler.GetMessage).Methods("GET")
	r.HandleFunc("/messages/{id}", messageHandler.UpdateMessage).Methods("PUT")
	r.HandleFunc("/messages/{id}", messageHandler.DeleteMessage).Methods("DELETE")
	r.HandleFunc("/users/{id}/messages", messageHandler.GetMessagesForUser).Methods("GET")
	r.HandleFunc("/messages/{id}/replies", messageHandler.GetReplies).Methods("GET")

	// Add more routes as needed
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"Endpoint Not Found"}`, http.StatusNotFound)
	})
	
	return r
}
