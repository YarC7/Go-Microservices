package main

import (
	"log"

	"go-microservices/order-service/cache"
	"go-microservices/order-service/controller"
	"go-microservices/order-service/db"
	"go-microservices/order-service/queue"
	"go-microservices/order-service/routes"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Initialize database connection
	database := db.GetDB()
	defer database.Close()

	// Initialize database schema
	db.InitSchema(database)

	// Initialize Redis
	if err := cache.InitRedis(); err != nil {
		log.Printf("Warning: Failed to initialize Redis: %v\n", err)
	}

	// Initialize RabbitMQ
	if err := queue.InitRabbitMQ(); err != nil {
		log.Printf("Warning: Failed to initialize RabbitMQ: %v\n", err)
	}
	defer queue.Close()

	// Declare queues
	orderQueue := queue.Config{
		QueueName:    "orders",
		RoutingKey:   "order.#",
		ExchangeName: "orders",
		ExchangeType: "topic",
	}
	if err := queue.DeclareQueue(orderQueue); err != nil {
		log.Printf("Warning: Failed to declare order queue: %v\n", err)
	}

	// Create order controller
	orderController := controller.NewOrderController(database)

	// Initialize router
	router := gin.Default()

	// Add prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Setup routes
	routes.SetupRoutes(router, orderController)

	// Start server
	log.Println("Order Service starting on port 8081...")
	if err := router.Run(":8081"); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
