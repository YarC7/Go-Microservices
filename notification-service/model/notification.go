package model

import "time"

// Notification represents a notification about an order
type Notification struct {
	ID          int       `json:"id"`
	OrderID     int       `json:"order_id"`
	CustomerID  int       `json:"customer_id"`
	Message     string    `json:"message"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	DeliveredAt time.Time `json:"delivered_at,omitempty"`
}

// OrderStatusUpdate used to receive order status updates
type OrderStatusUpdate struct {
	OrderID    int    `json:"order_id"`
	CustomerID int    `json:"customer_id"`
	Status     string `json:"status"`
}
