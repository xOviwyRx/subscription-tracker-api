package service

import (
	"errors"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"regexp"
	"subscription_tracker_api/internal/models"
	"subscription_tracker_api/internal/repository"
)

type SubscriptionService struct {
	repo   *repository.SubscriptionRepository
	logger *logrus.Logger
}

func NewSubscriptionService(repo *repository.SubscriptionRepository, logger *logrus.Logger) *SubscriptionService {
	return &SubscriptionService{repo: repo, logger: logger}
}

// CreateSubscriptionWithTransaction creates a new subscription with transaction-based validation
func (s *SubscriptionService) CreateSubscriptionWithTransaction(req *models.CreateSubscriptionRequest) (*models.Subscription, error) {
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

	subscription := &models.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}

	// Use transaction-based creation for duplicate checking
	err := s.repo.CreateWithTransaction(subscription)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":      req.UserID,
			"service_name": req.ServiceName,
			"start_date":   req.StartDate,
		}).Error("Failed to create subscription with transaction")
		return nil, err
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

// UpdateSubscriptionWithTransaction updates an existing subscription with transaction-based validation
func (s *SubscriptionService) UpdateSubscriptionWithTransaction(id uint, updates map[string]interface{}) (*models.Subscription, error) {
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

// DeleteSubscriptionWithValidation deletes a subscription with validation
func (s *SubscriptionService) DeleteSubscriptionWithValidation(id uint) error {
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

// CalculateTotalCost calculates total cost - main business requirement
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

	// Get subscriptions in date range
	subscriptions, err := s.repo.GetSubscriptionsInDateRange(req.UserID, req.ServiceName, req.StartDate, req.EndDate)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get subscriptions in date range")
		return nil, err
	}

	// Calculate total cost
	totalCost := 0
	for _, sub := range subscriptions {
		// TODO: Implement proper monthly calculation based on date ranges
		totalCost += sub.Price
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
		"subscription_count": len(subscriptions),
	}).Info("Total cost calculated")

	return response, nil
}

// Simple date validation helper
func isValidDate(date string) bool {
	matched, _ := regexp.MatchString(`^(0[1-9]|1[0-2])-[0-9]{4}$`, date)
	return matched
}
