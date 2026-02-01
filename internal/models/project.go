package models

import (
	"time"

	"gorm.io/gorm"
)

// Project represents a project that can be scanned
type Project struct {
	ID             string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name           string         `gorm:"size:200;not null" json:"name"`
	Description    string         `gorm:"type:text" json:"description,omitempty"`
	OrganizationID string         `gorm:"type:uuid;not null" json:"organization_id"`
	Organization   Organization   `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	RepositoryURL  string         `gorm:"size:500" json:"repository_url,omitempty"`
	DefaultBranch  string         `gorm:"size:100;default:'main'" json:"default_branch"`
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	ScanCount      int            `gorm:"default:0" json:"scan_count"`
	LastScanAt     *time.Time     `json:"last_scan_at,omitempty"`
	CreatedBy      string         `gorm:"type:uuid" json:"created_by"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for Project
func (Project) TableName() string {
	return "projects"
}

// BeforeCreate hook to set defaults before creating a project
func (p *Project) BeforeCreate(tx *gorm.DB) error {
	if p.DefaultBranch == "" {
		p.DefaultBranch = "main"
	}
	if !p.IsActive {
		p.IsActive = true
	}
	return nil
}