package handlers

import (
	"net/http"

	"github.com/cloud-scan/cloudscan-apigateway/internal/api/middleware"
	"github.com/cloud-scan/cloudscan-apigateway/internal/service"
	"github.com/labstack/echo/v4"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Signup handles user registration
// POST /api/v1/auth/signup
func (h *AuthHandler) Signup(c echo.Context) error {
	var req service.SignupRequest

	if err := middleware.BindAndValidate(c, &req); err != nil {
		return err
	}

	authResp, err := h.authService.Signup(&req)
	if err != nil {
		if err == service.ErrEmailExists {
			return echo.NewHTTPError(http.StatusConflict, "email already exists")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, authResp)
}

// Login handles user authentication
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c echo.Context) error {
	var req service.LoginRequest

	if err := middleware.BindAndValidate(c, &req); err != nil {
		return err
	}

	authResp, err := h.authService.Login(&req)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid email or password")
		}
		if err == service.ErrUserInactive {
			return echo.NewHTTPError(http.StatusForbidden, "user is inactive")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, authResp)
}

// RefreshToken handles token refresh
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}

	if err := middleware.BindAndValidate(c, &req); err != nil {
		return err
	}

	authResp, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired refresh token")
	}

	return c.JSON(http.StatusOK, authResp)
}

// GetCurrentUser returns the currently authenticated user
// GET /api/v1/auth/me
func (h *AuthHandler) GetCurrentUser(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	return c.JSON(http.StatusOK, user)
}

// Logout handles user logout
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c echo.Context) error {
	// For JWT-based auth, logout is handled client-side by removing the token
	// Here we just return success
	// In the future, we could implement token blacklisting using Redis

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "logged out successfully",
	})
}