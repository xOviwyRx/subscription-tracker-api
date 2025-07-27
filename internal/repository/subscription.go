package repository

import (
	"subscription_tracker_api/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SubscriptionRepository handles database operations for subscriptions
type SubscriptionRepository struct {
	db *gorm.DB
}

// NewSubscriptionRepository creates a new subscription repository
func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// Create creates a new subscription
func (r *SubscriptionRepository) Create(subscription *models.Subscription) error {
	return r.db.Create(subscription).Error
}

// GetByID retrieves a subscription by ID
func (r *SubscriptionRepository) GetByID(id uint) (*models.Subscription, error) {
	var subscription models.Subscription
	err := r.db.First(&subscription, id).Error
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

// Update updates a subscription
func (r *SubscriptionRepository) Update(subscription *models.Subscription) error {
	return r.db.Save(subscription).Error
}

// Delete deletes a subscription (soft delete)
func (r *SubscriptionRepository) Delete(id uint) error {
	return r.db.Delete(&models.Subscription{}, id).Error
}

// List retrieves all subscriptions with optional filtering
func (r *SubscriptionRepository) List(userID *uuid.UUID, serviceName *string, limit, offset int) ([]models.Subscription, error) {
	var subscriptions []models.Subscription
	query := r.db.Model(&models.Subscription{})

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	if serviceName != nil {
		query = query.Where("service_name ILIKE ?", "%"+*serviceName+"%")
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&subscriptions).Error
	return subscriptions, err
}

// GetSubscriptionsInDateRange retrieves subscriptions that overlap with the given date range
func (r *SubscriptionRepository) GetSubscriptionsInDateRange(userID *uuid.UUID, serviceName *string, startDate, endDate string) ([]models.Subscription, error) {
	var subscriptions []models.Subscription
	query := r.db.Model(&models.Subscription{})

	// Filter by user ID if provided
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	// Filter by service name if provided
	if serviceName != nil {
		query = query.Where("service_name ILIKE ?", "%"+*serviceName+"%")
	}

	// Filter by date range - subscriptions that overlap with the given period
	query = query.Where(
		"(start_date <= ? AND (end_date IS NULL OR end_date >= ?))",
		endDate, startDate,
	)

	err := query.Find(&subscriptions).Error
	return subscriptions, err
}
