// File: internal/service/vote_service.go

package service

import (
	"errors"
	"fmt"

	"redditclone/internal/models"
	"redditclone/internal/repository"
)

// VoteService defines the methods for vote-related business logic.
type VoteService interface {
	CastVote(vote *models.Vote) error
	ChangeVote(vote *models.Vote) error
	RemoveVote(userID, postID, commentID int) error
	GetVote(userID, postID, commentID int) (*models.Vote, error)
	UpdateKarma(targetType string, targetID int, delta int) error
}

type voteService struct {
	VoteRepo    repository.VoteRepository
	PostRepo    repository.PostRepository
	CommentRepo repository.CommentRepository
}

// NewVoteService creates a new VoteService.
func NewVoteService(voteRepo repository.VoteRepository, postRepo repository.PostRepository, commentRepo repository.CommentRepository) VoteService {
	return &voteService{
		VoteRepo:    voteRepo,
		PostRepo:    postRepo,
		CommentRepo: commentRepo,
	}
}

// CastVote allows a user to cast a vote on a post or comment.
func (s *voteService) CastVote(vote *models.Vote) error {
	// Validate input
	// if vote.VoteType != models.Upvote && vote.VoteType != models.Downvote {
	// 	return errors.New("CastVote: invalid vote type")
	// }
	if (vote.PostID == nil && vote.CommentID == nil) || (vote.PostID != nil && vote.CommentID != nil) {
		return errors.New("CastVote: vote must be on either a post or a comment")
	}

	// Check if user has already voted on the target
	var existingVote *models.Vote
	var err error
	if vote.PostID != nil {
		existingVote, err = s.VoteRepo.GetVoteByUserAndPost(vote.UserID, *vote.PostID)
	} else {
		existingVote, err = s.VoteRepo.GetVoteByUserAndComment(vote.UserID, *vote.CommentID)
	}

	if err == nil && existingVote != nil {
		return errors.New("CastVote: user has already voted on this target")
	} else if err != nil && err.Error() != "GetVoteByUserAndPost: vote not found" && err.Error() != "GetVoteByUserAndComment: vote not found" {
		return fmt.Errorf("CastVote: %v", err)
	}

	// Create the vote via the repository
	err = s.VoteRepo.CreateVote(vote)
	if err != nil {
		return err
	}

	// Update karma on the target
	var delta int
	if vote.VoteType =="upvote" {
		delta = 1
	} else {
		delta = -1
	}

	if vote.PostID != nil {
		err = s.UpdateKarma("post", *vote.PostID, delta)
	} else {
		err = s.UpdateKarma("comment", *vote.CommentID, delta)
	}
	if err != nil {
		return err
	}

	return nil
}

// ChangeVote allows a user to change their existing vote.
func (s *voteService) ChangeVote(vote *models.Vote) error {
	// Validate input
	if vote.VoteType != "upvote" && vote.VoteType != "downvote" {
		return errors.New("ChangeVote: invalid vote type")
	}
	if (vote.PostID == nil && vote.CommentID == nil) || (vote.PostID != nil && vote.CommentID != nil) {
		return errors.New("ChangeVote: vote must be on either a post or a comment")
	}

	// Retrieve existing vote
	var existingVote *models.Vote
	var err error
	if vote.PostID != nil {
		existingVote, err = s.VoteRepo.GetVoteByUserAndPost(vote.UserID, *vote.PostID)
	} else {
		existingVote, err = s.VoteRepo.GetVoteByUserAndComment(vote.UserID, *vote.CommentID)
	}
	if err != nil {
		return fmt.Errorf("ChangeVote: %v", err)
	}

	// If the vote type is the same, do nothing
	if existingVote.VoteType == vote.VoteType {
		return errors.New("ChangeVote: vote type is already set to the desired value")
	}

	// Update the vote via the repository
	existingVote.VoteType = vote.VoteType
	err = s.VoteRepo.UpdateVote(existingVote)
	if err != nil {
		return err
	}

	// Update karma on the target
	var delta int
	if vote.VoteType == "upvote" {
		delta = 2 // From -1 to +1
	} else {
		delta = -2 // From +1 to -1
	}

	if vote.PostID != nil {
		err = s.UpdateKarma("post", *vote.PostID, delta)
	} else {
		err = s.UpdateKarma("comment", *vote.CommentID, delta)
	}
	if err != nil {
		return err
	}

	return nil
}

// RemoveVote allows a user to remove their vote from a post or comment.
func (s *voteService) RemoveVote(userID, postID, commentID int) error {
	// Retrieve existing vote
	var existingVote *models.Vote
	var err error
	if postID != 0 {
		existingVote, err = s.VoteRepo.GetVoteByUserAndPost(userID, postID)
	} else {
		existingVote, err = s.VoteRepo.GetVoteByUserAndComment(userID, commentID)
	}
	if err != nil {
		return fmt.Errorf("RemoveVote: %v", err)
	}

	// Determine the delta based on the vote type
	var delta int
	if existingVote.VoteType == "upvote" {
		delta = -1
	} else {
		delta = 1
	}

	// Delete the vote via the repository
	err = s.VoteRepo.DeleteVote(existingVote)
	if err != nil {
		return err
	}

	// Update karma on the target
	if existingVote.PostID != nil {
		err = s.UpdateKarma("post", *existingVote.PostID, delta)
	} else if existingVote.CommentID != nil {
		err = s.UpdateKarma("comment", *existingVote.CommentID, delta)
	} else {
		return errors.New("RemoveVote: invalid vote target")
	}
	if err != nil {
		return err
	}

	return nil
}

// GetVote retrieves a user's vote on a specific post or comment.
func (s *voteService) GetVote(userID, postID, commentID int) (*models.Vote, error) {
	if postID != 0 && commentID != 0 {
		return nil, errors.New("GetVote: vote must be on either a post or a comment")
	}

	var vote *models.Vote
	var err error
	if postID != 0 {
		vote, err = s.VoteRepo.GetVoteByUserAndPost(userID, postID)
	} else {
		vote, err = s.VoteRepo.GetVoteByUserAndComment(userID, commentID)
	}

	if err != nil {
		return nil, err
	}
	return vote, nil
}

// UpdateKarma updates the karma of a post or comment.
func (s *voteService) UpdateKarma(targetType string, targetID int, delta int) error {
	switch targetType {
	case "post":
		post, err := s.PostRepo.GetPostByID(targetID)
		if err != nil {
			return err
		}
		post.Karma += delta
		err = s.PostRepo.UpdatePost(post)
		if err != nil {
			return err
		}
	case "comment":
		comment, err := s.CommentRepo.GetCommentByID(targetID)
		if err != nil {
			return err
		}
		comment.Karma += delta
		err = s.CommentRepo.UpdateComment(comment)
		if err != nil {
			return err
		}
	default:
		return errors.New("UpdateKarma: invalid target type")
	}
	return nil
}
