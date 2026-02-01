package api

import (
	"github.com/cloud-scan/cloudscan-apigateway/internal/api/handlers"
	"github.com/cloud-scan/cloudscan-apigateway/internal/api/middleware"
	"github.com/cloud-scan/cloudscan-apigateway/internal/config"
	"github.com/cloud-scan/cloudscan-apigateway/internal/utils"
	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

// SetupRoutes configures all API routes
func SetupRoutes(
	e *echo.Echo,
	cfg *config.Config,
	jwtMgr *utils.JWTManager,
	redisClient *redis.Client,
	authHandler *handlers.AuthHandler,
	orgHandler *handlers.OrganizationHandler,
	projectHandler *handlers.ProjectHandler,
	scanHandler *handlers.ScanHandler,
	storageHandler *handlers.StorageHandler,
) {
	// Global middleware
	e.Use(echomiddleware.Recover())
	e.Use(middleware.RequestLogger())
	e.Use(middleware.CORSConfig(cfg.Server.Environment))

	// Health check endpoints (no auth required)
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]interface{}{
			"status":  "healthy",
			"service": "api-gateway",
		})
	})

	e.GET("/ready", func(c echo.Context) error {
		// TODO: Check database, Redis, and gRPC connections
		return c.JSON(200, map[string]interface{}{
			"status": "ready",
		})
	})

	// API v1 routes
	v1 := e.Group("/api/v1")

	// Rate limiting (applied to all API routes)
	if cfg.RateLimit.Enabled {
		v1.Use(middleware.RateLimiter(redisClient, cfg.RateLimit.RequestsPerMinute))
	}

	// Public auth routes (no JWT required)
	auth := v1.Group("/auth")
	{
		auth.POST("/signup", authHandler.Signup)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
	}

	// Protected routes (JWT required)
	protected := v1.Group("")
	protected.Use(middleware.JWTMiddleware(jwtMgr))

	// Auth endpoints (requires authentication)
	{
		protected.GET("/auth/me", authHandler.GetCurrentUser)
		protected.POST("/auth/logout", authHandler.Logout)
	}

	// Organization endpoints
	orgs := protected.Group("/organizations")
	{
		orgs.GET("", orgHandler.List)             // List all (superadmin only)
		orgs.GET("/:id", orgHandler.GetByID)      // Get organization
		orgs.PUT("/:id", orgHandler.Update)       // Update organization
	}

	// Project endpoints
	projects := protected.Group("/projects")
	{
		projects.POST("", projectHandler.Create)       // Create project
		projects.GET("", projectHandler.List)          // List projects
		projects.GET("/:id", projectHandler.GetByID)   // Get project
		projects.PUT("/:id", projectHandler.Update)    // Update project
		projects.DELETE("/:id", projectHandler.Delete) // Delete project
	}

	// Scan endpoints (proxy to Orchestrator)
	scans := protected.Group("/scans")
	{
		scans.POST("", scanHandler.CreateScan)                // Create scan
		scans.GET("", scanHandler.ListScans)                  // List scans
		scans.GET("/:id", scanHandler.GetScan)                // Get scan
		scans.PUT("/:id/cancel", scanHandler.CancelScan)      // Cancel scan
		scans.GET("/:id/findings", scanHandler.GetFindings)   // Get findings
	}

	// Storage endpoints (proxy to Storage service)
	storage := protected.Group("/storage")
	{
		storage.POST("/upload", storageHandler.Upload)        // Upload file
		storage.GET("/download/:id", storageHandler.Download) // Download file
		storage.DELETE("/:id", storageHandler.Delete)         // Delete file
	}
}