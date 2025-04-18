package routes

import (
	"go-microservices/order-service/controller"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures the API routes for the order service
func SetupRoutes(router *gin.Engine, orderController *controller.OrderController) {
	// Order routes
	router.POST("/orders", orderController.CreateOrder)
	router.POST("/orders/batch", orderController.CreateBatchOrders)
	router.GET("/orders", orderController.GetOrders)
	router.GET("/orders/:id", orderController.GetOrder)
	router.PUT("/orders/:id", orderController.UpdateOrder)
	router.DELETE("/orders/:id", orderController.DeleteOrder)
	router.PATCH("/orders/:id/status", orderController.UpdateOrderStatus)
}
