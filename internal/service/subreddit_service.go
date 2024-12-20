// File: internal/service/subreddit_service.go

package service

import (
	"errors"

	"redditclone/internal/models"
	"redditclone/internal/repository"
)

// SubredditService defines the methods for subreddit-related business logic.
type SubredditService interface {
	CreateSubreddit(subreddit *models.Subreddit) error
	GetSubredditByID(id int) (*models.Subreddit, error)
	GetSubredditByName(name string) (*models.Subreddit, error)
	UpdateSubreddit(subreddit *models.Subreddit) error
	DeleteSubreddit(id int) error
	JoinSubreddit(userID, subredditID int) error
	LeaveSubreddit(userID, subredditID int) error
}

type subredditService struct {
	SubredditRepo repository.SubredditRepository
	// Add additional repositories if necessary, e.g., for managing memberships
}

// NewSubredditService creates a new SubredditService.
func NewSubredditService(subredditRepo repository.SubredditRepository) SubredditService {
	return &subredditService{SubredditRepo: subredditRepo}
}

// CreateSubreddit handles the creation of a new subreddit.
func (s *subredditService) CreateSubreddit(subreddit *models.Subreddit) error {
	// Validate input
	if subreddit.Name == "" {
		return errors.New("CreateSubreddit: name is required")
	}

	// Check if subreddit name already exists
	existingSubreddit, err := s.SubredditRepo.GetSubredditByName(subreddit.Name)
	if err == nil && existingSubreddit != nil {
		return errors.New("CreateSubreddit: subreddit name already exists")
	}

	// Create the subreddit via the repository
	err = s.SubredditRepo.CreateSubreddit(subreddit)
	if err != nil {
		return err
	}

	return nil
}

// GetSubredditByID retrieves a subreddit by its ID.
func (s *subredditService) GetSubredditByID(id int) (*models.Subreddit, error) {
	subreddit, err := s.SubredditRepo.GetSubredditByID(id)
	if err != nil {
		return nil, err
	}
	return subreddit, nil
}

// GetSubredditByName retrieves a subreddit by its name.
func (s *subredditService) GetSubredditByName(name string) (*models.Subreddit, error) {
	subreddit, err := s.SubredditRepo.GetSubredditByName(name)
	if err != nil {
		return nil, err
	}
	return subreddit, nil
}

// UpdateSubreddit updates a subreddit's information.
func (s *subredditService) UpdateSubreddit(subreddit *models.Subreddit) error {
	// Validate input
	if subreddit.Name == "" {
		return errors.New("UpdateSubreddit: name is required")
	}

	// Check if subreddit exists
	existingSubreddit, err := s.SubredditRepo.GetSubredditByID(subreddit.ID)
	if err != nil {
		return err
	}

	// Optionally, check if the new name is already taken by another subreddit
	if existingSubreddit.Name != subreddit.Name {
		anotherSubreddit, err := s.SubredditRepo.GetSubredditByName(subreddit.Name)
		if err == nil && anotherSubreddit != nil {
			return errors.New("UpdateSubreddit: new subreddit name already exists")
		}
	}

	// Update the subreddit via the repository
	err = s.SubredditRepo.UpdateSubreddit(subreddit)
	if err != nil {
		return err
	}

	return nil
}

// DeleteSubreddit removes a subreddit from the system.
func (s *subredditService) DeleteSubreddit(id int) error {
	// Perform any additional checks if necessary
	err := s.SubredditRepo.DeleteSubreddit(id)
	if err != nil {
		return err
	}
	return nil
}

// JoinSubreddit allows a user to join a subreddit.
// TODO: Implement membership management if required.
func (s *subredditService) JoinSubreddit(userID, subredditID int) error {
	// Placeholder for joining logic.
	// For simplicity, assuming no separate membership table.
	// In a real application, you'd manage a memberships table.

	// Example: Check if subreddit exists
	_, err := s.SubredditRepo.GetSubredditByID(subredditID)
	if err != nil {
		return err
	}

	// Implement join logic, e.g., add entry to memberships table.

	return nil
}

// LeaveSubreddit allows a user to leave a subreddit.
// TODO: Implement membership management if required.
func (s *subredditService) LeaveSubreddit(userID, subredditID int) error {
	// Placeholder for leaving logic.
	// For simplicity, assuming no separate membership table.
	// In a real application, you'd manage a memberships table.

	// Example: Check if subreddit exists
	_, err := s.SubredditRepo.GetSubredditByID(subredditID)
	if err != nil {
		return err
	}

	// Implement leave logic, e.g., remove entry from memberships table.

	return nil
}
