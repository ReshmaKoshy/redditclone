// File: internal/service/comment_service.go

package service

import (
	"errors"

	"redditclone/internal/models"
	"redditclone/internal/repository"
)

// CommentService defines the methods for comment-related business logic.
type CommentService interface {
	AddComment(comment *models.Comment) error
	ReplyToComment(comment *models.Comment) error
	GetCommentByID(id int) (*models.Comment, error)
	GetCommentsByPost(postID int, limit, offset int) ([]*models.Comment, error)
	GetReplies(parentID int, limit, offset int) ([]*models.Comment, error)
	UpdateComment(comment *models.Comment) error
	DeleteComment(id int) error
}

type commentService struct {
	CommentRepo   repository.CommentRepository
	PostRepo      repository.PostRepository
	UserRepo      repository.UserRepository
	SubredditRepo repository.SubredditRepository
	// Add additional repositories if necessary
}

// NewCommentService creates a new CommentService.
func NewCommentService(commentRepo repository.CommentRepository, postRepo repository.PostRepository, userRepo repository.UserRepository, subredditRepo repository.SubredditRepository) CommentService {
	return &commentService{
		CommentRepo:   commentRepo,
		PostRepo:      postRepo,
		UserRepo:      userRepo,
		SubredditRepo: subredditRepo,
	}
}

// AddComment adds a new comment to a post.
func (s *commentService) AddComment(comment *models.Comment) error {
	// Validate input
	if comment.Content == "" {
		return errors.New("AddComment: content is required")
	}

	// Check if author exists
	_, err := s.UserRepo.GetUserByID(comment.AuthorID)
	if err != nil {
		return errors.New("AddComment: author does not exist")
	}

	// Check if post exists
	post, err := s.PostRepo.GetPostByID(comment.PostID)
	if err != nil {
		return errors.New("AddComment: post does not exist")
	}
	_=post

	// Optionally, check if the user has joined the subreddit's post

	// Create the comment via the repository
	err = s.CommentRepo.CreateComment(comment)
	if err != nil {
		return err
	}

	return nil
}

// ReplyToComment adds a reply to an existing comment.
func (s *commentService) ReplyToComment(comment *models.Comment) error {
	// Validate input
	if comment.Content == "" {
		return errors.New("ReplyToComment: content is required")
	}

	// Check if author exists
	_, err := s.UserRepo.GetUserByID(comment.AuthorID)
	if err != nil {
		return errors.New("ReplyToComment: author does not exist")
	}

	// Check if parent comment exists
	parentComment, err := s.CommentRepo.GetCommentByID(*comment.ParentID)
	if err != nil {
		return errors.New("ReplyToComment: parent comment does not exist")
	}

	// The reply should belong to the same post as the parent comment
	comment.PostID = parentComment.PostID

	// Create the reply via the repository
	err = s.CommentRepo.CreateComment(comment)
	if err != nil {
		return err
	}

	return nil
}

// GetCommentByID retrieves a comment by its ID.
func (s *commentService) GetCommentByID(id int) (*models.Comment, error) {
	comment, err := s.CommentRepo.GetCommentByID(id)
	if err != nil {
		return nil, err
	}
	return comment, nil
}

// GetCommentsByPost retrieves top-level comments for a specific post with pagination.
func (s *commentService) GetCommentsByPost(postID int, limit, offset int) ([]*models.Comment, error) {
	// Check if post exists
	_, err := s.PostRepo.GetPostByID(postID)
	if err != nil {
		return nil, errors.New("GetCommentsByPost: post does not exist")
	}

	comments, err := s.CommentRepo.GetCommentsByPost(postID, limit, offset)
	if err != nil {
		return nil, err
	}
	return comments, nil
}

// GetReplies retrieves replies to a specific comment with pagination.
func (s *commentService) GetReplies(parentID int, limit, offset int) ([]*models.Comment, error) {
	// Check if parent comment exists
	_, err := s.CommentRepo.GetCommentByID(parentID)
	if err != nil {
		return nil, errors.New("GetReplies: parent comment does not exist")
	}

	replies, err := s.CommentRepo.GetReplies(parentID, limit, offset)
	if err != nil {
		return nil, err
	}
	return replies, nil
}

// UpdateComment updates a comment's content.
func (s *commentService) UpdateComment(comment *models.Comment) error {
	// Validate input
	if comment.Content == "" {
		return errors.New("UpdateComment: content is required")
	}

	// Check if comment exists
	existingComment, err := s.CommentRepo.GetCommentByID(comment.ID)
	if err != nil {
		return errors.New("UpdateComment: comment does not exist")
	}

	// Optionally, check if the user is authorized to update the comment (e.g., the author)

	// Update fields
	existingComment.Content = comment.Content

	// Update the comment via the repository
	err = s.CommentRepo.UpdateComment(existingComment)
	if err != nil {
		return err
	}

	return nil
}

// DeleteComment removes a comment from the system.
func (s *commentService) DeleteComment(id int) error {
	// Optionally, check if the user is authorized to delete the comment

	err := s.CommentRepo.DeleteComment(id)
	if err != nil {
		return err
	}
	return nil
}
