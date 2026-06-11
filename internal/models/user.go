package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRole defines the role of a user.
type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

// User represents a registered application user.
type User struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email           string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash    string         `gorm:"type:varchar(255);not null" json:"-"`
	FirstName       string         `gorm:"type:varchar(100)" json:"first_name"`
	LastName        string         `gorm:"type:varchar(100)" json:"last_name"`
	AvatarURL       string         `gorm:"type:varchar(512)" json:"avatar_url,omitempty"`
	Role            UserRole       `gorm:"type:varchar(20);default:'user'" json:"role"`
	StellarAddress  string         `gorm:"type:varchar(60);uniqueIndex" json:"stellar_address,omitempty"`
	IsEmailVerified bool           `gorm:"default:false" json:"is_email_verified"`
	IsActive        bool           `gorm:"default:true" json:"is_active"`
	LastLoginAt     *time.Time     `json:"last_login_at,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
	Tokens          []Token        `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// FullName returns the user's full display name.
func (u *User) FullName() string {
	if u.FirstName == "" && u.LastName == "" {
		return u.Email
	}
	return u.FirstName + " " + u.LastName
}
