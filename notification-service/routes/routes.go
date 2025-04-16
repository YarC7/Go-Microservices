package routes

import (
	"go-microservices/notification-service/controller"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures the API routes for the notification service
func SetupRoutes(router *gin.Engine, notificationController *controller.NotificationController) {
	// Notification routes
	router.POST("/notifications", notificationController.CreateNotification)
	router.GET("/notifications", notificationController.GetNotifications)
	router.GET("/notifications/:id", notificationController.GetNotification)
	router.GET("/notifications/customer/:customerId", notificationController.GetCustomerNotifications)
	router.PUT("/notifications/:id/deliver", notificationController.MarkDelivered)

	// Order status update route
	router.POST("/notifications/order-status", notificationController.ProcessOrderStatusUpdate)
}
