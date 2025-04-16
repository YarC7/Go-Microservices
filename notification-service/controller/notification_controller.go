package controller

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go-microservices/notification-service/model"

	"github.com/gin-gonic/gin"
)

// NotificationController handles notification-related requests
type NotificationController struct {
	DB *sql.DB
}

// NewNotificationController creates a new notification controller
func NewNotificationController(db *sql.DB) *NotificationController {
	return &NotificationController{DB: db}
}

// CreateNotification handles creation of a new notification
func (nc *NotificationController) CreateNotification(c *gin.Context) {
	var notification model.Notification
	if err := c.BindJSON(&notification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	notification.CreatedAt = time.Now()

	var id int
	err := nc.DB.QueryRow(
		"INSERT INTO notifications (order_id, customer_id, message, status, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		notification.OrderID, notification.CustomerID, notification.Message, notification.Status, notification.CreatedAt).Scan(&id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	notification.ID = id
	c.JSON(http.StatusCreated, notification)
}

// GetNotifications returns all notifications
func (nc *NotificationController) GetNotifications(c *gin.Context) {
	rows, err := nc.DB.Query("SELECT id, order_id, customer_id, message, status, created_at, delivered_at FROM notifications")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var notifications []model.Notification
	for rows.Next() {
		var n model.Notification
		var deliveredAt sql.NullTime
		if err := rows.Scan(&n.ID, &n.OrderID, &n.CustomerID, &n.Message, &n.Status, &n.CreatedAt, &deliveredAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if deliveredAt.Valid {
			n.DeliveredAt = deliveredAt.Time
		}
		notifications = append(notifications, n)
	}

	c.JSON(http.StatusOK, notifications)
}

// GetNotification returns a specific notification by ID
func (nc *NotificationController) GetNotification(c *gin.Context) {
	id := c.Param("id")
	var notification model.Notification
	var deliveredAt sql.NullTime

	err := nc.DB.QueryRow("SELECT id, order_id, customer_id, message, status, created_at, delivered_at FROM notifications WHERE id = $1", id).
		Scan(&notification.ID, &notification.OrderID, &notification.CustomerID, &notification.Message, &notification.Status, &notification.CreatedAt, &deliveredAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if deliveredAt.Valid {
		notification.DeliveredAt = deliveredAt.Time
	}

	c.JSON(http.StatusOK, notification)
}

// GetCustomerNotifications returns all notifications for a customer
func (nc *NotificationController) GetCustomerNotifications(c *gin.Context) {
	customerID := c.Param("customerId")
	rows, err := nc.DB.Query("SELECT id, order_id, customer_id, message, status, created_at, delivered_at FROM notifications WHERE customer_id = $1", customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var notifications []model.Notification
	for rows.Next() {
		var n model.Notification
		var deliveredAt sql.NullTime
		if err := rows.Scan(&n.ID, &n.OrderID, &n.CustomerID, &n.Message, &n.Status, &n.CreatedAt, &deliveredAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if deliveredAt.Valid {
			n.DeliveredAt = deliveredAt.Time
		}
		notifications = append(notifications, n)
	}

	c.JSON(http.StatusOK, notifications)
}

// MarkDelivered marks a notification as delivered
func (nc *NotificationController) MarkDelivered(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	now := time.Now()
	result, err := nc.DB.Exec("UPDATE notifications SET delivered_at = $1 WHERE id = $2", now, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as delivered", "delivered_at": now})
}

// ProcessOrderStatusUpdate processes an order status update and creates a notification
func (nc *NotificationController) ProcessOrderStatusUpdate(c *gin.Context) {
	var update model.OrderStatusUpdate
	if err := c.BindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create a notification from the status update
	message := fmt.Sprintf("Your order #%d status has changed to: %s", update.OrderID, update.Status)
	now := time.Now()

	var id int
	err := nc.DB.QueryRow(
		"INSERT INTO notifications (order_id, customer_id, message, status, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		update.OrderID, update.CustomerID, message, update.Status, now).Scan(&id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// In a real application, you would send the notification through email, SMS, etc.
	c.JSON(http.StatusOK, gin.H{
		"message":         "Order status notification created",
		"notification_id": id,
	})
}
