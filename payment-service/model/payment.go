package model

import (
	"time"
)

// Payment represents a payment transaction
type Payment struct {
	ID                int       `json:"id" db:"id"`
	OrderID           int       `json:"order_id" db:"order_id"`
	CustomerID        int       `json:"customer_id" db:"customer_id"`
	Amount            float64   `json:"amount" db:"amount"`
	Currency          string    `json:"currency" db:"currency"`
	Status            string    `json:"status" db:"status"`
	StripePaymentID   string    `json:"stripe_payment_id" db:"stripe_payment_id"`
	StripeClientSecret string   `json:"stripe_client_secret,omitempty" db:"stripe_client_secret"`
	PaymentMethod     string    `json:"payment_method" db:"payment_method"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// PaymentRequest represents a payment creation request
type PaymentRequest struct {
	OrderID    int     `json:"order_id" binding:"required"`
	CustomerID int     `json:"customer_id" binding:"required"`
	Amount     float64 `json:"amount" binding:"required,min=0.01"`
	Currency   string  `json:"currency" binding:"required"`
}

// PaymentConfirmRequest represents a payment confirmation request
type PaymentConfirmRequest struct {
	PaymentIntentID string `json:"payment_intent_id" binding:"required"`
}

// PaymentResponse represents a payment response
type PaymentResponse struct {
	Payment      Payment `json:"payment"`
	ClientSecret string  `json:"client_secret,omitempty"`
	Message      string  `json:"message,omitempty"`
}

// PaymentStatus constants
const (
	PaymentStatusPending   = "pending"
	PaymentStatusSucceeded = "succeeded"
	PaymentStatusFailed    = "failed"
	PaymentStatusCanceled  = "canceled"
)