package controller

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"go-microservices/payment-service/model"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
)

type PaymentController struct {
	db *sql.DB
}

func NewPaymentController(db *sql.DB) *PaymentController {
	// Initialize Stripe
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	if stripe.Key == "" {
		log.Fatal("STRIPE_SECRET_KEY is required but not set")
	}	
	return &PaymentController{
		db: db,
	}
}

// CreatePayment creates a new payment intent with Stripe
func (pc *PaymentController) CreatePayment(c *gin.Context) {
	var req model.PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert amount to cents for Stripe (Stripe expects amounts in cents)
	amountCents := int64(req.Amount * 100)

	// Create payment intent with Stripe
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amountCents),
		Currency: stripe.String(req.Currency),
		Metadata: map[string]string{
			"order_id":    strconv.Itoa(req.OrderID),
			"customer_id": strconv.Itoa(req.CustomerID),
		},
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment intent: " + err.Error()})
		return
	}

	// Save payment to database
	payment := model.Payment{
		OrderID:            req.OrderID,
		CustomerID:         req.CustomerID,
		Amount:             req.Amount,
		Currency:           req.Currency,
		Status:             model.PaymentStatusPending,
		StripePaymentID:    pi.ID,
		StripeClientSecret: pi.ClientSecret,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	query := `
		INSERT INTO payments (order_id, customer_id, amount, currency, status, stripe_payment_id, stripe_client_secret, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	err = pc.db.QueryRow(query, payment.OrderID, payment.CustomerID, payment.Amount, payment.Currency, 
		payment.Status, payment.StripePaymentID, payment.StripeClientSecret, payment.CreatedAt, payment.UpdatedAt).Scan(&payment.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save payment: " + err.Error()})
		return
	}

	response := model.PaymentResponse{
		Payment:      payment,
		ClientSecret: pi.ClientSecret,
		Message:      "Payment intent created successfully",
	}

	c.JSON(http.StatusCreated, response)
}

// ConfirmPayment confirms a payment and updates the status
func (pc *PaymentController) ConfirmPayment(c *gin.Context) {
	var req model.PaymentConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Retrieve payment intent from Stripe
	pi, err := paymentintent.Get(req.PaymentIntentID, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve payment intent: " + err.Error()})
		return
	}

	// Update payment status in database
	status := model.PaymentStatusPending
	switch pi.Status {
	case stripe.PaymentIntentStatusSucceeded:
		status = model.PaymentStatusSucceeded
	case stripe.PaymentIntentStatusCanceled:
		status = model.PaymentStatusCanceled
	case stripe.PaymentIntentStatusProcessing:
		status = model.PaymentStatusPending
	default:
		status = model.PaymentStatusFailed
	}

	query := `
		UPDATE payments 
		SET status = $1, payment_method = $2, updated_at = $3
		WHERE stripe_payment_id = $4
		RETURNING id, order_id, customer_id, amount, currency, status, stripe_payment_id, payment_method, created_at, updated_at
	`

	var payment model.Payment
	err = pc.db.QueryRow(query, status, string(pi.PaymentMethod.Type), time.Now(), pi.ID).Scan(
		&payment.ID, &payment.OrderID, &payment.CustomerID, &payment.Amount, &payment.Currency,
		&payment.Status, &payment.StripePaymentID, &payment.PaymentMethod, &payment.CreatedAt, &payment.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment: " + err.Error()})
		return
	}

	response := model.PaymentResponse{
		Payment: payment,
		Message: "Payment status updated successfully",
	}

	c.JSON(http.StatusOK, response)
}

// GetPayment retrieves a payment by ID
func (pc *PaymentController) GetPayment(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment ID"})
		return
	}

	query := `
		SELECT id, order_id, customer_id, amount, currency, status, stripe_payment_id, 
		       COALESCE(payment_method, '') as payment_method, created_at, updated_at
		FROM payments WHERE id = $1
	`

	var payment model.Payment
	err = pc.db.QueryRow(query, id).Scan(
		&payment.ID, &payment.OrderID, &payment.CustomerID, &payment.Amount, &payment.Currency,
		&payment.Status, &payment.StripePaymentID, &payment.PaymentMethod, &payment.CreatedAt, &payment.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve payment: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// GetPaymentsByOrder retrieves all payments for an order
func (pc *PaymentController) GetPaymentsByOrder(c *gin.Context) {
	orderIDParam := c.Param("orderId")
	orderID, err := strconv.Atoi(orderIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	query := `
		SELECT id, order_id, customer_id, amount, currency, status, stripe_payment_id,
		       COALESCE(payment_method, '') as payment_method, created_at, updated_at
		FROM payments WHERE order_id = $1 ORDER BY created_at DESC
	`

	rows, err := pc.db.Query(query, orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve payments: " + err.Error()})
		return
	}
	defer rows.Close()

	var payments []model.Payment
	for rows.Next() {
		var payment model.Payment
		err := rows.Scan(
			&payment.ID, &payment.OrderID, &payment.CustomerID, &payment.Amount, &payment.Currency,
			&payment.Status, &payment.StripePaymentID, &payment.PaymentMethod, &payment.CreatedAt, &payment.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan payment: " + err.Error()})
			return
		}
		payments = append(payments, payment)
	}

	c.JSON(http.StatusOK, payments)
}

// HealthCheck returns the health status of the payment service
func (pc *PaymentController) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "payment-service",
		"time":    time.Now().UTC(),
	})
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}