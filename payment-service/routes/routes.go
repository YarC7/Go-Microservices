package routes

import (
	"go-microservices/payment-service/controller"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures the payment service routes
func SetupRoutes(router *gin.Engine, paymentController *controller.PaymentController) {
	// Health check
	router.GET("/health", paymentController.HealthCheck)

	// Payment routes
	paymentRoutes := router.Group("/payments")
	{
		paymentRoutes.POST("/", paymentController.CreatePayment)            // Create payment intent
		paymentRoutes.POST("/confirm", paymentController.ConfirmPayment)    // Confirm payment
		paymentRoutes.GET("/:id", paymentController.GetPayment)            // Get payment by ID
		paymentRoutes.GET("/order/:orderId", paymentController.GetPaymentsByOrder) // Get payments by order ID
	}
}