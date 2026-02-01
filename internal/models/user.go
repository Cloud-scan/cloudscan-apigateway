package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID             string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email          string         `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash   string         `gorm:"not null" json:"-"` // Don't include in JSON responses
	FirstName      string         `gorm:"size:100" json:"first_name"`
	LastName       string         `gorm:"size:100" json:"last_name"`
	Role           string         `gorm:"size:50;default:'user'" json:"role"` // admin, user, viewer
	OrganizationID string         `gorm:"type:uuid;not null" json:"organization_id"`
	Organization   Organization   `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	LastLoginAt    *time.Time     `json:"last_login_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "users"
}

// BeforeCreate hook to set defaults before creating a user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Role == "" {
		u.Role = "user"
	}
	if !u.IsActive {
		u.IsActive = true
	}
	return nil
}