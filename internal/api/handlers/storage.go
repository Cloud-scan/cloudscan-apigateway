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

// StorageHandler handles storage endpoints (proxies to Storage service)
type StorageHandler struct {
	storageClient *grpcClient.StorageClient
}

// NewStorageHandler creates a new storage handler
func NewStorageHandler(storageClient *grpcClient.StorageClient) *StorageHandler {
	return &StorageHandler{
		storageClient: storageClient,
	}
}

// Upload handles artifact upload request (returns presigned URL)
// POST /api/v1/storage/upload
func (h *StorageHandler) Upload(c echo.Context) error {
	organizationID := middleware.GetOrganizationID(c)

	var req struct {
		ScanID          string `json:"scan_id" validate:"required"`
		Type            string `json:"type" validate:"required"`
		Filename        string `json:"filename" validate:"required"`
		ContentType     string `json:"content_type"`
		SizeBytes       int64  `json:"size_bytes" validate:"required"`
		ExpiresInHours  int32  `json:"expires_in_hours"`
	}

	if err := middleware.BindAndValidate(c, &req); err != nil {
		return err
	}

	// Default expiration
	if req.ExpiresInHours == 0 {
		req.ExpiresInHours = 24
	}

	// Convert artifact type string to proto enum
	var artifactType pb.ArtifactType
	switch req.Type {
	case "SOURCE_CODE":
		artifactType = pb.ArtifactType_SOURCE_CODE
	case "SCAN_RESULTS":
		artifactType = pb.ArtifactType_SCAN_RESULTS
	case "REPORT":
		artifactType = pb.ArtifactType_REPORT
	case "LOG":
		artifactType = pb.ArtifactType_LOG
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "invalid artifact type: "+req.Type)
	}

	// Create artifact and get presigned upload URL via gRPC
	createReq := &pb.CreateArtifactRequest{
		ScanId:         req.ScanID,
		OrganizationId: organizationID,
		Type:           artifactType,
		Filename:       req.Filename,
		ContentType:    req.ContentType,
		SizeBytes:      req.SizeBytes,
		ExpiresInHours: req.ExpiresInHours,
	}

	resp, err := h.storageClient.CreateArtifact(c.Request().Context(), createReq)
	if err != nil {
		log.WithError(err).Error("Failed to create artifact via gRPC")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create artifact")
	}

	return c.JSON(http.StatusOK, resp)
}

// Download handles artifact download request (returns presigned URL)
// GET /api/v1/storage/download/:id
func (h *StorageHandler) Download(c echo.Context) error {
	artifactID := c.Param("id")

	// Parse expiration hours from query param
	expiresInHours := int32(1) // default 1 hour
	if exp := c.QueryParam("expires_in_hours"); exp != "" {
		if v, err := strconv.Atoi(exp); err == nil && v > 0 {
			expiresInHours = int32(v)
		}
	}

	// Get artifact metadata and presigned download URL via gRPC
	resp, err := h.storageClient.GetArtifact(c.Request().Context(), artifactID, expiresInHours)
	if err != nil {
		log.WithError(err).WithField("artifact_id", artifactID).Error("Failed to get artifact via gRPC")
		return echo.NewHTTPError(http.StatusNotFound, "artifact not found")
	}

	return c.JSON(http.StatusOK, resp)
}

// Delete handles artifact deletion
// DELETE /api/v1/storage/:id
func (h *StorageHandler) Delete(c echo.Context) error {
	artifactID := c.Param("id")

	err := h.storageClient.DeleteArtifact(c.Request().Context(), artifactID)
	if err != nil {
		log.WithError(err).WithField("artifact_id", artifactID).Error("Failed to delete artifact via gRPC")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete artifact")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "artifact deleted successfully",
	})
}

// ListArtifacts lists artifacts for a scan
// GET /api/v1/storage/artifacts
func (h *StorageHandler) ListArtifacts(c echo.Context) error {
	organizationID := middleware.GetOrganizationID(c)

	// Parse query parameters
	scanID := c.QueryParam("scan_id")
	artifactTypeStr := c.QueryParam("type")
	pageSize := 50
	if ps := c.QueryParam("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 {
			pageSize = v
		}
	}
	pageToken := c.QueryParam("page_token")

	// Convert artifact type string to proto enum
	var artifactType pb.ArtifactType
	if artifactTypeStr != "" {
		switch artifactTypeStr {
		case "SOURCE_CODE":
			artifactType = pb.ArtifactType_SOURCE_CODE
		case "SCAN_RESULTS":
			artifactType = pb.ArtifactType_SCAN_RESULTS
		case "REPORT":
			artifactType = pb.ArtifactType_REPORT
		case "LOG":
			artifactType = pb.ArtifactType_LOG
		}
	}

	listReq := &pb.ListArtifactsRequest{
		ScanId:         scanID,
		OrganizationId: organizationID,
		Type:           artifactType,
		PageSize:       int32(pageSize),
		PageToken:      pageToken,
	}

	resp, err := h.storageClient.ListArtifacts(c.Request().Context(), listReq)
	if err != nil {
		log.WithError(err).Error("Failed to list artifacts via gRPC")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list artifacts")
	}

	return c.JSON(http.StatusOK, resp)
}