// File: internal/service/user_service.go

package service

import (
	"errors"
	"fmt"

	"redditclone/internal/models"
	"redditclone/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

// UserService defines the methods for user-related business logic.
type UserService interface {
	RegisterUser(user *models.User) error
	GetUserProfile(id int) (*models.User, error)
	UpdateUserProfile(user *models.User) error
	DeleteUser(id int) error
	AuthenticateUser(username, password string) (*models.User, error)
}

type userService struct {
	UserRepo repository.UserRepository
}

// NewUserService creates a new UserService.
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{UserRepo: userRepo}
}

// RegisterUser handles user registration logic.
func (s *userService) RegisterUser(user *models.User) error {
	// Validate input
	if user.Username == "" || user.Email == "" || user.Password == "" {
		fmt.Println(user.Username+user.Email+user.Password)

		return errors.New("RegisterUser: all fields are required")
	}

	// Check if username already exists
	existingUser, err := s.UserRepo.GetUserByUsername(user.Username)
	if err == nil && existingUser != nil {
		return errors.New("RegisterUser: username already taken")
	}

	// Hash the password
	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword

	// Create the user in the repository
	err = s.UserRepo.CreateUser(user)
	if err != nil {
		return err
	}

	return nil
}

// GetUserProfile retrieves a user's profile by ID.
func (s *userService) GetUserProfile(id int) (*models.User, error) {
	user, err := s.UserRepo.GetUserByID(id)
	if err != nil {
		return nil, err
	}

	// Optionally, you can remove sensitive information before returning
	user.Password = ""
	return user, nil
}

// UpdateUserProfile updates a user's profile information.
func (s *userService) UpdateUserProfile(user *models.User) error {
	// Validate input
	if user.Username == "" || user.Email == "" {
		return errors.New("UpdateUserProfile: username and email are required")
	}

	// Retrieve existing user
	existingUser, err := s.UserRepo.GetUserByID(user.ID)
	if err != nil {
		return err
	}

	// If password is being updated, hash it
	if user.Password != "" {
		hashedPassword, err := hashPassword(user.Password)
		if err != nil {
			return err
		}
		existingUser.Password = hashedPassword
	}

	// Update fields
	existingUser.Username = user.Username
	existingUser.Email = user.Email

	// Update the user in the repository
	err = s.UserRepo.UpdateUser(existingUser)
	if err != nil {
		return err
	}

	return nil
}

// DeleteUser removes a user from the system.
func (s *userService) DeleteUser(id int) error {
	// Perform any additional checks if necessary
	err := s.UserRepo.DeleteUser(id)
	if err != nil {
		return err
	}
	return nil
}

// AuthenticateUser verifies user credentials.
func (s *userService) AuthenticateUser(username, password string) (*models.User, error) {
	user, err := s.UserRepo.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	// Compare hashed password with the provided password
	err = comparePasswords(user.Password, password)
	if err != nil {
		return nil, errors.New("AuthenticateUser: invalid credentials")
	}

	// Remove sensitive information before returning
	user.Password = ""
	return user, nil
}

// Helper function to hash passwords
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// Helper function to compare passwords
func comparePasswords(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
