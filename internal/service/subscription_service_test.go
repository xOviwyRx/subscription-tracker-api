package service

import (
	"errors"
	"testing"

	"subscription_tracker_api/internal/models"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSubscriptionRepository for testing service layer
type MockSubscriptionRepository struct {
	mock.Mock
}

func (m *MockSubscriptionRepository) CreateWithTransaction(subscription *models.Subscription) error {
	args := m.Called(subscription)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) GetByID(id uint) (*models.Subscription, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) UpdateWithTransaction(subscription *models.Subscription) error {
	args := m.Called(subscription)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) DeleteWithValidation(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) List(userID *uuid.UUID, serviceName *string, limit, offset int) ([]models.Subscription, error) {
	args := m.Called(userID, serviceName, limit, offset)
	return args.Get(0).([]models.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) GetSubscriptionsInDateRange(userID *uuid.UUID, serviceName *string, startDate, endDate string) ([]models.Subscription, error) {
	args := m.Called(userID, serviceName, startDate, endDate)
	return args.Get(0).([]models.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) CalculateTotalCostInDB(userID *uuid.UUID, serviceName *string, startDate, endDate string, totalMonths int) (int, error) {
	args := m.Called(userID, serviceName, startDate, endDate, totalMonths)
	return args.Get(0).(int), args.Error(1)
}

func setupTestService() (*SubscriptionService, *MockSubscriptionRepository) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel) // Suppress logs during testing

	mockRepo := &MockSubscriptionRepository{}
	service := NewSubscriptionService(mockRepo, logger)

	return service, mockRepo
}

func TestCreateSubscriptionWithTransaction_Success(t *testing.T) {
	service, mockRepo := setupTestService()

	userID := uuid.New()
	req := &models.CreateSubscriptionRequest{
		ServiceName: "Netflix",
		Price:       999,
		UserID:      userID,
		StartDate:   "01-2024",
	}

	// Setup mock to expect the subscription creation
	mockRepo.On("CreateWithTransaction", mock.MatchedBy(func(sub *models.Subscription) bool {
		return sub.ServiceName == "Netflix" && sub.Price == 999 && sub.UserID == userID
	})).Return(nil).Run(func(args mock.Arguments) {
		// Simulate database setting ID
		sub := args.Get(0).(*models.Subscription)
		sub.ID = 1
	})

	// Call service
	result, err := service.CreateSubscription(req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint(1), result.ID)
	assert.Equal(t, "Netflix", result.ServiceName)
	assert.Equal(t, 999, result.Price)
	assert.Equal(t, userID, result.UserID)

	mockRepo.AssertExpectations(t)
}

func TestCreateSubscriptionWithTransaction_ValidationErrors(t *testing.T) {
	service, _ := setupTestService()

	testCases := []struct {
		name        string
		req         *models.CreateSubscriptionRequest
		expectedErr string
	}{
		{
			name: "empty service name",
			req: &models.CreateSubscriptionRequest{
				ServiceName: "",
				Price:       999,
				UserID:      uuid.New(),
				StartDate:   "01-2024",
			},
			expectedErr: "invalid input data",
		},
		{
			name: "zero price",
			req: &models.CreateSubscriptionRequest{
				ServiceName: "Netflix",
				Price:       0,
				UserID:      uuid.New(),
				StartDate:   "01-2024",
			},
			expectedErr: "invalid input data",
		},
		{
			name: "nil user ID",
			req: &models.CreateSubscriptionRequest{
				ServiceName: "Netflix",
				Price:       999,
				UserID:      uuid.Nil,
				StartDate:   "01-2024",
			},
			expectedErr: "invalid input data",
		},
		{
			name: "invalid start date format",
			req: &models.CreateSubscriptionRequest{
				ServiceName: "Netflix",
				Price:       999,
				UserID:      uuid.New(),
				StartDate:   "2024-01",
			},
			expectedErr: "start_date must be in MM-YYYY format",
		},
		{
			name: "invalid end date format",
			req: &models.CreateSubscriptionRequest{
				ServiceName: "Netflix",
				Price:       999,
				UserID:      uuid.New(),
				StartDate:   "01-2024",
				EndDate:     stringPtr("2024-02"),
			},
			expectedErr: "end_date must be in MM-YYYY format",
		},
		{
			name: "end date before start date",
			req: &models.CreateSubscriptionRequest{
				ServiceName: "Netflix",
				Price:       999,
				UserID:      uuid.New(),
				StartDate:   "02-2024",
				EndDate:     stringPtr("01-2024"),
			},
			expectedErr: "end_date must be after start_date",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := service.CreateSubscription(tc.req)

			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

func TestUpdateSubscriptionWithTransaction_Success(t *testing.T) {
	service, mockRepo := setupTestService()

	userID := uuid.New()
	existingSubscription := &models.Subscription{
		ID:          1,
		ServiceName: "Netflix",
		Price:       999,
		UserID:      userID,
		StartDate:   "01-2024",
	}

	updates := map[string]interface{}{
		"price": float64(1199),
	}

	// Mock getting existing subscription
	mockRepo.On("GetByID", uint(1)).Return(existingSubscription, nil)

	// Mock successful update
	mockRepo.On("UpdateWithTransaction", mock.MatchedBy(func(sub *models.Subscription) bool {
		return sub.ID == 1 && sub.Price == 1199
	})).Return(nil)

	// Call service
	result, err := service.UpdateSubscription(1, updates)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1199, result.Price)

	mockRepo.AssertExpectations(t)
}

func TestUpdateSubscriptionWithTransaction_NotFound(t *testing.T) {
	service, mockRepo := setupTestService()

	// Mock subscription not found
	mockRepo.On("GetByID", uint(999)).Return(nil, errors.New("record not found"))

	updates := map[string]interface{}{
		"price": float64(1199),
	}

	// Call service
	result, err := service.UpdateSubscription(999, updates)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "subscription not found")

	mockRepo.AssertExpectations(t)
}

func TestCalculateTotalCost_Success(t *testing.T) {
	service, mockRepo := setupTestService()

	userID := uuid.New()
	serviceName := "Netflix"
	req := &models.CostCalculationRequest{
		UserID:      &userID,
		ServiceName: &serviceName,
		StartDate:   "01-2024",
		EndDate:     "03-2024",
	}

	subscriptions := []models.Subscription{
		{
			ID:          1,
			ServiceName: "Netflix",
			Price:       999,
			UserID:      userID,
			StartDate:   "01-2024",
		},
	}

	// Mock database aggregation
	mockRepo.On("CalculateTotalCostInDB", &userID, &serviceName, "01-2024", "03-2024", 3).Return(2997, nil)

	// Mock getting subscriptions for response
	mockRepo.On("GetSubscriptionsInDateRange", &userID, &serviceName, "01-2024", "03-2024").Return(subscriptions, nil)

	// Call service
	result, err := service.CalculateTotalCost(req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2997, result.TotalCost)
	assert.Equal(t, "01-2024", result.StartDate)
	assert.Equal(t, "03-2024", result.EndDate)
	assert.Equal(t, len(subscriptions), len(result.Subscriptions))

	mockRepo.AssertExpectations(t)
}

func TestCalculateTotalCost_ValidationErrors(t *testing.T) {
	service, _ := setupTestService()

	testCases := []struct {
		name        string
		req         *models.CostCalculationRequest
		expectedErr string
	}{
		{
			name: "invalid start date format",
			req: &models.CostCalculationRequest{
				StartDate: "2024-01",
				EndDate:   "03-2024",
			},
			expectedErr: "start_date must be in MM-YYYY format",
		},
		{
			name: "invalid end date format",
			req: &models.CostCalculationRequest{
				StartDate: "01-2024",
				EndDate:   "2024-03",
			},
			expectedErr: "end_date must be in MM-YYYY format",
		},
		{
			name: "end date before start date",
			req: &models.CostCalculationRequest{
				StartDate: "03-2024",
				EndDate:   "01-2024",
			},
			expectedErr: "end_date must be after start_date",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := service.CalculateTotalCost(tc.req)

			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

func TestIsValidDate(t *testing.T) {
	testCases := []struct {
		name     string
		date     string
		expected bool
	}{
		{"valid date", "01-2024", true},
		{"valid date dec", "12-2024", true},
		{"invalid format year-month", "2024-01", false},
		{"invalid month", "13-2024", false},
		{"invalid month zero", "00-2024", false},
		{"missing dash", "012024", false},
		{"extra characters", "01-2024-01", false},
		{"empty string", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isValidDate(tc.date)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCalculateMonthsBetween(t *testing.T) {
	testCases := []struct {
		name      string
		startDate string
		endDate   string
		expected  int
	}{
		{"same month", "01-2024", "01-2024", 1},
		{"consecutive months", "01-2024", "02-2024", 2},
		{"quarter", "01-2024", "03-2024", 3},
		{"year span", "12-2023", "01-2024", 2},
		{"full year", "01-2024", "12-2024", 12},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calculateMonthsBetween(tc.startDate, tc.endDate)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
