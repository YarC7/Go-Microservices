package main

import (
	"log"

	"go-microservices/payment-service/controller"
	"go-microservices/payment-service/db"
	"go-microservices/payment-service/routes"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Initialize database connection
	database := db.GetDB()
	defer database.Close()

	// Initialize database schema
	db.InitSchema(database)

	// Create payment controller
	paymentController := controller.NewPaymentController(database)

	// Initialize router
	router := gin.Default()

	// Add prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Setup routes
	routes.SetupRoutes(router, paymentController)

	// Start server
	log.Println("Payment Service starting on port 8084...")
	if err := router.Run(":8084"); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}