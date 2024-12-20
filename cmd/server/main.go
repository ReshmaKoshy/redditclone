// File: cmd/server/main.go

package main

import (
	"log"
	"net/http"
	"os"

	"redditclone/internal/repository"
	"redditclone/internal/service"
	"redditclone/pkg/database"
	"redditclone/pkg/router"
)

func main() {
	// Retrieve the server port from environment variables or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize the database connection.
	// The Connect function should establish a connection to your chosen database
	// and return a *sql.DB instance along with any error encountered.
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	defer func() {
		if cerr := db.Close(); cerr != nil {
			log.Printf("Error closing the database connection: %v", cerr)
		}
	}()

	if err := database.InitializeSchema(db); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	  }

	// Initialize the HTTP router.
	// The NewRouter function should set up all the necessary routes and return an http.Handler.
	userRepo := repository.NewUserRepository(db)
	subredditRepo := repository.NewSubredditRepository(db)
	postRepo := repository.NewPostRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	voteRepo := repository.NewVoteRepository(db)
	messageRepo := repository.NewMessageRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepo)
	subredditService := service.NewSubredditService(subredditRepo)
	postService := service.NewPostService(postRepo, subredditRepo, userRepo)
	commentService := service.NewCommentService(commentRepo, postRepo, userRepo, subredditRepo)
	voteService := service.NewVoteService(voteRepo, postRepo, commentRepo)
	messageService := service.NewMessageService(messageRepo, userRepo)

	// Initialize the HTTP router with services
	r := router.NewRouter(userService, subredditService, postService, commentService, voteService, messageService)


	// Define the server address.
	addr := ":" + port

	// Log the server start-up.
	log.Printf("Starting server on port %s", port)

	// Start the HTTP server.
	// http.ListenAndServe listens on the TCP network address addr and then calls Serve with handler to handle requests on incoming connections.
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
