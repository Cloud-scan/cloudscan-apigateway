package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// CORSConfig returns CORS middleware configuration
func CORSConfig(environment string) echo.MiddlewareFunc {
	config := middleware.CORSConfig{
		AllowOrigins: []string{"*"}, // Will be restricted in production
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.PATCH, echo.OPTIONS},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
		MaxAge: 3600,
	}

	// In production, restrict origins
	if environment == "production" {
		config.AllowOrigins = []string{
			"https://cloudscan.io",
			"https://app.cloudscan.io",
		}
	} else if environment == "development" {
		config.AllowOrigins = []string{
			"http://localhost:3000",
			"http://localhost:5173", // Vite dev server
			"http://127.0.0.1:3000",
			"http://127.0.0.1:5173",
		}
	}

	return middleware.CORSWithConfig(config)
}