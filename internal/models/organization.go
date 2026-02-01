package models

import (
	"time"

	"gorm.io/gorm"
)

// Organization represents an organization/tenant in the system
type Organization struct {
	ID          string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string         `gorm:"uniqueIndex;not null;size:200" json:"name"`
	DisplayName string         `gorm:"size:200" json:"display_name"`
	Description string         `gorm:"type:text" json:"description,omitempty"`
	Plan        string         `gorm:"size:50;default:'free'" json:"plan"` // free, pro, enterprise
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	MaxProjects int            `gorm:"default:10" json:"max_projects"`
	MaxUsers    int            `gorm:"default:5" json:"max_users"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for Organization
func (Organization) TableName() string {
	return "organizations"
}

// BeforeCreate hook to set defaults before creating an organization
func (o *Organization) BeforeCreate(tx *gorm.DB) error {
	if o.Plan == "" {
		o.Plan = "free"
	}
	if !o.IsActive {
		o.IsActive = true
	}
	if o.MaxProjects == 0 {
		o.MaxProjects = 10
	}
	if o.MaxUsers == 0 {
		o.MaxUsers = 5
	}
	return nil
}