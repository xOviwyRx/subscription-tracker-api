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

// Create subscription
func (s *SubscriptionService) CreateSubscription(req *models.CreateSubscriptionRequest) (*models.Subscription, error) {
	// Basic validation
	if req.ServiceName == "" || req.Price <= 0 || req.UserID == uuid.Nil {
		return nil, errors.New("invalid input data")
	}

	if !isValidDate(req.StartDate) {
		return nil, errors.New("start_date must be in MM-YYYY format")
	}

	subscription := &models.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}

	return subscription, s.repo.Create(subscription)
}

// Get subscription by ID
func (s *SubscriptionService) GetSubscriptionByID(id uint) (*models.Subscription, error) {
	return s.repo.GetByID(id)
}

// Update subscription
func (s *SubscriptionService) UpdateSubscription(id uint, updates map[string]interface{}) (*models.Subscription, error) {
	subscription, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Apply simple updates
	if serviceName, ok := updates["service_name"].(string); ok && serviceName != "" {
		subscription.ServiceName = serviceName
	}
	if price, ok := updates["price"].(float64); ok && price > 0 {
		subscription.Price = int(price)
	}

	return subscription, s.repo.Update(subscription)
}

// Delete subscription
func (s *SubscriptionService) DeleteSubscription(id uint) error {
	return s.repo.Delete(id)
}

// List subscriptions
func (s *SubscriptionService) ListSubscriptions(userID *uuid.UUID, serviceName *string, limit, offset int) ([]models.Subscription, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.List(userID, serviceName, limit, offset)
}

// Calculate total cost - main business requirement
func (s *SubscriptionService) CalculateTotalCost(req *models.CostCalculationRequest) (*models.CostCalculationResponse, error) {
	// Get subscriptions in date range
	subscriptions, err := s.repo.GetSubscriptionsInDateRange(req.UserID, req.ServiceName, req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	// Simple cost calculation
	totalCost := 0
	for _, sub := range subscriptions {
		totalCost += sub.Price // Simplified: assume 1 month cost for now
	}

	return &models.CostCalculationResponse{
		TotalCost:     totalCost,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		UserID:        req.UserID,
		ServiceName:   req.ServiceName,
		Subscriptions: subscriptions,
	}, nil
}

// Simple date validation helper
func isValidDate(date string) bool {
	matched, _ := regexp.MatchString(`^(0[1-9]|1[0-2])-[0-9]{4}`, date)
	return matched
}
