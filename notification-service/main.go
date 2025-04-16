package main

import (
	"log"

	"go-microservices/notification-service/controller"
	"go-microservices/notification-service/db"
	"go-microservices/notification-service/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database connection
	database := db.GetDB()
	defer database.Close()

	// Initialize database schema
	db.InitSchema(database)

	// Create notification controller
	notificationController := controller.NewNotificationController(database)

	// Initialize router
	router := gin.Default()

	// Setup routes
	routes.SetupRoutes(router, notificationController)

	// Start server
	log.Println("Notification Service starting on port 8083...")
	if err := router.Run(":8083"); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
