package service

import (
	"errors"
	"time"

	"github.com/cloud-scan/cloudscan-apigateway/internal/models"
	"github.com/cloud-scan/cloudscan-apigateway/internal/repository"
	"github.com/cloud-scan/cloudscan-apigateway/internal/utils"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailExists        = errors.New("email already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserInactive       = errors.New("user is inactive")
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo *repository.UserRepository
	orgRepo  *repository.OrganizationRepository
	jwtMgr   *utils.JWTManager
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userRepo *repository.UserRepository,
	orgRepo *repository.OrganizationRepository,
	jwtMgr *utils.JWTManager,
) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		orgRepo:  orgRepo,
		jwtMgr:   jwtMgr,
	}
}

// SignupRequest represents a signup request
type SignupRequest struct {
	Email            string `json:"email" validate:"required,email"`
	Password         string `json:"password" validate:"required,min=8"`
	FirstName        string `json:"first_name" validate:"required"`
	LastName         string `json:"last_name" validate:"required"`
	OrganizationName string `json:"organization_name" validate:"required"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	User         *models.User `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int          `json:"expires_in"` // seconds
}

// Signup creates a new user and organization
func (s *AuthService) Signup(req *SignupRequest) (*AuthResponse, error) {
	// Check if email already exists
	existingUser, err := s.userRepo.FindByEmail(req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrEmailExists
	}

	// Hash password
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create organization
	org := &models.Organization{
		Name:        req.OrganizationName,
		DisplayName: req.OrganizationName,
		Plan:        "free",
		IsActive:    true,
		MaxProjects: 10,
		MaxUsers:    5,
	}

	if err := s.orgRepo.Create(org); err != nil {
		return nil, err
	}

	// Create user
	user := &models.User{
		Email:          req.Email,
		PasswordHash:   passwordHash,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Role:           "admin", // First user in org is admin
		OrganizationID: org.ID,
		IsActive:       true,
	}

	if err := s.userRepo.Create(user); err != nil {
		// Rollback organization creation if user creation fails
		_ = s.orgRepo.Delete(org.ID)
		return nil, err
	}

	// Load organization relationship
	user.Organization = *org

	// Generate tokens
	accessToken, err := s.jwtMgr.GenerateToken(user.ID, user.Email, user.OrganizationID, user.Role)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtMgr.GenerateRefreshToken(user.ID, user.Email, user.OrganizationID, user.Role)
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"user_id": user.ID,
		"email":   user.Email,
		"org_id":  user.OrganizationID,
	}).Info("User signed up successfully")

	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    24 * 3600, // 24 hours
	}, nil
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(req *LoginRequest) (*AuthResponse, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Verify password
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	// Update last login time
	now := time.Now()
	user.LastLoginAt = &now
	if err := s.userRepo.Update(user); err != nil {
		log.WithError(err).Warn("Failed to update last login time")
	}

	// Generate tokens
	accessToken, err := s.jwtMgr.GenerateToken(user.ID, user.Email, user.OrganizationID, user.Role)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtMgr.GenerateRefreshToken(user.ID, user.Email, user.OrganizationID, user.Role)
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"user_id": user.ID,
		"email":   user.Email,
	}).Info("User logged in successfully")

	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    24 * 3600, // 24 hours
	}, nil
}

// RefreshToken generates new tokens from a refresh token
func (s *AuthService) RefreshToken(refreshToken string) (*AuthResponse, error) {
	// Validate refresh token
	claims, err := s.jwtMgr.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Find user
	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Generate new tokens
	accessToken, err := s.jwtMgr.GenerateToken(user.ID, user.Email, user.OrganizationID, user.Role)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.jwtMgr.GenerateRefreshToken(user.ID, user.Email, user.OrganizationID, user.Role)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    24 * 3600,
	}, nil
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(userID string) (*models.User, error) {
	return s.userRepo.FindByID(userID)
}