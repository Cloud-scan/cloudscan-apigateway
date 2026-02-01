package service

import (
	"github.com/cloud-scan/cloudscan-apigateway/internal/models"
	"github.com/cloud-scan/cloudscan-apigateway/internal/repository"
)

// OrganizationService handles organization business logic
type OrganizationService struct {
	orgRepo *repository.OrganizationRepository
}

// NewOrganizationService creates a new organization service
func NewOrganizationService(orgRepo *repository.OrganizationRepository) *OrganizationService {
	return &OrganizationService{
		orgRepo: orgRepo,
	}
}

// GetByID retrieves an organization by ID
func (s *OrganizationService) GetByID(id string) (*models.Organization, error) {
	return s.orgRepo.FindByID(id)
}

// Update updates an organization
func (s *OrganizationService) Update(org *models.Organization) error {
	return s.orgRepo.Update(org)
}

// List lists organizations with pagination
func (s *OrganizationService) List(limit, offset int) ([]*models.Organization, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.orgRepo.List(limit, offset)
}