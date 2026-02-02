package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/cloud-scan/cloudscan-apigateway/internal/api/middleware"
	grpcClient "github.com/cloud-scan/cloudscan-apigateway/internal/grpc"
	"github.com/cloud-scan/cloudscan-apigateway/internal/repository"
	pb "github.com/cloud-scan/cloudscan-apigateway/proto"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

// ScanHandler handles scan endpoints (proxies to Orchestrator)
type ScanHandler struct {
	orchestratorClient *grpcClient.OrchestratorClient
	projectRepo        *repository.ProjectRepository
}

// NewScanHandler creates a new scan handler
func NewScanHandler(orchestratorClient *grpcClient.OrchestratorClient, projectRepo *repository.ProjectRepository) *ScanHandler {
	return &ScanHandler{
		orchestratorClient: orchestratorClient,
		projectRepo:        projectRepo,
	}
}

// CreateScan creates a new scan
// POST /api/v1/scans
func (h *ScanHandler) CreateScan(c echo.Context) error {
	organizationID := middleware.GetOrganizationID(c)

	var req struct {
		ProjectID         string   `json:"project_id" validate:"required"`
		ScanTypes         []string `json:"scan_types" validate:"required,min=1"`
		GitURL            string   `json:"git_url"`
		GitBranch         string   `json:"git_branch"`
		GitCommit         string   `json:"git_commit"`
		SourceArtifactID  string   `json:"source_artifact_id"`
	}

	if err := middleware.BindAndValidate(c, &req); err != nil {
		return err
	}

	// Fetch project to get default repository settings
	project, err := h.projectRepo.FindByID(req.ProjectID)
	if err != nil {
		log.WithError(err).WithField("project_id", req.ProjectID).Error("Failed to fetch project")
		return echo.NewHTTPError(http.StatusNotFound, "project not found")
	}

	// Use project defaults if git URL/branch not provided
	// This allows scans to inherit from project settings by default
	gitURL := req.GitURL
	gitBranch := req.GitBranch

	// Only use project defaults for Git flow (not artifact flow)
	if req.SourceArtifactID == "" {
		// Git flow: Use project's repository if not overridden
		if gitURL == "" {
			gitURL = project.RepositoryURL
		}
		if gitBranch == "" {
			gitBranch = project.DefaultBranch
		}
	}

	// Convert scan types from strings to proto enum
	scanTypes := make([]pb.ScanType, 0, len(req.ScanTypes))
	for _, st := range req.ScanTypes {
		switch st {
		case "SAST":
			scanTypes = append(scanTypes, pb.ScanType_SAST)
		case "SCA":
			scanTypes = append(scanTypes, pb.ScanType_SCA)
		case "SECRETS":
			scanTypes = append(scanTypes, pb.ScanType_SECRETS)
		case "LICENSE":
			scanTypes = append(scanTypes, pb.ScanType_LICENSE)
		default:
			return echo.NewHTTPError(http.StatusBadRequest, "invalid scan type: "+st)
		}
	}

	// Create scan via gRPC
	createReq := &pb.CreateScanRequest{
		OrganizationId:   organizationID,
		ProjectId:        req.ProjectID,
		ScanTypes:        scanTypes,
		GitUrl:           gitURL,
		GitBranch:        gitBranch,
		GitCommit:        req.GitCommit,
		SourceArtifactId: req.SourceArtifactID,
	}

	scan, err := h.orchestratorClient.CreateScan(c.Request().Context(), createReq)
	if err != nil {
		log.WithError(err).Error("Failed to create scan via gRPC")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create scan")
	}

	return c.JSON(http.StatusCreated, scan)
}

// ListScans lists all scans
// GET /api/v1/scans
func (h *ScanHandler) ListScans(c echo.Context) error {
	organizationID := middleware.GetOrganizationID(c)

	// Parse query parameters
	projectID := c.QueryParam("project_id")
	statusStr := c.QueryParam("status")
	pageSize := 10
	if ps := c.QueryParam("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 {
			pageSize = v
		}
	}
	pageToken := c.QueryParam("page_token")

	// Convert status string to proto enum
	var status pb.ScanStatus
	if statusStr != "" {
		switch statusStr {
		case "QUEUED":
			status = pb.ScanStatus_QUEUED
		case "RUNNING":
			status = pb.ScanStatus_RUNNING
		case "COMPLETED":
			status = pb.ScanStatus_COMPLETED
		case "FAILED":
			status = pb.ScanStatus_FAILED
		case "CANCELLED":
			status = pb.ScanStatus_CANCELLED
		}
	}

	listReq := &pb.ListScansRequest{
		OrganizationId: organizationID,
		ProjectId:      projectID,
		Status:         status,
		PageSize:       int32(pageSize),
		PageToken:      pageToken,
	}

	resp, err := h.orchestratorClient.ListScans(c.Request().Context(), listReq)
	if err != nil {
		log.WithError(err).Error("Failed to list scans via gRPC")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list scans")
	}

	return c.JSON(http.StatusOK, resp)
}

// GetScan retrieves a scan by ID
// GET /api/v1/scans/:id
func (h *ScanHandler) GetScan(c echo.Context) error {
	scanID := c.Param("id")

	scan, err := h.orchestratorClient.GetScan(c.Request().Context(), scanID)
	if err != nil {
		log.WithError(err).WithField("scan_id", scanID).Error("Failed to get scan via gRPC")
		return echo.NewHTTPError(http.StatusNotFound, "scan not found")
	}

	return c.JSON(http.StatusOK, scan)
}

// CancelScan cancels a running scan
// PUT /api/v1/scans/:id/cancel
func (h *ScanHandler) CancelScan(c echo.Context) error {
	scanID := c.Param("id")

	err := h.orchestratorClient.CancelScan(c.Request().Context(), scanID)
	if err != nil {
		log.WithError(err).WithField("scan_id", scanID).Error("Failed to cancel scan via gRPC")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to cancel scan")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "scan cancelled successfully",
	})
}

// GetFindings retrieves findings for a scan
// GET /api/v1/scans/:id/findings
func (h *ScanHandler) GetFindings(c echo.Context) error {
	scanID := c.Param("id")

	// Parse query parameters
	scanTypeStr := c.QueryParam("scan_type")
	severityStr := c.QueryParam("severity")
	pageSize := 50
	if ps := c.QueryParam("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 {
			pageSize = v
		}
	}
	pageToken := c.QueryParam("page_token")

	// Convert scan type string to proto enum
	var scanType pb.ScanType
	if scanTypeStr != "" {
		switch scanTypeStr {
		case "SAST":
			scanType = pb.ScanType_SAST
		case "SCA":
			scanType = pb.ScanType_SCA
		case "SECRETS":
			scanType = pb.ScanType_SECRETS
		case "LICENSE":
			scanType = pb.ScanType_LICENSE
		}
	}

	// Convert severity string to proto enum
	var severity pb.Severity
	if severityStr != "" {
		switch severityStr {
		case "CRITICAL":
			severity = pb.Severity_CRITICAL
		case "HIGH":
			severity = pb.Severity_HIGH
		case "MEDIUM":
			severity = pb.Severity_MEDIUM
		case "LOW":
			severity = pb.Severity_LOW
		case "INFO":
			severity = pb.Severity_INFO
		}
	}

	findingsReq := &pb.GetFindingsRequest{
		ScanId:    scanID,
		ScanType:  scanType,
		Severity:  severity,
		PageSize:  int32(pageSize),
		PageToken: pageToken,
	}

	resp, err := h.orchestratorClient.GetFindings(c.Request().Context(), findingsReq)
	if err != nil {
		log.WithError(err).WithField("scan_id", scanID).Error("Failed to get findings via gRPC")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get findings")
	}

	return c.JSON(http.StatusOK, resp)
}

// GetScansSummary returns summary statistics for scans
// GET /api/v1/scans/summary
// TODO: This is a temporary implementation using multiple API calls.
// In the future, this should be implemented as a dedicated RPC method in the orchestrator
// with optimized database queries for better performance.
func (h *ScanHandler) GetScansSummary(c echo.Context) error {
	organizationID := middleware.GetOrganizationID(c)

	// Get total scans count (just need total_count, not actual data)
	allScansReq := &pb.ListScansRequest{
		OrganizationId: organizationID,
		PageSize:       1, // Minimize data transfer, we only need total_count
	}
	allScansResp, err := h.orchestratorClient.ListScans(c.Request().Context(), allScansReq)
	if err != nil {
		log.WithError(err).Error("Failed to get total scans count")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get summary")
	}

	// Get running scans count
	runningReq := &pb.ListScansRequest{
		OrganizationId: organizationID,
		Status:         pb.ScanStatus_RUNNING,
		PageSize:       1,
	}
	runningResp, err := h.orchestratorClient.ListScans(c.Request().Context(), runningReq)
	if err != nil {
		log.WithError(err).Error("Failed to get running scans count")
		// Don't fail completely, just set to 0
		runningResp = &pb.ListScansResponse{TotalCount: 0}
	}

	// Get queued scans count
	queuedReq := &pb.ListScansRequest{
		OrganizationId: organizationID,
		Status:         pb.ScanStatus_QUEUED,
		PageSize:       1,
	}
	queuedResp, err := h.orchestratorClient.ListScans(c.Request().Context(), queuedReq)
	if err != nil {
		log.WithError(err).Error("Failed to get queued scans count")
		queuedResp = &pb.ListScansResponse{TotalCount: 0}
	}

	// Get recent completed scans to calculate today's count and average duration
	completedReq := &pb.ListScansRequest{
		OrganizationId: organizationID,
		Status:         pb.ScanStatus_COMPLETED,
		PageSize:       100, // Get last 100 completed scans
	}
	completedResp, err := h.orchestratorClient.ListScans(c.Request().Context(), completedReq)
	if err != nil {
		log.WithError(err).Error("Failed to get completed scans")
		completedResp = &pb.ListScansResponse{Scans: []*pb.Scan{}}
	}

	// Calculate completed today and average duration
	completedToday := int32(0)
	totalDuration := int64(0)
	durationCount := int32(0)

	today := time.Now().UTC().Truncate(24 * time.Hour)

	for _, scan := range completedResp.Scans {
		if scan.CompletedAt != nil {
			completedTime := scan.CompletedAt.AsTime()

			// Check if completed today
			if completedTime.After(today) {
				completedToday++
			}

			// Calculate duration
			if scan.CreatedAt != nil {
				duration := completedTime.Sub(scan.CreatedAt.AsTime()).Seconds()
				totalDuration += int64(duration)
				durationCount++
			}
		}
	}

	averageDuration := int32(0)
	if durationCount > 0 {
		averageDuration = int32(totalDuration / int64(durationCount))
	}

	// Build response
	summary := map[string]interface{}{
		"total_scans":      allScansResp.TotalCount,
		"active_scans":     runningResp.TotalCount + queuedResp.TotalCount,
		"completed_today":  completedToday,
		"average_duration": averageDuration,
	}

	return c.JSON(http.StatusOK, summary)
}