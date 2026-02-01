package repository

import (
	"github.com/cloud-scan/cloudscan-apigateway/internal/models"
	"gorm.io/gorm"
)

// OrganizationRepository handles organization data access
type OrganizationRepository struct {
	db *gorm.DB
}

// NewOrganizationRepository creates a new organization repository
func NewOrganizationRepository(db *gorm.DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

// Create creates a new organization
func (r *OrganizationRepository) Create(org *models.Organization) error {
	return r.db.Create(org).Error
}

// FindByID finds an organization by ID
func (r *OrganizationRepository) FindByID(id string) (*models.Organization, error) {
	var org models.Organization
	err := r.db.Where("id = ?", id).First(&org).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// FindByName finds an organization by name
func (r *OrganizationRepository) FindByName(name string) (*models.Organization, error) {
	var org models.Organization
	err := r.db.Where("name = ?", name).First(&org).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// Update updates an organization
func (r *OrganizationRepository) Update(org *models.Organization) error {
	return r.db.Save(org).Error
}

// Delete soft deletes an organization
func (r *OrganizationRepository) Delete(id string) error {
	return r.db.Delete(&models.Organization{}, "id = ?", id).Error
}

// List lists all organizations with pagination
func (r *OrganizationRepository) List(limit, offset int) ([]*models.Organization, int64, error) {
	var orgs []*models.Organization
	var total int64

	// Get total count
	if err := r.db.Model(&models.Organization{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.Limit(limit).Offset(offset).Find(&orgs).Error
	if err != nil {
		return nil, 0, err
	}

	return orgs, total, nil
}