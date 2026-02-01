package repository

import (
	"github.com/cloud-scan/cloudscan-apigateway/internal/models"
	"gorm.io/gorm"
)

// ProjectRepository handles project data access
type ProjectRepository struct {
	db *gorm.DB
}

// NewProjectRepository creates a new project repository
func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// Create creates a new project
func (r *ProjectRepository) Create(project *models.Project) error {
	return r.db.Create(project).Error
}

// FindByID finds a project by ID
func (r *ProjectRepository) FindByID(id string) (*models.Project, error) {
	var project models.Project
	err := r.db.Preload("Organization").Where("id = ?", id).First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// Update updates a project
func (r *ProjectRepository) Update(project *models.Project) error {
	return r.db.Save(project).Error
}

// Delete soft deletes a project
func (r *ProjectRepository) Delete(id string) error {
	return r.db.Delete(&models.Project{}, "id = ?", id).Error
}

// ListByOrganization lists all projects in an organization
func (r *ProjectRepository) ListByOrganization(organizationID string, limit, offset int) ([]*models.Project, int64, error) {
	var projects []*models.Project
	var total int64

	query := r.db.Where("organization_id = ?", organizationID)

	// Get total count
	if err := query.Model(&models.Project{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Limit(limit).Offset(offset).Find(&projects).Error
	if err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

// IncrementScanCount increments the scan count for a project
func (r *ProjectRepository) IncrementScanCount(id string) error {
	return r.db.Model(&models.Project{}).Where("id = ?", id).
		UpdateColumn("scan_count", gorm.Expr("scan_count + ?", 1)).Error
}