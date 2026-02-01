package handlers

import (
	"net/http"
	"strconv"

	"github.com/cloud-scan/cloudscan-apigateway/internal/api/middleware"
	"github.com/cloud-scan/cloudscan-apigateway/internal/service"
	"github.com/labstack/echo/v4"
)

// ProjectHandler handles project endpoints
type ProjectHandler struct {
	projectService *service.ProjectService
}

// NewProjectHandler creates a new project handler
func NewProjectHandler(projectService *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
	}
}

// Create creates a new project
// POST /api/v1/projects
func (h *ProjectHandler) Create(c echo.Context) error {
	var req service.CreateProjectRequest

	if err := middleware.BindAndValidate(c, &req); err != nil {
		return err
	}

	organizationID := middleware.GetOrganizationID(c)
	userID := middleware.GetUserID(c)

	project, err := h.projectService.Create(&req, organizationID, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, project)
}

// GetByID retrieves a project by ID
// GET /api/v1/projects/:id
func (h *ProjectHandler) GetByID(c echo.Context) error {
	id := c.Param("id")
	organizationID := middleware.GetOrganizationID(c)

	project, err := h.projectService.GetByID(id, organizationID)
	if err != nil {
		if err == service.ErrProjectNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "project not found")
		}
		if err == service.ErrUnauthorized {
			return echo.NewHTTPError(http.StatusForbidden, "unauthorized access to project")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, project)
}

// Update updates a project
// PUT /api/v1/projects/:id
func (h *ProjectHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var req service.CreateProjectRequest

	if err := middleware.BindAndValidate(c, &req); err != nil {
		return err
	}

	organizationID := middleware.GetOrganizationID(c)

	project, err := h.projectService.Update(id, &req, organizationID)
	if err != nil {
		if err == service.ErrProjectNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "project not found")
		}
		if err == service.ErrUnauthorized {
			return echo.NewHTTPError(http.StatusForbidden, "unauthorized access to project")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, project)
}

// Delete deletes a project
// DELETE /api/v1/projects/:id
func (h *ProjectHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	organizationID := middleware.GetOrganizationID(c)

	if err := h.projectService.Delete(id, organizationID); err != nil {
		if err == service.ErrProjectNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "project not found")
		}
		if err == service.ErrUnauthorized {
			return echo.NewHTTPError(http.StatusForbidden, "unauthorized access to project")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "project deleted successfully",
	})
}

// List lists all projects for the organization
// GET /api/v1/projects
func (h *ProjectHandler) List(c echo.Context) error {
	organizationID := middleware.GetOrganizationID(c)

	// Get pagination parameters
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	if limit == 0 {
		limit = 20
	}

	projects, total, err := h.projectService.ListByOrganization(organizationID, limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"projects": projects,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}