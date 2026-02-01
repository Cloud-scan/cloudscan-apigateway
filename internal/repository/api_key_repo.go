package repository

import (
	"time"

	"github.com/cloud-scan/cloudscan-apigateway/internal/models"
	"gorm.io/gorm"
)

// APIKeyRepository handles API key data access
type APIKeyRepository struct {
	db *gorm.DB
}

// NewAPIKeyRepository creates a new API key repository
func NewAPIKeyRepository(db *gorm.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

// Create creates a new API key
func (r *APIKeyRepository) Create(apiKey *models.APIKey) error {
	return r.db.Create(apiKey).Error
}

// FindByID finds an API key by ID
func (r *APIKeyRepository) FindByID(id string) (*models.APIKey, error) {
	var apiKey models.APIKey
	err := r.db.Preload("Organization").Where("id = ?", id).First(&apiKey).Error
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

// FindByKeyHash finds an API key by its hash
func (r *APIKeyRepository) FindByKeyHash(keyHash string) (*models.APIKey, error) {
	var apiKey models.APIKey
	err := r.db.Preload("Organization").Where("key_hash = ? AND is_active = ?", keyHash, true).First(&apiKey).Error
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

// Update updates an API key
func (r *APIKeyRepository) Update(apiKey *models.APIKey) error {
	return r.db.Save(apiKey).Error
}

// Delete soft deletes an API key
func (r *APIKeyRepository) Delete(id string) error {
	return r.db.Delete(&models.APIKey{}, "id = ?", id).Error
}

// ListByOrganization lists all API keys for an organization
func (r *APIKeyRepository) ListByOrganization(organizationID string, limit, offset int) ([]*models.APIKey, int64, error) {
	var apiKeys []*models.APIKey
	var total int64

	query := r.db.Where("organization_id = ?", organizationID)

	// Get total count
	if err := query.Model(&models.APIKey{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Limit(limit).Offset(offset).Find(&apiKeys).Error
	if err != nil {
		return nil, 0, err
	}

	return apiKeys, total, nil
}

// UpdateLastUsed updates the last used timestamp for an API key
func (r *APIKeyRepository) UpdateLastUsed(id string) error {
	now := time.Now()
	return r.db.Model(&models.APIKey{}).Where("id = ?", id).Update("last_used_at", &now).Error
}