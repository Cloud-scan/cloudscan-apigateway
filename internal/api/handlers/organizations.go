package handlers

import (
	"net/http"
	"strconv"

	"github.com/cloud-scan/cloudscan-apigateway/internal/api/middleware"
	"github.com/cloud-scan/cloudscan-apigateway/internal/service"
	"github.com/labstack/echo/v4"
)

// OrganizationHandler handles organization endpoints
type OrganizationHandler struct {
	orgService *service.OrganizationService
}

// NewOrganizationHandler creates a new organization handler
func NewOrganizationHandler(orgService *service.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{
		orgService: orgService,
	}
}

// GetByID retrieves an organization by ID
// GET /api/v1/organizations/:id
func (h *OrganizationHandler) GetByID(c echo.Context) error {
	id := c.Param("id")
	organizationID := middleware.GetOrganizationID(c)

	// Users can only access their own organization
	if id != organizationID {
		return echo.NewHTTPError(http.StatusForbidden, "unauthorized access to organization")
	}

	org, err := h.orgService.GetByID(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "organization not found")
	}

	return c.JSON(http.StatusOK, org)
}

// Update updates an organization
// PUT /api/v1/organizations/:id
func (h *OrganizationHandler) Update(c echo.Context) error {
	id := c.Param("id")
	organizationID := middleware.GetOrganizationID(c)
	role := middleware.GetRole(c)

	// Only admins can update organization
	if role != "admin" {
		return echo.NewHTTPError(http.StatusForbidden, "only admins can update organization")
	}

	// Users can only update their own organization
	if id != organizationID {
		return echo.NewHTTPError(http.StatusForbidden, "unauthorized access to organization")
	}

	var req struct {
		DisplayName string `json:"display_name"`
		Description string `json:"description"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	org, err := h.orgService.GetByID(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "organization not found")
	}

	org.DisplayName = req.DisplayName
	org.Description = req.Description

	if err := h.orgService.Update(org); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, org)
}

// List lists all organizations (admin only)
// GET /api/v1/organizations
func (h *OrganizationHandler) List(c echo.Context) error {
	role := middleware.GetRole(c)

	// Only super admins can list all organizations
	if role != "superadmin" {
		return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	if limit == 0 {
		limit = 20
	}

	orgs, total, err := h.orgService.List(limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"organizations": orgs,
		"total":         total,
		"limit":         limit,
		"offset":        offset,
	})
}