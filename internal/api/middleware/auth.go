package middleware

import (
	"net/http"
	"strings"

	"github.com/cloud-scan/cloudscan-apigateway/internal/utils"
	"github.com/labstack/echo/v4"
)

// JWTMiddleware validates JWT tokens and sets user context
func JWTMiddleware(jwtMgr *utils.JWTManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			// Extract token from "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header format")
			}

			tokenString := parts[1]

			// Validate token
			claims, err := jwtMgr.ValidateToken(tokenString)
			if err != nil {
				if err == utils.ErrExpiredToken {
					return echo.NewHTTPError(http.StatusUnauthorized, "token expired")
				}
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			// Set claims in context
			c.Set("user_id", claims.UserID)
			c.Set("email", claims.Email)
			c.Set("organization_id", claims.OrganizationID)
			c.Set("role", claims.Role)

			return next(c)
		}
	}
}

// GetUserID extracts user ID from context
func GetUserID(c echo.Context) string {
	userID, ok := c.Get("user_id").(string)
	if !ok {
		return ""
	}
	return userID
}

// GetOrganizationID extracts organization ID from context
func GetOrganizationID(c echo.Context) string {
	orgID, ok := c.Get("organization_id").(string)
	if !ok {
		return ""
	}
	return orgID
}

// GetRole extracts user role from context
func GetRole(c echo.Context) string {
	role, ok := c.Get("role").(string)
	if !ok {
		return ""
	}
	return role
}