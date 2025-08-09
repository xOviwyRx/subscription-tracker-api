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

// CreateWithTransaction creates a new subscription within a transaction
func (r *SubscriptionRepository) CreateWithTransaction(subscription *models.Subscription) error {
	r.logger.WithFields(logrus.Fields{
		"user_id":      subscription.UserID,
		"service_name": subscription.ServiceName,
		"start_date":   subscription.StartDate,
	}).Info("Starting subscription creation with transaction")

	return r.db.Transaction(func(tx *gorm.DB) error {
		// Check if subscription already exists for the same user and service
		var existing models.Subscription
		err := tx.Where("user_id = ? AND service_name = ? AND start_date = ?",
			subscription.UserID, subscription.ServiceName, subscription.StartDate).
			First(&existing).Error

		if err == nil {
			r.logger.WithFields(logrus.Fields{
				"user_id":      subscription.UserID,
				"service_name": subscription.ServiceName,
				"start_date":   subscription.StartDate,
			}).Info("Duplicate subscription detected, preventing creation")
			return errors.New("subscription already exists for this user and service in the same period")
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		// Create the subscription
		err = tx.Create(subscription).Error
		if err == nil {
			r.logger.WithFields(logrus.Fields{
				"subscription_id": subscription.ID,
				"user_id":         subscription.UserID,
				"service_name":    subscription.ServiceName,
			}).Info("Subscription created in database successfully")
		}
		return err
	})
}

// GetByID retrieves a subscription by ID
func (r *SubscriptionRepository) GetByID(id uint) (*models.Subscription, error) {
	r.logger.WithField("subscription_id", id).Info("Retrieving subscription by ID")

	var subscription models.Subscription
	err := r.db.First(&subscription, id).Error
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

// UpdateWithTransaction updates a subscription within a transaction
func (r *SubscriptionRepository) UpdateWithTransaction(subscription *models.Subscription) error {
	r.logger.WithField("subscription_id", subscription.ID).Info("Starting subscription update with transaction")

	return r.db.Transaction(func(tx *gorm.DB) error {
		// First check if the subscription exists
		var existing models.Subscription
		if err := tx.First(&existing, subscription.ID).Error; err != nil {
			return err
		}

		// Check for conflicts if service name or dates are being changed
		if subscription.ServiceName != existing.ServiceName ||
			subscription.StartDate != existing.StartDate {

			r.logger.WithFields(logrus.Fields{
				"subscription_id": subscription.ID,
				"old_service":     existing.ServiceName,
				"new_service":     subscription.ServiceName,
				"old_start_date":  existing.StartDate,
				"new_start_date":  subscription.StartDate,
			}).Info("Checking for conflicts due to service name or date changes")

			var conflicting models.Subscription
			err := tx.Where("user_id = ? AND service_name = ? AND start_date = ? AND id != ?",
				subscription.UserID, subscription.ServiceName, subscription.StartDate, subscription.ID).
				First(&conflicting).Error

			if err == nil {
				r.logger.WithFields(logrus.Fields{
					"subscription_id": subscription.ID,
					"conflicting_id":  conflicting.ID,
					"service_name":    subscription.ServiceName,
					"start_date":      subscription.StartDate,
				}).Info("Conflict detected, preventing update")
				return errors.New("another subscription already exists for this user and service in the same period")
			}

			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}

		// Update the subscription
		err := tx.Save(subscription).Error
		if err == nil {
			r.logger.WithFields(logrus.Fields{
				"subscription_id": subscription.ID,
				"service_name":    subscription.ServiceName,
			}).Info("Subscription updated in database successfully")
		}
		return err
	})
}

// DeleteWithValidation deletes a subscription with validation within a transaction
func (r *SubscriptionRepository) DeleteWithValidation(id uint) error {
	r.logger.WithField("subscription_id", id).Info("Starting subscription deletion with validation")

	return r.db.Transaction(func(tx *gorm.DB) error {
		// First check if the subscription exists
		var subscription models.Subscription
		if err := tx.First(&subscription, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				r.logger.WithField("subscription_id", id).Info("Subscription not found for deletion")
				return errors.New("subscription not found")
			}
			return err
		}

		r.logger.WithFields(logrus.Fields{
			"subscription_id": id,
			"service_name":    subscription.ServiceName,
			"user_id":         subscription.UserID,
		}).Info("Subscription found, proceeding with deletion")

		// Perform the deletion
		err := tx.Delete(&subscription).Error
		if err == nil {
			r.logger.WithField("subscription_id", id).Info("Subscription deleted from database successfully")
		}
		return err
	})
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
