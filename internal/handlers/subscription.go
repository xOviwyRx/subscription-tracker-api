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
// @Summary Create a new subscription
// @Description Create a new subscription record
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body models.CreateSubscriptionRequest true "Subscription data"
// @Success 201 {object} models.Subscription
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions [post]
func (h *SubscriptionHandler) CreateSubscription(c *gin.Context) {
	var req models.CreateSubscriptionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to bind JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	subscription, err := h.service.CreateSubscription(&req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create subscription")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, subscription)
}

// GetSubscription retrieves a subscription by ID
// @Summary Get subscription by ID
// @Description Get a single subscription by its ID
// @Tags subscriptions
// @Produce json
// @Param id path int true "Subscription ID"
// @Success 200 {object} models.Subscription
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /subscriptions/{id} [get]
func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	subscription, err := h.service.GetSubscriptionByID(uint(id))
	if err != nil {
		h.logger.WithError(err).Error("Failed to get subscription")
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// UpdateSubscription updates an existing subscription
// @Summary Update subscription
// @Description Update an existing subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path int true "Subscription ID"
// @Param updates body map[string]interface{} true "Update data"
// @Success 200 {object} models.Subscription
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /subscriptions/{id} [put]
func (h *SubscriptionHandler) UpdateSubscription(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	subscription, err := h.service.UpdateSubscription(uint(id), updates)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update subscription")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// DeleteSubscription deletes a subscription
// @Summary Delete subscription
// @Description Delete a subscription by ID
// @Tags subscriptions
// @Param id path int true "Subscription ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /subscriptions/{id} [delete]
func (h *SubscriptionHandler) DeleteSubscription(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	err = h.service.DeleteSubscription(uint(id))
	if err != nil {
		h.logger.WithError(err).Error("Failed to delete subscription")
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListSubscriptions retrieves all subscriptions with optional filtering
// @Summary List subscriptions
// @Description Get list of subscriptions with optional filtering
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "Filter by user ID (UUID)"
// @Param service_name query string false "Filter by service name"
// @Param limit query int false "Limit number of results (default: 50, max: 100)"
// @Param offset query int false "Offset for pagination (default: 0)"
// @Success 200 {array} models.Subscription
// @Failure 400 {object} map[string]string
// @Router /subscriptions [get]
func (h *SubscriptionHandler) ListSubscriptions(c *gin.Context) {
	// Parse query parameters
	var userID *uuid.UUID
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		parsedUUID, err := uuid.Parse(userIDStr)
		if err != nil {
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

	subscriptions, err := h.service.ListSubscriptions(userID, serviceName, limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list subscriptions")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve subscriptions"})
		return
	}

	c.JSON(http.StatusOK, subscriptions)
}

// CalculateTotalCost calculates total cost of subscriptions for a period
// @Summary Calculate total subscription cost
// @Description Calculate the total cost of subscriptions within a date range with optional filtering
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "Filter by user ID (UUID)"
// @Param service_name query string false "Filter by service name"
// @Param start_date query string true "Start date (MM-YYYY format)"
// @Param end_date query string true "End date (MM-YYYY format)"
// @Success 200 {object} models.CostCalculationResponse
// @Failure 400 {object} map[string]string
// @Router /subscriptions/cost [get]
func (h *SubscriptionHandler) CalculateTotalCost(c *gin.Context) {
	// Build request from query parameters
	req := &models.CostCalculationRequest{
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
	}

	// Validate required parameters
	if req.StartDate == "" || req.EndDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	// Parse optional user_id
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		parsedUUID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id format"})
			return
		}
		req.UserID = &parsedUUID
	}

	// Parse optional service_name
	if serviceNameStr := c.Query("service_name"); serviceNameStr != "" {
		req.ServiceName = &serviceNameStr
	}

	response, err := h.service.CalculateTotalCost(req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to calculate total cost")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
