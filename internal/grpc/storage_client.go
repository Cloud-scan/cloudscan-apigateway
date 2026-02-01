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

// StorageClient wraps the gRPC client for the Storage service
type StorageClient struct {
	conn   *grpc.ClientConn
	client pb.StorageServiceClient
}

// NewStorageClient creates a new Storage gRPC client
func NewStorageClient(address string) (*StorageClient, error) {
	log.WithField("address", address).Info("Connecting to Storage service")

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
		return nil, fmt.Errorf("failed to connect to Storage: %w", err)
	}

	client := pb.NewStorageServiceClient(conn)

	log.Info("Successfully connected to Storage service")
	return &StorageClient{
		conn:   conn,
		client: client,
	}, nil
}

// CreateArtifact creates a new artifact and returns presigned upload URL
func (c *StorageClient) CreateArtifact(ctx context.Context, req *pb.CreateArtifactRequest) (*pb.CreateArtifactResponse, error) {
	return c.client.CreateArtifact(ctx, req)
}

// GetArtifact retrieves artifact metadata and presigned download URL
func (c *StorageClient) GetArtifact(ctx context.Context, id string, expiresInHours int32) (*pb.GetArtifactResponse, error) {
	req := &pb.GetArtifactRequest{
		Id:              id,
		ExpiresInHours: expiresInHours,
	}
	return c.client.GetArtifact(ctx, req)
}

// DeleteArtifact deletes an artifact
func (c *StorageClient) DeleteArtifact(ctx context.Context, id string) error {
	req := &pb.DeleteArtifactRequest{Id: id}
	_, err := c.client.DeleteArtifact(ctx, req)
	return err
}

// ListArtifacts lists artifacts with filters
func (c *StorageClient) ListArtifacts(ctx context.Context, req *pb.ListArtifactsRequest) (*pb.ListArtifactsResponse, error) {
	return c.client.ListArtifacts(ctx, req)
}

// InitiateMultipartUpload starts a multipart upload
func (c *StorageClient) InitiateMultipartUpload(ctx context.Context, artifactID string) (*pb.InitiateMultipartResponse, error) {
	req := &pb.InitiateMultipartRequest{ArtifactId: artifactID}
	return c.client.InitiateMultipartUpload(ctx, req)
}

// GetMultipartUploadParts gets presigned URLs for multipart upload parts
func (c *StorageClient) GetMultipartUploadParts(ctx context.Context, req *pb.GetMultipartPartsRequest) (*pb.GetMultipartPartsResponse, error) {
	return c.client.GetMultipartUploadParts(ctx, req)
}

// CompleteMultipartUpload completes a multipart upload
func (c *StorageClient) CompleteMultipartUpload(ctx context.Context, req *pb.CompleteMultipartRequest) (*pb.CompleteMultipartResponse, error) {
	return c.client.CompleteMultipartUpload(ctx, req)
}

// AbortMultipartUpload aborts a multipart upload
func (c *StorageClient) AbortMultipartUpload(ctx context.Context, artifactID, uploadID string) error {
	req := &pb.AbortMultipartRequest{
		ArtifactId: artifactID,
		UploadId:   uploadID,
	}
	_, err := c.client.AbortMultipartUpload(ctx, req)
	return err
}

// Close closes the gRPC connection
func (c *StorageClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}