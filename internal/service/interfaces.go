package service

import (
	"github.com/google/uuid"
	"subscription_tracker_api/internal/models"
)

// SubscriptionServiceInterface defines what the handlers need from the service
type SubscriptionServiceInterface interface {
	CreateSubscriptionWithTransaction(req *models.CreateSubscriptionRequest) (*models.Subscription, error)
	GetSubscriptionByID(id uint) (*models.Subscription, error)
	UpdateSubscriptionWithTransaction(id uint, updates map[string]interface{}) (*models.Subscription, error)
	DeleteSubscriptionWithValidation(id uint) error
	ListSubscriptions(userID *uuid.UUID, serviceName *string, limit, offset int) ([]models.Subscription, error)
	CalculateTotalCost(req *models.CostCalculationRequest) (*models.CostCalculationResponse, error)
}
