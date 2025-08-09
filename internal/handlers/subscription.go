package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"subscription_tracker_api/internal/models"
	"subscription_tracker_api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type SubscriptionHandler struct {
	service service.SubscriptionServiceInterface
	logger  *logrus.Logger
}

func NewSubscriptionHandler(service service.SubscriptionServiceInterface, logger *logrus.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		service: service,
		logger:  logger,
	}
}

// CreateSubscription creates a new subscription
// @Summary Create a new subscription
// @Description Create a new subscription for a user
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body models.CreateSubscriptionRequest true "Subscription data"
// @Success 201 {object} models.Subscription "Subscription created successfully"
// @Failure 400 {object} models.ErrorResponse "Bad Request - Invalid input data or validation errors"
// @Failure 409 {object} models.ErrorResponse "Conflict - Duplicate subscription or constraint violation"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error - Database or server errors"
// @Router /subscriptions [post]
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

		// Determine appropriate status code based on error type
		statusCode := h.getStatusCodeForError(err)

		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"subscription_id": subscription.ID,
		"user_id":         subscription.UserID,
	}).Info("Subscription creation request completed successfully")

	c.JSON(http.StatusCreated, subscription)
}

// GetSubscription retrieves a subscription by ID
// @Summary Get subscription by ID
// @Description Retrieve a single subscription by its ID
// @Tags subscriptions
// @Produce json
// @Param id path int true "Subscription ID"
// @Success 200 {object} models.Subscription "Subscription retrieved successfully"
// @Failure 400 {object} models.ErrorResponse "Bad Request - Invalid subscription ID format"
// @Failure 404 {object} models.ErrorResponse "Not Found - Subscription not found"
// @Router /subscriptions/{id} [get]
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
// @Summary Update an existing subscription
// @Description Update subscription details by ID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path int true "Subscription ID"
// @Param updates body map[string]interface{} true "Fields to update"
// @Success 200 {object} models.Subscription "Subscription updated successfully"
// @Failure 400 {object} models.ErrorResponse "Bad Request - Invalid input data or validation errors"
// @Failure 404 {object} models.ErrorResponse "Not Found - Subscription does not exist"
// @Failure 409 {object} models.ErrorResponse "Conflict - Constraint violation during update"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error - Database or server errors"
// @Router /subscriptions/{id} [put]
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

		// Determine appropriate status code based on error type
		statusCode := h.getStatusCodeForError(err)

		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"subscription_id": id,
		"service_name":    subscription.ServiceName,
	}).Info("Subscription update request completed successfully")

	c.JSON(http.StatusOK, subscription)
}

// DeleteSubscription deletes a subscription
// @Summary Delete subscription by ID
// @Description Delete a subscription by its ID
// @Tags subscriptions
// @Param id path int true "Subscription ID"
// @Success 204 "Subscription deleted successfully"
// @Failure 400 {object} models.ErrorResponse "Bad Request - Invalid subscription ID format"
// @Failure 404 {object} models.ErrorResponse "Not Found - Subscription not found"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error - Database or server errors"
// @Router /subscriptions/{id} [delete]
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
// @Summary List subscriptions
// @Description Retrieve subscriptions with optional filtering
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "Filter by user ID (UUID)"
// @Param service_name query string false "Filter by service name"
// @Param limit query int false "Number of results to return (default: 50)"
// @Param offset query int false "Number of results to skip (default: 0)"
// @Success 200 {array} models.Subscription "Subscriptions retrieved successfully"
// @Failure 400 {object} models.ErrorResponse "Bad Request - Invalid query parameters"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error - Failed to retrieve subscriptions"
// @Router /subscriptions [get]
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
// @Summary Calculate total cost of subscriptions
// @Description Calculate the total cost of subscriptions within a date range
// @Tags subscriptions
// @Produce json
// @Param start_date query string true "Start date in MM-YYYY format"
// @Param end_date query string true "End date in MM-YYYY format"
// @Param user_id query string false "Filter by user ID (UUID)"
// @Param service_name query string false "Filter by service name"
// @Success 200 {object} models.CostCalculationResponse "Cost calculation completed successfully"
// @Failure 400 {object} models.ErrorResponse "Bad Request - Invalid date format or missing required parameters"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error - Database query failed or server errors"
// @Router /subscriptions/calculate-cost [get]
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

		// Determine appropriate status code based on error type
		statusCode := h.getStatusCodeForError(err)

		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"total_cost":         response.TotalCost,
		"subscription_count": len(response.Subscriptions),
		"date_range":         req.StartDate + " to " + req.EndDate,
	}).Info("Cost calculation completed successfully")

	c.JSON(http.StatusOK, response)
}

// Helper method to determine appropriate HTTP status code based on error type
func (h *SubscriptionHandler) getStatusCodeForError(err error) int {
	errorMsg := err.Error()

	// Validation errors - 400 Bad Request
	validationErrors := []string{
		"invalid input data",
		"must be in MM-YYYY format",
		"must be after start_date",
		"must be before end_date",
		"are required",
		"must be greater than 0",
		"price must be greater than 0",
		"service_name, price, and user_id are required",
	}

	for _, validationErr := range validationErrors {
		if strings.Contains(errorMsg, validationErr) {
			return http.StatusBadRequest // 400
		}
	}

	// Not found errors - 404 Not Found
	notFoundErrors := []string{
		"not found",
		"does not exist",
		"record not found",
	}

	for _, notFoundErr := range notFoundErrors {
		if strings.Contains(errorMsg, notFoundErr) {
			return http.StatusNotFound // 404
		}
	}

	// Conflict errors - 409 Conflict (e.g., duplicate subscription)
	conflictErrors := []string{
		"already exists",
		"duplicate",
		"conflict",
		"constraint violation",
		"unique constraint",
		"UNIQUE constraint failed",
	}

	for _, conflictErr := range conflictErrors {
		if strings.Contains(errorMsg, conflictErr) {
			return http.StatusConflict // 409
		}
	}

	// Database/internal errors - 500 Internal Server Error
	return http.StatusInternalServerError // 500
}
