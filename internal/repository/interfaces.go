package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"subscription_tracker_api/internal/models"
)

// SubscriptionRepositoryInterface defines the contract for subscription data operations
type SubscriptionRepositoryInterface interface {
	Create(tx *gorm.DB, subscription *models.Subscription) error
	GetByID(id uint) (*models.Subscription, error)
	UpdateWithTransaction(subscription *models.Subscription) error
	DeleteWithValidation(id uint) error
	List(userID *uuid.UUID, serviceName *string, limit, offset int) ([]models.Subscription, error)
	GetSubscriptionsInDateRange(userID *uuid.UUID, serviceName *string, startDate, endDate string) ([]models.Subscription, error)
	CalculateTotalCostInDB(userID *uuid.UUID, serviceName *string, startDate, endDate string, totalMonths int) (int, error)
	ExistsByUserServiceAndDate(tx *gorm.DB, userID uuid.UUID, serviceName, startDate string) (bool, error)
}
