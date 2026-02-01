package models

import (
	"time"

	"gorm.io/gorm"
)

// APIKey represents an API key for service-to-service authentication
type APIKey struct {
	ID             string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name           string         `gorm:"size:200;not null" json:"name"`
	KeyHash        string         `gorm:"uniqueIndex;not null" json:"-"` // Hashed API key
	OrganizationID string         `gorm:"type:uuid;not null" json:"organization_id"`
	Organization   Organization   `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	Scopes         string         `gorm:"type:text" json:"scopes"` // Comma-separated scopes
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	ExpiresAt      *time.Time     `json:"expires_at,omitempty"`
	LastUsedAt     *time.Time     `json:"last_used_at,omitempty"`
	CreatedBy      string         `gorm:"type:uuid" json:"created_by"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for APIKey
func (APIKey) TableName() string {
	return "api_keys"
}

// BeforeCreate hook to set defaults before creating an API key
func (a *APIKey) BeforeCreate(tx *gorm.DB) error {
	if !a.IsActive {
		a.IsActive = true
	}
	return nil
}

// IsExpired checks if the API key is expired
func (a *APIKey) IsExpired() bool {
	if a.ExpiresAt == nil {
		return false
	}
	return a.ExpiresAt.Before(time.Now())
}