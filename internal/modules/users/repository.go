package users

import (
	"errors"

	"github.com/google/uuid"
	"github.com/nodus-protocol/backend/internal/models"
	"gorm.io/gorm"
)

// Repository handles database operations for the users domain.
type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// FindByID retrieves an active user by UUID.
func (r *Repository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ? AND is_active = true", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return &user, err
}

// Update saves updated fields on a user.
func (r *Repository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// SoftDelete marks a user as inactive (soft delete).
func (r *Repository) SoftDelete(id uuid.UUID) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Update("is_active", false).Error
}

var ErrUserNotFound = errors.New("user not found")
