package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TokenType represents the purpose of a stored token.
type TokenType string

const (
	TokenTypeRefresh       TokenType = "refresh"
	TokenTypeEmailVerify   TokenType = "email_verify"
	TokenTypePasswordReset TokenType = "password_reset"
)

// Token stores opaque tokens for refresh, email verification, and password resets.
type Token struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	Hash      string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"-"` // bcrypt hash of the token
	Type      TokenType      `gorm:"type:varchar(30);not null" json:"type"`
	ExpiresAt time.Time      `gorm:"not null" json:"expires_at"`
	Used      bool           `gorm:"default:false" json:"used"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	User      User           `gorm:"foreignKey:UserID" json:"-"`
}

func (t *Token) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// IsExpired returns true if the token has passed its expiry time.
func (t *Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}
