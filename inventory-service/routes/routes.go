package routes

import (
	"go-microservices/inventory-service/controller"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures the API routes for the inventory service
func SetupRoutes(router *gin.Engine, inventoryController *controller.InventoryController) {
	// Inventory routes
	router.POST("/inventory", inventoryController.CreateInventory)
	router.GET("/inventory", inventoryController.GetInventories)
	router.GET("/inventory/:id", inventoryController.GetInventory)
	router.PUT("/inventory/:id", inventoryController.UpdateInventory)
	router.DELETE("/inventory/:id", inventoryController.DeleteInventory)

	// Inventory check route for order service
	router.POST("/inventory/check", inventoryController.CheckInventory)
}
