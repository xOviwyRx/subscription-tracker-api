package repository

import (
	"errors"
	"subscription_tracker_api/internal/models"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// SubscriptionRepository handles database operations for subscriptions
type SubscriptionRepository struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// NewSubscriptionRepository creates a new subscription repository
func NewSubscriptionRepository(db *gorm.DB, logger *logrus.Logger) *SubscriptionRepository {
	return &SubscriptionRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new subscription
func (r *SubscriptionRepository) Create(tx *gorm.DB, subscription *models.Subscription) error {
	db := r.getDB(tx)
	return db.Create(subscription).Error
}

// GetByID retrieves a subscription by ID
func (r *SubscriptionRepository) GetByID(tx *gorm.DB, id uint) (*models.Subscription, error) {
	r.logger.WithField("subscription_id", id).Info("Retrieving subscription by ID")

	db := r.getDB(tx)
	var subscription models.Subscription
	err := db.First(&subscription, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.WithField("subscription_id", id).Info("Subscription not found in database")
		}
		return nil, err
	}

	r.logger.WithFields(logrus.Fields{
		"subscription_id": id,
		"service_name":    subscription.ServiceName,
		"user_id":         subscription.UserID,
	}).Info("Subscription retrieved from database successfully")

	return &subscription, nil
}

// Update updates a subscription
func (r *SubscriptionRepository) Update(tx *gorm.DB, subscription *models.Subscription) error {
	db := r.getDB(tx)
	return db.Save(subscription).Error
}

// Delete deletes a subscription
func (r *SubscriptionRepository) Delete(tx *gorm.DB, id uint) error {
	db := r.getDB(tx)
	return db.Delete(&models.Subscription{}, id).Error
}

// List retrieves all subscriptions with optional filtering
func (r *SubscriptionRepository) List(userID *uuid.UUID, serviceName *string, limit, offset int) ([]models.Subscription, error) {
	r.logger.WithFields(logrus.Fields{
		"user_id":      userID,
		"service_name": serviceName,
		"limit":        limit,
		"offset":       offset,
	}).Info("Retrieving subscriptions list with filters")

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
	if err == nil {
		r.logger.WithFields(logrus.Fields{
			"subscription_count": len(subscriptions),
			"user_id":            userID,
			"service_name":       serviceName,
		}).Info("Subscriptions list retrieved from database successfully")
	}

	return subscriptions, err
}

// GetSubscriptionsInDateRange retrieves subscriptions that overlap with the given date range
func (r *SubscriptionRepository) GetSubscriptionsInDateRange(userID *uuid.UUID, serviceName *string, startDate, endDate string) ([]models.Subscription, error) {
	r.logger.WithFields(logrus.Fields{
		"user_id":      userID,
		"service_name": serviceName,
		"start_date":   startDate,
		"end_date":     endDate,
	}).Info("Retrieving subscriptions in date range")

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
	if err == nil {
		r.logger.WithFields(logrus.Fields{
			"subscription_count": len(subscriptions),
			"date_range":         startDate + " to " + endDate,
			"user_id":            userID,
			"service_name":       serviceName,
		}).Info("Subscriptions in date range retrieved from database successfully")
	}

	return subscriptions, err
}

// CalculateTotalCostInDB performs cost calculation with database aggregation
func (r *SubscriptionRepository) CalculateTotalCostInDB(userID *uuid.UUID, serviceName *string, startDate, endDate string, totalMonths int) (int, error) {
	var result struct {
		TotalCost int `gorm:"column:total_cost"`
	}

	query := r.db.Model(&models.Subscription{}).
		Where("start_date <= ? AND (end_date IS NULL OR end_date >= ?)", endDate, startDate)

	// Apply filters
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if serviceName != nil {
		query = query.Where("service_name = ?", *serviceName)
	}

	// Database aggregation with month consideration
	err := query.Select("COALESCE(SUM(price * ?), 0) as total_cost", totalMonths).Scan(&result).Error
	if err != nil {
		return 0, err
	}

	return result.TotalCost, nil
}

// Helper to get the correct DB instance (transaction or regular)
func (r *SubscriptionRepository) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

// ExistsByUserServiceAndDate checks for duplicate subscriptions
func (r *SubscriptionRepository) ExistsByUserServiceAndDate(tx *gorm.DB, userID uuid.UUID, serviceName, startDate string) (bool, error) {
	db := r.getDB(tx)
	var count int64
	err := db.Model(&models.Subscription{}).
		Where("user_id = ? AND service_name = ? AND start_date = ?", userID, serviceName, startDate).
		Count(&count).Error
	return count > 0, err
}

// ExistsByID checks if subscription exists
func (r *SubscriptionRepository) ExistsByID(tx *gorm.DB, id uint) (bool, error) {
	db := r.getDB(tx)
	var count int64
	err := db.Model(&models.Subscription{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}
