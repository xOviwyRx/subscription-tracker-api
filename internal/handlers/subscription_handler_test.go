package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"subscription_tracker_api/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Simple mock service for testing
type MockSubscriptionService struct {
	mock.Mock
}

func (m *MockSubscriptionService) CreateSubscription(req *models.CreateSubscriptionRequest) (*models.Subscription, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (m *MockSubscriptionService) GetSubscriptionByID(id uint) (*models.Subscription, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (m *MockSubscriptionService) DeleteSubscription(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

// Add other interface methods as needed (can be empty for now)
func (m *MockSubscriptionService) UpdateSubscription(id uint, updates map[string]interface{}) (*models.Subscription, error) {
	return nil, nil
}
func (m *MockSubscriptionService) ListSubscriptions(userID *uuid.UUID, serviceName *string, limit, offset int) ([]models.Subscription, error) {
	return nil, nil
}
func (m *MockSubscriptionService) CalculateTotalCost(req *models.CostCalculationRequest) (*models.CostCalculationResponse, error) {
	return nil, nil
}

func setupTestHandler() (*SubscriptionHandler, *MockSubscriptionService) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel) // Suppress logs during testing

	mockService := &MockSubscriptionService{}
	handler := NewSubscriptionHandler(mockService, logger)

	return handler, mockService
}

func TestCreateSubscription_Success(t *testing.T) {
	handler, mockService := setupTestHandler()

	userID := uuid.New()
	requestBody := models.CreateSubscriptionRequest{
		ServiceName: "Netflix",
		Price:       999,
		UserID:      userID,
		StartDate:   "01-2024",
	}

	expectedSubscription := &models.Subscription{
		ID:          1,
		ServiceName: "Netflix",
		Price:       999,
		UserID:      userID,
		StartDate:   "01-2024",
	}

	// Fix: Change from CreateSubscriptionWithTransaction to CreateSubscription
	mockService.On("CreateSubscription", mock.AnythingOfType("*models.CreateSubscriptionRequest")).Return(expectedSubscription, nil)

	// Create request
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/subscriptions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder and context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Call handler
	handler.CreateSubscription(c)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Subscription
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedSubscription.ServiceName, response.ServiceName)
	assert.Equal(t, expectedSubscription.Price, response.Price)

	mockService.AssertExpectations(t)
}

func TestCreateSubscription_InvalidJSON(t *testing.T) {
	handler, _ := setupTestHandler()

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/subscriptions", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Call handler
	handler.CreateSubscription(c)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Invalid JSON format")
}

func TestGetSubscription_Success(t *testing.T) {
	handler, mockService := setupTestHandler()

	expectedSubscription := &models.Subscription{
		ID:          1,
		ServiceName: "Netflix",
		Price:       999,
		UserID:      uuid.New(),
		StartDate:   "01-2024",
	}

	// Setup mock
	mockService.On("GetSubscriptionByID", uint(1)).Return(expectedSubscription, nil)

	// Create request
	req := httptest.NewRequest("GET", "/subscriptions/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	// Call handler
	handler.GetSubscription(c)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Subscription
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedSubscription.ServiceName, response.ServiceName)

	mockService.AssertExpectations(t)
}

func TestGetSubscription_InvalidID(t *testing.T) {
	handler, _ := setupTestHandler()

	// Create request with invalid ID
	req := httptest.NewRequest("GET", "/subscriptions/invalid", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "id", Value: "invalid"}}

	// Call handler
	handler.GetSubscription(c)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Invalid subscription ID")
}

func TestGetSubscription_NotFound(t *testing.T) {
	handler, mockService := setupTestHandler()

	// Setup mock to return error
	mockService.On("GetSubscriptionByID", uint(999)).Return(nil, errors.New("subscription not found"))

	// Create request
	req := httptest.NewRequest("GET", "/subscriptions/999", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "id", Value: "999"}}

	// Call handler
	handler.GetSubscription(c)

	// Assertions
	assert.Equal(t, http.StatusNotFound, w.Code)

	mockService.AssertExpectations(t)
}

func TestDeleteSubscription_Success(t *testing.T) {
	handler, mockService := setupTestHandler()
	mockService.On("DeleteSubscription", uint(1)).Return(nil)

	// Use a full Gin router instead of just context
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/subscriptions/:id", handler.DeleteSubscription)

	req := httptest.NewRequest("DELETE", "/subscriptions/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetStatusCodeForError(t *testing.T) {
	handler := &SubscriptionHandler{logger: logrus.New()}

	tests := []struct {
		name           string
		error          error
		expectedStatus int
	}{
		{
			name:           "validation error",
			error:          errors.New("price must be greater than 0"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "not found error",
			error:          errors.New("subscription not found"),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "conflict error",
			error:          errors.New("subscription already exists"),
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "unknown error",
			error:          errors.New("database connection failed"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode := handler.getStatusCodeForError(tt.error)
			assert.Equal(t, tt.expectedStatus, statusCode)
		})
	}
}
