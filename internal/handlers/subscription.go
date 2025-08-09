package handlers

import (
	"net/http"
	"strconv"
	"subscription_tracker_api/internal/models"
	"subscription_tracker_api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type SubscriptionHandler struct {
	service *service.SubscriptionService
	logger  *logrus.Logger
}

func NewSubscriptionHandler(service *service.SubscriptionService, logger *logrus.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		service: service,
		logger:  logger,
	}
}

// CreateSubscription creates a new subscription
func (h *SubscriptionHandler) CreateSubscription(c *gin.Context) {
	h.logger.Info("Received request to create subscription")

	var req models.CreateSubscriptionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to bind JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":      req.UserID,
		"service_name": req.ServiceName,
		"price":        req.Price,
		"start_date":   req.StartDate,
	}).Info("Creating subscription with validated input")

	subscription, err := h.service.CreateSubscriptionWithTransaction(&req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create subscription")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"subscription_id": subscription.ID,
		"user_id":         subscription.UserID,
	}).Info("Subscription creation request completed successfully")

	c.JSON(http.StatusCreated, subscription)
}

// GetSubscription retrieves a subscription by ID
func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	idStr := c.Param("id")

	h.logger.WithField("subscription_id", idStr).Info("Received request to get subscription")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.WithError(err).WithField("subscription_id", idStr).Error("Invalid subscription ID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	subscription, err := h.service.GetSubscriptionByID(uint(id))
	if err != nil {
		h.logger.WithError(err).WithField("subscription_id", id).Error("Failed to get subscription")
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"subscription_id": id,
		"service_name":    subscription.ServiceName,
		"user_id":         subscription.UserID,
	}).Info("Subscription retrieved successfully")

	c.JSON(http.StatusOK, subscription)
}

// UpdateSubscription updates an existing subscription
func (h *SubscriptionHandler) UpdateSubscription(c *gin.Context) {
	idStr := c.Param("id")

	h.logger.WithField("subscription_id", idStr).Info("Received request to update subscription")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.WithError(err).WithField("subscription_id", idStr).Error("Invalid subscription ID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.WithError(err).WithField("subscription_id", id).Error("Failed to bind JSON for update")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"subscription_id": id,
		"updates":         updates,
	}).Info("Processing subscription update with validated input")

	subscription, err := h.service.UpdateSubscriptionWithTransaction(uint(id), updates)
	if err != nil {
		h.logger.WithError(err).WithField("subscription_id", id).Error("Failed to update subscription")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"subscription_id": id,
		"service_name":    subscription.ServiceName,
	}).Info("Subscription update request completed successfully")

	c.JSON(http.StatusOK, subscription)
}

// DeleteSubscription deletes a subscription
func (h *SubscriptionHandler) DeleteSubscription(c *gin.Context) {
	idStr := c.Param("id")

	h.logger.WithField("subscription_id", idStr).Info("Received request to delete subscription")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.WithError(err).WithField("subscription_id", idStr).Error("Invalid subscription ID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	err = h.service.DeleteSubscriptionWithValidation(uint(id))
	if err != nil {
		h.logger.WithError(err).WithField("subscription_id", id).Error("Failed to delete subscription")
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	h.logger.WithField("subscription_id", id).Info("Subscription deletion request completed successfully")

	c.Status(http.StatusNoContent)
}

// ListSubscriptions retrieves all subscriptions with optional filtering
func (h *SubscriptionHandler) ListSubscriptions(c *gin.Context) {
	h.logger.Info("Received request to list subscriptions")

	// Parse query parameters
	var userID *uuid.UUID
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		parsedUUID, err := uuid.Parse(userIDStr)
		if err != nil {
			h.logger.WithError(err).WithField("user_id", userIDStr).Error("Invalid user_id format")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id format"})
			return
		}
		userID = &parsedUUID
	}

	var serviceName *string
	if serviceNameStr := c.Query("service_name"); serviceNameStr != "" {
		serviceName = &serviceNameStr
	}

	limit := 50 // default
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	offset := 0 // default
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil {
			offset = parsedOffset
		}
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":      userID,
		"service_name": serviceName,
		"limit":        limit,
		"offset":       offset,
	}).Info("Processing list subscriptions request with filters")

	subscriptions, err := h.service.ListSubscriptions(userID, serviceName, limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list subscriptions")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve subscriptions"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"subscription_count": len(subscriptions),
		"user_id":            userID,
		"service_name":       serviceName,
	}).Info("Successfully retrieved subscriptions list")

	c.JSON(http.StatusOK, subscriptions)
}

// CalculateTotalCost calculates total cost of subscriptions for a period
func (h *SubscriptionHandler) CalculateTotalCost(c *gin.Context) {
	h.logger.Info("Received request to calculate total cost")

	// Build request from query parameters
	req := &models.CostCalculationRequest{
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
	}

	// Validate required parameters
	if req.StartDate == "" || req.EndDate == "" {
		h.logger.Error("Missing required parameters for cost calculation")
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	// Parse optional user_id
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		parsedUUID, err := uuid.Parse(userIDStr)
		if err != nil {
			h.logger.WithError(err).WithField("user_id", userIDStr).Error("Invalid user_id format")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id format"})
			return
		}
		req.UserID = &parsedUUID
	}

	// Parse optional service_name
	if serviceNameStr := c.Query("service_name"); serviceNameStr != "" {
		req.ServiceName = &serviceNameStr
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":      req.UserID,
		"service_name": req.ServiceName,
		"start_date":   req.StartDate,
		"end_date":     req.EndDate,
	}).Info("Processing cost calculation request with validated parameters")

	response, err := h.service.CalculateTotalCost(req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to calculate total cost")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"total_cost":         response.TotalCost,
		"subscription_count": len(response.Subscriptions),
		"date_range":         req.StartDate + " to " + req.EndDate,
	}).Info("Cost calculation completed successfully")

	c.JSON(http.StatusOK, response)
}
