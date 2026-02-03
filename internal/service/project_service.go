package service

import (
	"context"
	"errors"

	"github.com/cloud-scan/cloudscan-apigateway/internal/grpc"
	"github.com/cloud-scan/cloudscan-apigateway/internal/models"
	"github.com/cloud-scan/cloudscan-apigateway/internal/repository"
	pb "github.com/cloud-scan/cloudscan-apigateway/proto"
	log "github.com/sirupsen/logrus"
)

var (
	ErrProjectNotFound = errors.New("project not found")
	ErrUnauthorized    = errors.New("unauthorized access to project")
)

// ProjectService handles project business logic
type ProjectService struct {
	projectRepo         *repository.ProjectRepository
	orchestratorClient *grpc.OrchestratorClient
}

// NewProjectService creates a new project service
func NewProjectService(projectRepo *repository.ProjectRepository, orchestratorClient *grpc.OrchestratorClient) *ProjectService {
	return &ProjectService{
		projectRepo:         projectRepo,
		orchestratorClient: orchestratorClient,
	}
}

// CreateProjectRequest represents a project creation request
type CreateProjectRequest struct {
	Name          string `json:"name" validate:"required,min=3,max=200"`
	Description   string `json:"description"`
	RepositoryURL string `json:"repository_url" validate:"omitempty,url"`
	DefaultBranch string `json:"default_branch"`
}

// Create creates a new project
func (s *ProjectService) Create(req *CreateProjectRequest, organizationID, userID string) (*models.Project, error) {
	project := &models.Project{
		Name:           req.Name,
		Description:    req.Description,
		OrganizationID: organizationID,
		RepositoryURL:  req.RepositoryURL,
		DefaultBranch:  req.DefaultBranch,
		IsActive:       true,
		ScanCount:      0,
		CreatedBy:      userID,
	}

	if project.DefaultBranch == "" {
		project.DefaultBranch = "main"
	}

	if err := s.projectRepo.Create(project); err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"project_id": project.ID,
		"name":       project.Name,
		"org_id":     organizationID,
	}).Info("Project created successfully")

	return project, nil
}

// GetByID retrieves a project by ID
func (s *ProjectService) GetByID(id, organizationID string) (*models.Project, error) {
	project, err := s.projectRepo.FindByID(id)
	if err != nil {
		return nil, ErrProjectNotFound
	}

	// Check if project belongs to the organization
	if project.OrganizationID != organizationID {
		return nil, ErrUnauthorized
	}

	return project, nil
}

// Update updates a project
func (s *ProjectService) Update(id string, req *CreateProjectRequest, organizationID string) (*models.Project, error) {
	project, err := s.GetByID(id, organizationID)
	if err != nil {
		return nil, err
	}

	// Update fields
	project.Name = req.Name
	project.Description = req.Description
	project.RepositoryURL = req.RepositoryURL
	if req.DefaultBranch != "" {
		project.DefaultBranch = req.DefaultBranch
	}

	if err := s.projectRepo.Update(project); err != nil {
		return nil, err
	}

	return project, nil
}

// Delete soft deletes a project and cascades deletion to orchestrator
func (s *ProjectService) Delete(id, organizationID string) error {
	// Verify ownership
	if _, err := s.GetByID(id, organizationID); err != nil {
		return err
	}

	ctx := context.Background()

	// Delete all scans for this project from Orchestrator
	deletedCount, err := s.orchestratorClient.DeleteProjectScans(ctx, id)
	if err != nil {
		log.WithError(err).WithField("project_id", id).Error("Failed to delete project scans from orchestrator")
		// Continue with project deletion even if orchestrator deletion fails
		// Orphaned scans will be cleaned up by orchestrator's cleaner worker if enabled
	} else {
		log.WithFields(log.Fields{
			"project_id":    id,
			"deleted_scans": deletedCount,
		}).Info("Deleted project scans from orchestrator")
	}

	// Delete project from Gateway database
	return s.projectRepo.Delete(id)
}

// ListByOrganization lists all projects in an organization with real scan counts
func (s *ProjectService) ListByOrganization(organizationID string, limit, offset int) ([]*models.Project, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	projects, total, err := s.projectRepo.ListByOrganization(organizationID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// Enrich projects with real scan counts from Orchestrator
	ctx := context.Background()
	for _, project := range projects {
		// Fetch scan count from Orchestrator for this project
		scansResp, err := s.orchestratorClient.ListScans(ctx, &pb.ListScansRequest{
			OrganizationId: organizationID,
			ProjectId:      project.ID,
			PageSize:       1, // We only need the total count
		})
		if err != nil {
			log.WithError(err).WithField("project_id", project.ID).Warn("Failed to fetch scan count from orchestrator")
			// Continue with scan_count from database (might be stale)
			continue
		}

		// Update project with real scan count
		project.ScanCount = int(scansResp.TotalCount)
	}

	return projects, total, nil
}

// IncrementScanCount increments the scan count for a project
func (s *ProjectService) IncrementScanCount(id string) error {
	return s.projectRepo.IncrementScanCount(id)
}