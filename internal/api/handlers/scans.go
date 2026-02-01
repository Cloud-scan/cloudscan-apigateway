package handlers

import (
	"net/http"
	"strconv"

	"github.com/cloud-scan/cloudscan-apigateway/internal/api/middleware"
	grpcClient "github.com/cloud-scan/cloudscan-apigateway/internal/grpc"
	pb "github.com/cloud-scan/cloudscan-apigateway/proto"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

// ScanHandler handles scan endpoints (proxies to Orchestrator)
type ScanHandler struct {
	orchestratorClient *grpcClient.OrchestratorClient
}

// NewScanHandler creates a new scan handler
func NewScanHandler(orchestratorClient *grpcClient.OrchestratorClient) *ScanHandler {
	return &ScanHandler{
		orchestratorClient: orchestratorClient,
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
		GitUrl:           req.GitURL,
		GitBranch:        req.GitBranch,
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