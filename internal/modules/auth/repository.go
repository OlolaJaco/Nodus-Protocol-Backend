package auth

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/nodus-protocol/backend/internal/models"
	"gorm.io/gorm"
)

// Repository handles all database operations for the auth domain.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new auth Repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// CreateUser persists a new user record.
func (r *Repository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

// FindUserByEmail retrieves an active user by email.
func (r *Repository) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ? AND is_active = true", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return &user, err
}

// FindUserByID retrieves an active user by UUID.
func (r *Repository) FindUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ? AND is_active = true", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return &user, err
}

// UpdateUserLastLogin sets the last_login_at timestamp.
func (r *Repository) UpdateUserLastLogin(userID uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("last_login_at", now).Error
}

// MarkEmailVerified marks a user's email as verified.
func (r *Repository) MarkEmailVerified(userID uuid.UUID) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("is_email_verified", true).Error
}

// UpdatePassword updates a user's password hash.
func (r *Repository) UpdatePassword(userID uuid.UUID, newHash string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("password_hash", newHash).Error
}

// CreateToken persists a new token record.
func (r *Repository) CreateToken(token *models.Token) error {
	return r.db.Create(token).Error
}

// FindValidToken retrieves a non-expired, non-used token by hash and type.
func (r *Repository) FindValidToken(hash string, tokenType models.TokenType) (*models.Token, error) {
	var token models.Token
	err := r.db.Preload("User").
		Where("hash = ? AND type = ? AND used = false AND expires_at > ?", hash, tokenType, time.Now()).
		First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTokenNotFound
	}
	return &token, err
}

// MarkTokenUsed marks a token as used so it cannot be replayed.
func (r *Repository) MarkTokenUsed(tokenID uuid.UUID) error {
	return r.db.Model(&models.Token{}).Where("id = ?", tokenID).Update("used", true).Error
}

// DeleteRefreshTokensForUser invalidates all refresh tokens for a user (full logout).
func (r *Repository) DeleteRefreshTokensForUser(userID uuid.UUID) error {
	return r.db.Where("user_id = ? AND type = ?", userID, models.TokenTypeRefresh).Delete(&models.Token{}).Error
}

// DeleteExpiredTokens is a housekeeping function to clean up old tokens.
func (r *Repository) DeleteExpiredTokens() error {
	return r.db.Where("expires_at < ? OR used = true", time.Now()).Delete(&models.Token{}).Error
}

// Sentinel errors
var (
	ErrUserNotFound  = errors.New("user not found")
	ErrTokenNotFound = errors.New("token not found or expired")
)
