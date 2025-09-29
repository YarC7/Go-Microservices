package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-microservices/order-service/controller"
	"go-microservices/order-service/model"
	"go-microservices/order-service/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Define interfaces that match the service methods
type InventoryServiceInterface interface {
	CheckAvailability(productID int, quantity int) (bool, error)
}

type NotificationServiceInterface interface {
	SendOrderNotification(orderID int) error
}

// Mock InventoryService interface
type MockInventoryService struct {
	mock.Mock
}

func (m *MockInventoryService) CheckAvailability(productID int, quantity int) (bool, error) {
	args := m.Called(productID, quantity)
	return args.Bool(0), args.Error(1)
}

// Mock NotificationService interface
type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) SendOrderNotification(orderID int) error {
	args := m.Called(orderID)
	return args.Error(0)
}

func (m *MockNotificationService) SendOrderStatusUpdate(orderID int, customerID int, status string) error {
	args := m.Called(orderID, customerID, status)
	return args.Error(0)
}

func TestCreateOrder(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockInventory := new(MockInventoryService)
	mockNotification := new(MockNotificationService)

	tests := []struct {
		name          string
		order         model.Order
		setupMocks    func()
		expectedCode  int
		expectedError string
	}{
		{
			name: "Success",
			order: model.Order{
				ProductID:  1,
				CustomerID: 1,
				Quantity:   2,
			},
			setupMocks: func() {
				mockInventory.On("CheckAvailability", 1, 2).Return(true, nil)
				mockNotification.On("SendOrderNotification", mock.Anything).Return(nil)
			},
			expectedCode: http.StatusCreated,
		},
		{
			name: "Product Not Available",
			order: model.Order{
				ProductID:  1,
				CustomerID: 1,
				Quantity:   100,
			},
			setupMocks: func() {
				mockInventory.On("CheckAvailability", 1, 100).Return(false, nil)
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Product not available in requested quantity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			tt.setupMocks()

			// Create router and controller
			router := gin.New()
			controller := &controller.OrderController{
				InventoryService:    mockInventory,
				NotificationService: mockNotification,
			}

			// Setup route
			router.POST("/orders", controller.CreateOrder)

			// Create request
			orderJSON, _ := json.Marshal(tt.order)
			req := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(orderJSON))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response["error"])
			}

			// Verify mocks
			mockInventory.AssertExpectations(t)
			mockNotification.AssertExpectations(t)
		})
	}
}

func TestGetOrder(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		orderID       string
		setupMocks    func()
		expectedCode  int
		expectedOrder *model.Order
	}{
		{
			name:    "Success",
			orderID: "1",
			setupMocks: func() {
				// Setup cache mock if needed
			},
			expectedCode: http.StatusOK,
			expectedOrder: &model.Order{
				ID:         1,
				ProductID:  1,
				CustomerID: 1,
				Quantity:   2,
				Status:     "pending",
				CreatedAt:  time.Now(),
			},
		},
		{
			name:         "Not Found",
			orderID:      "999",
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			// Create router and controller
			router := gin.New()
			controller := &controller.OrderController{}

			// Setup route
			router.GET("/orders/:id", controller.GetOrder)

			// Create request
			req := httptest.NewRequest("GET", "/orders/"+tt.orderID, nil)
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectedOrder != nil {
				var response model.Order
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOrder.ID, response.ID)
				assert.Equal(t, tt.expectedOrder.ProductID, response.ProductID)
				assert.Equal(t, tt.expectedOrder.Status, response.Status)
			}
		})
	}
}

func TestCreateBatchOrders(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockInventory := new(MockInventoryService)
	mockNotification := new(MockNotificationService)

	tests := []struct {
		name          string
		orders        []model.Order
		setupMocks    func()
		expectedCode  int
		expectedStats map[string]interface{}
	}{
		{
			name: "Successful Batch",
			orders: []model.Order{
				{ProductID: 1, CustomerID: 1, Quantity: 1},
				{ProductID: 2, CustomerID: 1, Quantity: 2},
			},
			setupMocks: func() {
				mockInventory.On("CheckAvailability", 1, 1).Return(true, nil)
				mockInventory.On("CheckAvailability", 2, 2).Return(true, nil)
				mockNotification.On("SendOrderNotification", mock.Anything).Return(nil).Times(2)
			},
			expectedCode: http.StatusOK,
			expectedStats: map[string]interface{}{
				"total_orders": 2,
				"successful":   2,
				"failed":       0,
			},
		},
		{
			name: "Partial Success",
			orders: []model.Order{
				{ProductID: 1, CustomerID: 1, Quantity: 1},
				{ProductID: 2, CustomerID: 1, Quantity: 1000}, // Will fail availability check
			},
			setupMocks: func() {
				mockInventory.On("CheckAvailability", 1, 1).Return(true, nil)
				mockInventory.On("CheckAvailability", 2, 1000).Return(false, nil)
				mockNotification.On("SendOrderNotification", mock.Anything).Return(nil).Once()
			},
			expectedCode: http.StatusOK,
			expectedStats: map[string]interface{}{
				"total_orders": 2,
				"successful":   1,
				"failed":       1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			tt.setupMocks()

			// Create router and controller
			router := gin.New()
			controller := &controller.OrderController{
				InventoryService:    mockInventory,
				NotificationService: mockNotification,
			}

			// Setup route
			router.POST("/orders/batch", controller.CreateBatchOrders)

			// Create request
			ordersJSON, _ := json.Marshal(tt.orders)
			req := httptest.NewRequest("POST", "/orders/batch", bytes.NewBuffer(ordersJSON))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedCode, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			for key, value := range tt.expectedStats {
				assert.Equal(t, value, response[key])
			}

			// Verify mocks
			mockInventory.AssertExpectations(t)
			mockNotification.AssertExpectations(t)
		})
	}
}
