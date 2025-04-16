package main

import (
	"log"

	"go-microservices/inventory-service/controller"
	"go-microservices/inventory-service/db"
	"go-microservices/inventory-service/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database connection
	database := db.GetDB()
	defer database.Close()

	// Initialize database schema
	db.InitSchema(database)

	// Create inventory controller
	inventoryController := controller.NewInventoryController(database)

	// Initialize router
	router := gin.Default()

	// Setup routes
	routes.SetupRoutes(router, inventoryController)

	// Start server
	log.Println("Inventory Service starting on port 8082...")
	if err := router.Run(":8082"); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
