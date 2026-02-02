package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cloud-scan/cloudscan-apigateway/internal/api"
	"github.com/cloud-scan/cloudscan-apigateway/internal/api/handlers"
	"github.com/cloud-scan/cloudscan-apigateway/internal/config"
	grpcClient "github.com/cloud-scan/cloudscan-apigateway/internal/grpc"
	"github.com/cloud-scan/cloudscan-apigateway/internal/repository"
	"github.com/cloud-scan/cloudscan-apigateway/internal/service"
	"github.com/cloud-scan/cloudscan-apigateway/internal/utils"
	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

func main() {
	// Configure logging
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)

	log.WithFields(log.Fields{
		"version":   version,
		"commit":    commit,
		"buildDate": buildDate,
	}).Info("Starting CloudScan API Gateway")

	// Load configuration
	cfg := config.Load()

	// Set log level from config
	setLogLevel(cfg.Log.Level)

	// Initialize database connection
	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Info("Database connection established")

	// Note: Database migrations are handled externally via Kubernetes migration job
	// See migrations/001_initial_schema.up.sql
	log.Info("Skipping migrations (handled externally)")

	// Initialize Redis connection
	redisClient := initRedis(cfg)
	log.Info("Redis connection established")

	// Initialize gRPC clients
	orchestratorClient, err := grpcClient.NewOrchestratorClient(cfg.Services.OrchestratorGRPC)
	if err != nil {
		log.Fatalf("Failed to connect to Orchestrator gRPC: %v", err)
	}
	log.Info("Orchestrator gRPC connection established")

	storageClient, err := grpcClient.NewStorageClient(cfg.Services.StorageGRPC)
	if err != nil {
		log.Fatalf("Failed to connect to Storage gRPC: %v", err)
	}
	log.Info("Storage gRPC connection established")

	// Initialize JWT manager
	jwtMgr := utils.NewJWTManager(
		cfg.JWT.Secret,
		cfg.JWT.ExpirationHours,
		cfg.JWT.RefreshHours,
	)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	orgRepo := repository.NewOrganizationRepository(db)
	projectRepo := repository.NewProjectRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, orgRepo, jwtMgr)
	orgService := service.NewOrganizationService(orgRepo)
	projectService := service.NewProjectService(projectRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	orgHandler := handlers.NewOrganizationHandler(orgService)
	projectHandler := handlers.NewProjectHandler(projectService)
	scanHandler := handlers.NewScanHandler(orchestratorClient, projectRepo)
	storageHandler := handlers.NewStorageHandler(storageClient)

	// Initialize Echo server
	e := echo.New()
	e.HideBanner = true

	// Setup routes
	api.SetupRoutes(
		e,
		cfg,
		jwtMgr,
		redisClient,
		authHandler,
		orgHandler,
		projectHandler,
		scanHandler,
		storageHandler,
	)

	// Start server in goroutine
	go func() {
		addr := ":" + cfg.Server.Port
		log.Infof("API Gateway listening on %s", addr)
		if err := e.Start(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down API Gateway...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Close gRPC connections
	if err := orchestratorClient.Close(); err != nil {
		log.Errorf("Error closing Orchestrator gRPC connection: %v", err)
	}
	if err := storageClient.Close(); err != nil {
		log.Errorf("Error closing Storage gRPC connection: %v", err)
	}
	log.Info("gRPC connections closed")

	// Shutdown HTTP server
	if err := e.Shutdown(ctx); err != nil {
		log.Errorf("Error during shutdown: %v", err)
	}

	log.Info("API Gateway stopped")
}

// initDatabase initializes the database connection
func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.Database.GetDSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Get underlying SQL DB for connection pooling
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(cfg.Database.MaxConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MinConns)

	return db, nil
}


// initRedis initializes Redis client
func initRedis(cfg *config.Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.GetAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Warnf("Failed to connect to Redis: %v. Rate limiting will be disabled.", err)
	}

	return client
}

// setLogLevel sets the log level from config
func setLogLevel(level string) {
	switch level {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
}