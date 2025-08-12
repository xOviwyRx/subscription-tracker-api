package service

import (
	"errors"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"regexp"
	"strconv"
	"strings"
	"subscription_tracker_api/internal/models"
	"subscription_tracker_api/internal/repository"
)

type SubscriptionService struct {
	repo   repository.SubscriptionRepositoryInterface
	db     *gorm.DB // for transaction control
	logger *logrus.Logger
}

func NewSubscriptionService(repo repository.SubscriptionRepositoryInterface, db *gorm.DB, logger *logrus.Logger) *SubscriptionService {
	return &SubscriptionService{
		repo:   repo,
		db:     db,
		logger: logger,
	}
}

// CreateSubscription creates a new subscription with transaction-based validation
func (s *SubscriptionService) CreateSubscription(req *models.CreateSubscriptionRequest) (*models.Subscription, error) {
	// Enhanced validation
	if req.ServiceName == "" || req.Price <= 0 || req.UserID == uuid.Nil {
		return nil, errors.New("invalid input data: service_name, price, and user_id are required")
	}

	if !isValidDate(req.StartDate) {
		return nil, errors.New("start_date must be in MM-YYYY format")
	}

	// Validate end_date if provided
	if req.EndDate != nil && *req.EndDate != "" {
		if !isValidDate(*req.EndDate) {
			return nil, errors.New("end_date must be in MM-YYYY format")
		}

		// Check that end_date is after start_date
		if *req.EndDate <= req.StartDate {
			return nil, errors.New("end_date must be after start_date")
		}
	}

	// Start transaction at service level
	tx := s.db.Begin()
	if tx.Error != nil {
		s.logger.WithError(tx.Error).Error("Failed to begin transaction")
		return nil, errors.New("failed to start transaction")
	}

	// Safe defer: Rollback if panic occurs
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			s.logger.Error("Recovered from panic, transaction rolled back")
		}
	}()

	// Check for duplicates
	exists, err := s.repo.ExistsByUserServiceAndDate(tx, req.UserID, req.ServiceName, req.StartDate)
	if err != nil {
		s.logger.WithError(err).Error("Failed to check for duplicate subscription")
		return nil, errors.New("failed to validate subscription uniqueness")
	}

	if exists {
		return nil, errors.New("subscription already exists for this user, service, and date")
	}

	subscription := &models.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}

	// Create subscription using transaction
	err = s.repo.Create(tx, subscription)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":      req.UserID,
			"service_name": req.ServiceName,
			"start_date":   req.StartDate,
		}).Error("Failed to create subscription")
		return nil, errors.New("failed to create subscription")
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		s.logger.WithError(err).Error("Failed to commit transaction")
		return nil, errors.New("failed to save subscription")
	}

	s.logger.WithFields(logrus.Fields{
		"subscription_id": subscription.ID,
		"user_id":         req.UserID,
		"service_name":    req.ServiceName,
	}).Info("Subscription created successfully")

	return subscription, nil
}

// GetSubscriptionByID retrieves a subscription by ID
func (s *SubscriptionService) GetSubscriptionByID(id uint) (*models.Subscription, error) {
	return s.repo.GetByID(id)
}

// UpdateSubscription updates an existing subscription with transaction-based validation
func (s *SubscriptionService) UpdateSubscription(id uint, updates map[string]interface{}) (*models.Subscription, error) {
	// Get the current subscription first
	subscription, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("subscription not found")
	}

	// Track what's being updated for logging
	updatedFields := make(map[string]interface{})

	// Apply and validate updates
	if serviceName, ok := updates["service_name"].(string); ok && serviceName != "" {
		if serviceName != subscription.ServiceName {
			subscription.ServiceName = serviceName
			updatedFields["service_name"] = serviceName
		}
	}

	if price, ok := updates["price"].(float64); ok {
		if price <= 0 {
			return nil, errors.New("price must be greater than 0")
		}
		newPrice := int(price)
		if newPrice != subscription.Price {
			subscription.Price = newPrice
			updatedFields["price"] = newPrice
		}
	}

	if startDate, ok := updates["start_date"].(string); ok && startDate != "" {
		if !isValidDate(startDate) {
			return nil, errors.New("start_date must be in MM-YYYY format")
		}
		if startDate != subscription.StartDate {
			// Check if end_date would still be valid
			if subscription.EndDate != nil && *subscription.EndDate <= startDate {
				return nil, errors.New("start_date must be before end_date")
			}
			subscription.StartDate = startDate
			updatedFields["start_date"] = startDate
		}
	}

	if endDate, ok := updates["end_date"].(string); ok {
		if endDate != "" {
			if !isValidDate(endDate) {
				return nil, errors.New("end_date must be in MM-YYYY format")
			}
			if endDate <= subscription.StartDate {
				return nil, errors.New("end_date must be after start_date")
			}
			subscription.EndDate = &endDate
			updatedFields["end_date"] = endDate
		} else {
			subscription.EndDate = nil
			updatedFields["end_date"] = nil
		}
	}

	// Use transaction-based update for conflict checking
	err = s.repo.UpdateWithTransaction(subscription)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"subscription_id": id,
			"updates":         updatedFields,
		}).Error("Failed to update subscription with transaction")
		return nil, err
	}

	s.logger.WithFields(logrus.Fields{
		"subscription_id": id,
		"updated_fields":  updatedFields,
	}).Info("Subscription updated successfully")

	return subscription, nil
}

// DeleteSubscription deletes a subscription with validation
func (s *SubscriptionService) DeleteSubscription(id uint) error {
	err := s.repo.DeleteWithValidation(id)
	if err != nil {
		s.logger.WithError(err).WithField("subscription_id", id).Error("Failed to delete subscription")
		return err
	}

	s.logger.WithField("subscription_id", id).Info("Subscription deleted successfully")
	return nil
}

// ListSubscriptions retrieves subscriptions with optional filtering
func (s *SubscriptionService) ListSubscriptions(userID *uuid.UUID, serviceName *string, limit, offset int) ([]models.Subscription, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.List(userID, serviceName, limit, offset)
}

// CalculateTotalCost calculates total cost with proper month consideration and database aggregation
func (s *SubscriptionService) CalculateTotalCost(req *models.CostCalculationRequest) (*models.CostCalculationResponse, error) {
	// Validate date formats
	if !isValidDate(req.StartDate) {
		return nil, errors.New("start_date must be in MM-YYYY format")
	}
	if !isValidDate(req.EndDate) {
		return nil, errors.New("end_date must be in MM-YYYY format")
	}

	// Validate date range
	if req.EndDate <= req.StartDate {
		return nil, errors.New("end_date must be after start_date")
	}

	// Calculate total months in requested period
	totalMonths := calculateMonthsBetween(req.StartDate, req.EndDate)

	// Use repository method for database aggregation
	totalCost, err := s.repo.CalculateTotalCostInDB(req.UserID, req.ServiceName, req.StartDate, req.EndDate, totalMonths)
	if err != nil {
		s.logger.WithError(err).Error("Failed to calculate total cost in database")
		return nil, err
	}

	// Get subscriptions for response details
	subscriptions, err := s.repo.GetSubscriptionsInDateRange(req.UserID, req.ServiceName, req.StartDate, req.EndDate)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get subscriptions in date range")
		return nil, err
	}

	response := &models.CostCalculationResponse{
		TotalCost:     totalCost,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		UserID:        req.UserID,
		ServiceName:   req.ServiceName,
		Subscriptions: subscriptions,
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":            req.UserID,
		"service_name":       req.ServiceName,
		"start_date":         req.StartDate,
		"end_date":           req.EndDate,
		"total_cost":         totalCost,
		"total_months":       totalMonths,
		"subscription_count": len(subscriptions),
	}).Info("Total cost calculated with database aggregation")

	return response, nil
}

// Helper function to calculate months between MM-YYYY dates
func calculateMonthsBetween(startDate, endDate string) int {
	startParts := strings.Split(startDate, "-")
	endParts := strings.Split(endDate, "-")

	startMonth, _ := strconv.Atoi(startParts[0])
	startYear, _ := strconv.Atoi(startParts[1])
	endMonth, _ := strconv.Atoi(endParts[0])
	endYear, _ := strconv.Atoi(endParts[1])

	return (endYear-startYear)*12 + (endMonth - startMonth) + 1
}

// Simple date validation helper
func isValidDate(date string) bool {
	matched, _ := regexp.MatchString(`^(0[1-9]|1[0-2])-[0-9]{4}$`, date)
	return matched
}
