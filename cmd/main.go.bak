package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)

	log.WithFields(log.Fields{
		"version":   version,
		"commit":    commit,
		"buildDate": buildDate,
	}).Info("Starting CloudScan API Gateway")

	// Get configuration from environment
	port := getEnv("PORT", "8080")
	orchestratorURL := getEnv("ORCHESTRATOR_URL", "http://orchestrator:8081")
	storageURL := getEnv("STORAGE_URL", "http://storage:8082")
	websocketURL := getEnv("WEBSOCKET_URL", "http://websocket:9090")

	log.WithFields(log.Fields{
		"port":            port,
		"orchestratorURL": orchestratorURL,
		"storageURL":      storageURL,
		"websocketURL":    websocketURL,
	}).Info("API Gateway configuration loaded")

	// Start HTTP server
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Health check endpoints
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":    "healthy",
			"service":   "api-gateway",
			"version":   version,
			"commit":    commit,
			"buildDate": buildDate,
			"timestamp": time.Now().UTC(),
		})
	})

	e.GET("/ready", func(c echo.Context) error {
		// TODO: Check connectivity to backend services
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "ready",
		})
	})

	// API routes
	api := e.Group("/api/v1")

	// Scan management routes
	api.POST("/scans", func(c echo.Context) error {
		// TODO: Create new scan (proxy to orchestrator)
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Create scan endpoint - to be implemented",
		})
	})

	api.GET("/scans", func(c echo.Context) error {
		// TODO: List scans
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "List scans endpoint - to be implemented",
			"scans":   []interface{}{},
		})
	})

	api.GET("/scans/:id", func(c echo.Context) error {
		// TODO: Get scan details
		id := c.Param("id")
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Get scan endpoint - to be implemented",
			"scanId":  id,
		})
	})

	api.GET("/scans/:id/results", func(c echo.Context) error {
		// TODO: Get scan results (proxy to storage)
		id := c.Param("id")
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Get scan results endpoint - to be implemented",
			"scanId":  id,
		})
	})

	// Project management routes
	api.POST("/projects", func(c echo.Context) error {
		// TODO: Create project
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Create project endpoint - to be implemented",
		})
	})

	api.GET("/projects", func(c echo.Context) error {
		// TODO: List projects
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "List projects endpoint - to be implemented",
			"projects": []interface{}{},
		})
	})

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh

		log.Info("Shutting down API Gateway...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := e.Shutdown(ctx); err != nil {
			log.Errorf("Error during shutdown: %v", err)
		}
	}()

	addr := ":" + port
	log.Infof("API Gateway listening on %s", addr)
	if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func init() {
	// Set up logging
	logLevel := getEnv("LOG_LEVEL", "info")
	switch logLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	log.Infof("Log level set to: %s", logLevel)
}