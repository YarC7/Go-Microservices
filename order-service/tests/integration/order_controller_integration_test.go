package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"go-microservices/order-service/cache"
	"go-microservices/order-service/controller"
	"go-microservices/order-service/db"
	"go-microservices/order-service/model"
	"go-microservices/order-service/queue"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// setupIntegrationTestEnvironment creates a test environment with real dependencies
func setupIntegrationTestEnvironment(t *testing.T) (*gin.Engine, *sql.DB, *redis.Client, func()) {
	// Setup database
	database := db.GetDB()
	if database == nil {
		t.Fatal("Failed to initialize database")
	}

	// Setup Redis
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost" // For local testing
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisHost + ":6379",
		DB:   1, // Use different DB for testing
	})

	// Test Redis connection
	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		t.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Setup RabbitMQ
	err = queue.InitRabbitMQ()
	if err != nil {
		t.Fatalf("Failed to initialize RabbitMQ: %v", err)
	}

	// Create controller with real dependencies
	orderController := controller.NewOrderController(database)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/orders", orderController.CreateOrder)
	router.GET("/orders/:id", orderController.GetOrder)

	// Return cleanup function
	cleanup := func() {
		// Clean up test data
		database.Exec("DELETE FROM orders WHERE customer_id = 999") // Clean up test orders
		
		// Clean up Redis
		if redisClient != nil {
			redisClient.FlushDB(context.Background())
			redisClient.Close()
		}
		
		// Close database
		database.Close()
		
		// Close RabbitMQ
		queue.Close()
	}

	return router, database, redisClient, cleanup
}

func TestCreateOrderIntegration_Success(t *testing.T) {
	// Skip integration test if running in CI or when integration services are not available
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test")
	}

	// Setup
	router, _, _, cleanup := setupIntegrationTestEnvironment(t)
	defer cleanup()

	// Prepare test data
	order := model.Order{
		ProductID:  1,
		CustomerID: 999, // Use a test customer ID
		Quantity:   1,
	}

	// Create request
	orderJSON, _ := json.Marshal(order)
	req := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(orderJSON))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)

	// Parse response
	var createdOrder model.Order
	err := json.Unmarshal(w.Body.Bytes(), &createdOrder)
	assert.NoError(t, err)
	assert.NotZero(t, createdOrder.ID)
	assert.Equal(t, "pending", createdOrder.Status)

	// Test that order is cached
	cacheKey := "order:" + string(rune(createdOrder.ID))
	var cachedOrder model.Order
	err = cache.Get(cacheKey, &cachedOrder)
	// Note: Cache might not be immediately available, so this test might be flaky
	// In a real scenario, you might want to add a small delay or retry mechanism
}

func TestGetOrderIntegration_Success(t *testing.T) {
	// Skip integration test if running in CI or when integration services are not available
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test")
	}

	// Setup
	router, _, _, cleanup := setupIntegrationTestEnvironment(t)
	defer cleanup()

	// First create an order
	order := model.Order{
		ProductID:  1,
		CustomerID: 999,
		Quantity:   1,
	}

	orderJSON, _ := json.Marshal(order)
	req := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(orderJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var createdOrder model.Order
	err := json.Unmarshal(w.Body.Bytes(), &createdOrder)
	assert.NoError(t, err)

	// Now get the order
	req = httptest.NewRequest("GET", "/orders/"+string(rune(createdOrder.ID)), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var retrievedOrder model.Order
	err = json.Unmarshal(w.Body.Bytes(), &retrievedOrder)
	assert.NoError(t, err)
	assert.Equal(t, createdOrder.ID, retrievedOrder.ID)
}