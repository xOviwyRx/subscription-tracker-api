package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Subscription represents a user's subscription to a service
type Subscription struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	ServiceName string         `json:"service_name" gorm:"not null" validate:"required"`
	Price       int            `json:"price" gorm:"not null" validate:"required,min=1"`
	UserID      uuid.UUID      `json:"user_id" gorm:"type:uuid;not null" validate:"required"`
	StartDate   string         `json:"start_date" gorm:"not null" validate:"required"` // Format: MM-YYYY
	EndDate     *string        `json:"end_date,omitempty"`                             // Optional, Format: MM-YYYY
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// CreateSubscriptionRequest represents the request payload for creating a subscription
type CreateSubscriptionRequest struct {
	ServiceName string    `json:"service_name" validate:"required"`
	Price       int       `json:"price" validate:"required,min=1"`
	UserID      uuid.UUID `json:"user_id" validate:"required"`
	StartDate   string    `json:"start_date" validate:"required"` // Format: MM-YYYY
	EndDate     *string   `json:"end_date,omitempty"`             // Optional, Format: MM-YYYY
}

// CostCalculationRequest represents the request for calculating total cost
type CostCalculationRequest struct {
	UserID      *uuid.UUID `form:"user_id"`
	ServiceName *string    `form:"service_name"`
	StartDate   string     `form:"start_date" validate:"required"` // Format: MM-YYYY
	EndDate     string     `form:"end_date" validate:"required"`   // Format: MM-YYYY
}

// CostCalculationResponse represents the response for cost calculation
type CostCalculationResponse struct {
	TotalCost     int            `json:"total_cost"`
	StartDate     string         `json:"start_date"`
	EndDate       string         `json:"end_date"`
	UserID        *uuid.UUID     `json:"user_id,omitempty"`
	ServiceName   *string        `json:"service_name,omitempty"`
	Subscriptions []Subscription `json:"subscriptions"`
}
