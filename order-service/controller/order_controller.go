package controller

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"go-microservices/order-service/cache"
	"go-microservices/order-service/metrics"
	"go-microservices/order-service/model"
	"go-microservices/order-service/queue"
	"go-microservices/order-service/service"
	"go-microservices/order-service/worker"

	"github.com/gin-gonic/gin"
)

// OrderController handles order-related requests
type OrderController struct {
	DB                  *sql.DB
	InventoryService    *service.InventoryService
	NotificationService *service.NotificationService
	PaymentService      *service.PaymentService
}

// NewOrderController creates a new order controller
func NewOrderController(db *sql.DB) *OrderController {
	return &OrderController{
		DB:                  db,
		InventoryService:    service.NewInventoryService(),
		NotificationService: service.NewNotificationService(),
		PaymentService:      service.NewPaymentService(),
	}
}

// CreateOrder handles creation of a new order
func (oc *OrderController) CreateOrder(c *gin.Context) {
	var order model.Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check inventory availability using circuit breaker
	available, err := oc.InventoryService.CheckAvailability(order.ProductID, order.Quantity)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Failed to check inventory: " + err.Error()})
		return
	}
	if !available {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product not available in requested quantity"})
		return
	}

	// Insert order into database
	err = oc.insertOrder(&order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order: " + err.Error()})
		return
	}

	// Publish order created event to RabbitMQ
	orderMsg, _ := json.Marshal(order)
	if err := queue.PublishMessage(queue.Config{
		QueueName:    "orders",
		RoutingKey:   "order.created",
		ExchangeName: "orders",
	}, orderMsg); err != nil {
		log.Printf("Warning: Failed to publish order created event: %v\n", err)
	}

	// Send notification using circuit breaker
	go func() {
		if err := oc.NotificationService.SendOrderNotification(order.ID); err != nil {
			log.Printf("Failed to send notification: %v\n", err)
		}
	}()

	c.JSON(http.StatusCreated, order)
}

// CreateOrderWithPayment handles creation of a new order with payment intent
func (oc *OrderController) CreateOrderWithPayment(c *gin.Context) {
	var orderWithPayment struct {
		model.Order
		Currency string `json:"currency" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&orderWithPayment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check inventory availability using circuit breaker
	available, err := oc.InventoryService.CheckAvailability(orderWithPayment.ProductID, orderWithPayment.Quantity)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Failed to check inventory: " + err.Error()})
		return
	}
	if !available {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product not available in requested quantity"})
		return
	}

	// Insert order into database
	err = oc.insertOrder(&orderWithPayment.Order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order: " + err.Error()})
		return
	}

	// Create payment intent
	paymentResp, err := oc.PaymentService.CreatePayment(
		orderWithPayment.ID,
		orderWithPayment.CustomerID,
		orderWithPayment.TotalPrice,
		orderWithPayment.Currency,
	)
	if err != nil {
		log.Printf("Warning: Failed to create payment intent: %v\n", err)
		// Still return the order but indicate payment failed
		c.JSON(http.StatusCreated, gin.H{
			"order": orderWithPayment.Order,
			"payment_error": "Failed to create payment intent: " + err.Error(),
		})
		return
	}

	// Publish order created event to RabbitMQ
	orderMsg, _ := json.Marshal(orderWithPayment.Order)
	if err := queue.PublishMessage(queue.Config{
		QueueName:    "orders",
		RoutingKey:   "order.created",
		ExchangeName: "orders",
	}, orderMsg); err != nil {
		log.Printf("Warning: Failed to publish order created event: %v\n", err)
	}

	// Send notification using circuit breaker
	go func() {
		if err := oc.NotificationService.SendOrderNotification(orderWithPayment.ID); err != nil {
			log.Printf("Failed to send notification: %v\n", err)
		}
	}()

	c.JSON(http.StatusCreated, gin.H{
		"order": orderWithPayment.Order,
		"payment": paymentResp,
	})
}

// GetOrders returns all orders
func (oc *OrderController) GetOrders(c *gin.Context) {
	rows, err := oc.DB.Query("SELECT id, customer_id, product_id, quantity, total_price, status FROM orders")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.ID, &o.CustomerID, &o.ProductID, &o.Quantity, &o.TotalPrice, &o.Status); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		orders = append(orders, o)
	}

	c.JSON(http.StatusOK, orders)
}

// GetOrder returns a specific order by ID
func (oc *OrderController) GetOrder(c *gin.Context) {
	orderID := c.Param("id")

	// Try to get order from cache first
	var order model.Order
	cacheKey := "order:" + orderID
	err := cache.GetOrSet(cacheKey, &order, 30*time.Minute, func() (interface{}, error) {
		// If not in cache, get from database
		return oc.getOrderFromDB(orderID)
	})

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get order: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, order)
}

// UpdateOrder updates an order
func (oc *OrderController) UpdateOrder(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Get existing order to compare status change
	var existingOrder model.Order
	err = oc.DB.QueryRow("SELECT id, customer_id, product_id, quantity, total_price, status FROM orders WHERE id = $1", id).
		Scan(&existingOrder.ID, &existingOrder.CustomerID, &existingOrder.ProductID, &existingOrder.Quantity, &existingOrder.TotalPrice, &existingOrder.Status)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var updatedOrder model.Order
	if err := c.BindJSON(&updatedOrder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := oc.DB.Exec(
		"UPDATE orders SET customer_id = $1, product_id = $2, quantity = $3, total_price = $4, status = $5 WHERE id = $6",
		updatedOrder.CustomerID, updatedOrder.ProductID, updatedOrder.Quantity, updatedOrder.TotalPrice, updatedOrder.Status, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// If status changed, send notification
	if existingOrder.Status != updatedOrder.Status {
		err = oc.NotificationService.SendOrderStatusUpdate(id, updatedOrder.CustomerID, updatedOrder.Status)
		if err != nil {
			// Log the error but continue (non-blocking)
			fmt.Printf("Failed to send status update notification: %v\n", err)
		}
	}

	updatedOrder.ID = id
	c.JSON(http.StatusOK, updatedOrder)
}

// DeleteOrder deletes an order
func (oc *OrderController) DeleteOrder(c *gin.Context) {
	id := c.Param("id")

	// Get the order first
	var order model.Order
	err := oc.DB.QueryRow("SELECT id, customer_id, product_id, quantity, total_price, status FROM orders WHERE id = $1", id).
		Scan(&order.ID, &order.CustomerID, &order.ProductID, &order.Quantity, &order.TotalPrice, &order.Status)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Delete the order
	result, err := oc.DB.Exec("DELETE FROM orders WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// Send notification that order was deleted/cancelled
	err = oc.NotificationService.SendOrderStatusUpdate(order.ID, order.CustomerID, "cancelled")
	if err != nil {
		// Log the error but continue (non-blocking)
		fmt.Printf("Failed to send cancellation notification: %v\n", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order deleted successfully"})
}

// UpdateOrderStatus updates only the status of an order
func (oc *OrderController) UpdateOrderStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var statusUpdate struct {
		Status string `json:"status"`
	}
	if err := c.BindJSON(&statusUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing order to get customer ID
	var order model.Order
	err = oc.DB.QueryRow("SELECT id, customer_id, product_id, quantity, total_price, status FROM orders WHERE id = $1", id).
		Scan(&order.ID, &order.CustomerID, &order.ProductID, &order.Quantity, &order.TotalPrice, &order.Status)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update order status
	result, err := oc.DB.Exec("UPDATE orders SET status = $1 WHERE id = $2", statusUpdate.Status, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	metrics.OrderStatusUpdated.WithLabelValues(statusUpdate.Status).Inc()

	// Update active orders metric based on status
	if statusUpdate.Status == "completed" || statusUpdate.Status == "cancelled" {
		metrics.ActiveOrders.Dec()
	}

	// Send notification about status change
	err = oc.NotificationService.SendOrderStatusUpdate(id, order.CustomerID, statusUpdate.Status)
	if err != nil {
		// Log the error but continue (non-blocking)
		fmt.Printf("Failed to send status update notification: %v\n", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Order status updated successfully",
		"order_id": id,
		"status":   statusUpdate.Status,
	})
}

// CreateBatchOrders handles creation of multiple orders in parallel
func (oc *OrderController) CreateBatchOrders(c *gin.Context) {
	var orders []model.Order
	if err := c.ShouldBindJSON(&orders); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Configure batch processing
	numWorkers := 10            // Số lượng worker xử lý song song
	timeout := 30 * time.Second // Thời gian timeout cho toàn bộ batch

	// Process orders in parallel using worker pool
	results := worker.ProcessBatch(orders, numWorkers, timeout)

	// Count successes and failures
	successful := 0
	failed := 0
	failedOrders := make([]map[string]interface{}, 0)

	for _, result := range results {
		if result.Error != nil {
			failed++
			failedOrders = append(failedOrders, map[string]interface{}{
				"order_id": result.OrderID,
				"error":    result.Error.Error(),
			})
		} else {
			successful++
		}
	}

	// Return summary
	c.JSON(http.StatusOK, gin.H{
		"total_orders":    len(orders),
		"successful":      successful,
		"failed":          failed,
		"failed_orders":   failedOrders,
		"processing_time": timeout,
	})
}

func (oc *OrderController) insertOrder(order *model.Order) error {
	query := `
		INSERT INTO orders (product_id, quantity, status, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	order.Status = "pending"
	order.CreatedAt = time.Now()

	return oc.DB.QueryRow(
		query,
		order.ProductID,
		order.Quantity,
		order.Status,
		order.CreatedAt,
	).Scan(&order.ID)
}

func (oc *OrderController) getOrderFromDB(orderID string) (*model.Order, error) {
	var order model.Order
	query := `
		SELECT id, product_id, quantity, status, created_at
		FROM orders
		WHERE id = $1`

	err := oc.DB.QueryRow(query, orderID).Scan(
		&order.ID,
		&order.ProductID,
		&order.Quantity,
		&order.Status,
		&order.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &order, nil
}
