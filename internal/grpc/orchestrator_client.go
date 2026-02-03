package grpc

import (
	"context"
	"fmt"
	"time"

	pb "github.com/cloud-scan/cloudscan-apigateway/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	log "github.com/sirupsen/logrus"
)

// OrchestratorClient wraps the gRPC client for the Orchestrator service
type OrchestratorClient struct {
	conn   *grpc.ClientConn
	client pb.ScanServiceClient
}

// NewOrchestratorClient creates a new Orchestrator gRPC client
func NewOrchestratorClient(address string) (*OrchestratorClient, error) {
	log.WithField("address", address).Info("Connecting to Orchestrator service")

	// Configure keep-alive
	kaParams := keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             3 * time.Second,
		PermitWithoutStream: true,
	}

	// Create connection
	conn, err := grpc.Dial(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(kaParams),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Orchestrator: %w", err)
	}

	client := pb.NewScanServiceClient(conn)

	log.Info("Successfully connected to Orchestrator service")
	return &OrchestratorClient{
		conn:   conn,
		client: client,
	}, nil
}

// CreateScan creates a new security scan
func (c *OrchestratorClient) CreateScan(ctx context.Context, req *pb.CreateScanRequest) (*pb.Scan, error) {
	resp, err := c.client.CreateScan(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create scan: %w", err)
	}
	return resp.Scan, nil
}

// GetScan retrieves a scan by ID
func (c *OrchestratorClient) GetScan(ctx context.Context, id string) (*pb.Scan, error) {
	req := &pb.GetScanRequest{Id: id}
	return c.client.GetScan(ctx, req)
}

// ListScans lists scans with filters
func (c *OrchestratorClient) ListScans(ctx context.Context, req *pb.ListScansRequest) (*pb.ListScansResponse, error) {
	return c.client.ListScans(ctx, req)
}

// CancelScan cancels a running scan
func (c *OrchestratorClient) CancelScan(ctx context.Context, id string) error {
	req := &pb.CancelScanRequest{Id: id}
	_, err := c.client.CancelScan(ctx, req)
	return err
}

// GetFindings retrieves findings for a scan
func (c *OrchestratorClient) GetFindings(ctx context.Context, req *pb.GetFindingsRequest) (*pb.GetFindingsResponse, error) {
	return c.client.GetFindings(ctx, req)
}

// DeleteScan deletes a scan and all its data (findings, artifacts, k8s job)
func (c *OrchestratorClient) DeleteScan(ctx context.Context, id string) error {
	req := &pb.DeleteScanRequest{Id: id}
	_, err := c.client.DeleteScan(ctx, req)
	return err
}

// DeleteProjectScans deletes all scans for a project
func (c *OrchestratorClient) DeleteProjectScans(ctx context.Context, projectID string) (int32, error) {
	req := &pb.DeleteProjectScansRequest{ProjectId: projectID}
	resp, err := c.client.DeleteProjectScans(ctx, req)
	if err != nil {
		return 0, err
	}
	return resp.DeletedCount, nil
}

// Close closes the gRPC connection
func (c *OrchestratorClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}