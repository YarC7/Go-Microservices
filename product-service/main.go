package main

import (
	"log"

	"go-microservices/product-service/controller"
	"go-microservices/product-service/db"
	"go-microservices/product-service/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database connection
	database := db.GetDB()
	defer database.Close()

	// Initialize database schema
	db.InitSchema(database)

	// Create product controller
	productController := controller.NewProductController(database)

	// Initialize router
	router := gin.Default()

	// Setup routes
	routes.SetupRoutes(router, productController)

	// Start server
	log.Println("Product Service starting on port 8080...")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
