package users

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/nodus-protocol/backend/internal/models"
	"github.com/nodus-protocol/backend/internal/utils"
	"go.uber.org/zap"
)

// Service holds business logic for user profile management.
type Service struct {
	repo *Repository
	log  *zap.Logger
}

func NewService(repo *Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// GetProfile returns the user profile for the given ID.
func (s *Service) GetProfile(userID uuid.UUID) (*models.User, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// UpdateProfile updates the user's name and/or avatar.
func (s *Service) UpdateProfile(userID uuid.UUID, firstName, lastName, avatarURL string) (*models.User, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if firstName != "" {
		user.FirstName = firstName
	}
	if lastName != "" {
		user.LastName = lastName
	}
	if avatarURL != "" {
		user.AvatarURL = avatarURL
	}

	if err := s.repo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	s.log.Info("user profile updated", zap.String("user_id", userID.String()))
	return user, nil
}

// ChangePassword verifies the old password and sets a new one.
func (s *Service) ChangePassword(userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return ErrUserNotFound
	}

	if !utils.CheckPassword(oldPassword, user.PasswordHash) {
		return ErrInvalidPassword
	}

	newHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("password hashing failed: %w", err)
	}

	user.PasswordHash = newHash
	return s.repo.Update(user)
}

// DeleteAccount soft-deletes the user's account.
func (s *Service) DeleteAccount(userID uuid.UUID) error {
	return s.repo.SoftDelete(userID)
}

var (
	ErrInvalidPassword = errors.New("current password is incorrect")
)
