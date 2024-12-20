// File: internal/service/post_service.go

package service

import (
	"errors"

	"redditclone/internal/models"
	"redditclone/internal/repository"
)

// PostService defines the methods for post-related business logic.
type PostService interface {
	CreatePost(post *models.Post) error
	GetPostByID(id int) (*models.Post, error)
	GetPostsBySubreddit(subredditID int, limit, offset int) ([]*models.Post, error)
	GetFeedPosts(limit, offset int) ([]*models.Post, error)
	UpdatePost(post *models.Post) error
	DeletePost(id int) error
}

type postService struct {
	PostRepo      repository.PostRepository
	SubredditRepo repository.SubredditRepository
	UserRepo      repository.UserRepository
}

// NewPostService creates a new PostService.
func NewPostService(postRepo repository.PostRepository, subredditRepo repository.SubredditRepository, userRepo repository.UserRepository) PostService {
	return &postService{
		PostRepo:      postRepo,
		SubredditRepo: subredditRepo,
		UserRepo:      userRepo,
	}
}

// CreatePost handles the creation of a new post.
func (s *postService) CreatePost(post *models.Post) error {
	// Validate input
	if post.Title == "" || post.Content == "" {
		return errors.New("CreatePost: title and content are required")
	}

	// Check if author exists
	_, err := s.UserRepo.GetUserByID(post.AuthorID)
	if err != nil {
		return errors.New("CreatePost: author does not exist")
	}

	// Check if subreddit exists
	subreddit, err := s.SubredditRepo.GetSubredditByID(post.SubredditID)
	if err != nil {
		return errors.New("CreatePost: subreddit does not exist")
	}
	_=subreddit

	// Optionally, check if the user has joined the subreddit

	// Create the post via the repository
	err = s.PostRepo.CreatePost(post)
	if err != nil {
		return err
	}

	return nil
}

// GetPostByID retrieves a post by its ID.
func (s *postService) GetPostByID(id int) (*models.Post, error) {
	post, err := s.PostRepo.GetPostByID(id)
	if err != nil {
		return nil, err
	}
	return post, nil
}

// GetPostsBySubreddit retrieves posts from a specific subreddit with pagination.
func (s *postService) GetPostsBySubreddit(subredditID int, limit, offset int) ([]*models.Post, error) {
	// Check if subreddit exists
	_, err := s.SubredditRepo.GetSubredditByID(subredditID)
	if err != nil {
		return nil, errors.New("GetPostsBySubreddit: subreddit does not exist")
	}

	posts, err := s.PostRepo.GetPostsBySubreddit(subredditID, limit, offset)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

// GetFeedPosts retrieves posts for the feed with pagination.
func (s *postService) GetFeedPosts(limit, offset int) ([]*models.Post, error) {
	posts, err := s.PostRepo.GetFeedPosts(limit, offset)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

// UpdatePost updates a post's information.
func (s *postService) UpdatePost(post *models.Post) error {
	// Validate input
	if post.Title == "" || post.Content == "" {
		return errors.New("UpdatePost: title and content are required")
	}

	// Check if post exists
	existingPost, err := s.PostRepo.GetPostByID(post.ID)
	if err != nil {
		return errors.New("UpdatePost: post does not exist")
	}

	// Optionally, check if the user is authorized to update the post (e.g., the author)

	// Update fields
	existingPost.Title = post.Title
	existingPost.Content = post.Content

	// Update the post via the repository
	err = s.PostRepo.UpdatePost(existingPost)
	if err != nil {
		return err
	}

	return nil
}

// DeletePost removes a post from the system.
func (s *postService) DeletePost(id int) error {
	// Optionally, check if the user is authorized to delete the post

	err := s.PostRepo.DeletePost(id)
	if err != nil {
		return err
	}
	return nil
}
