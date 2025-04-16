package routes

import (
	"go-microservices/product-service/controller"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures the API routes for the product service
func SetupRoutes(router *gin.Engine, productController *controller.ProductController) {
	// Product routes
	router.POST("/products", productController.CreateProduct)
	router.GET("/products", productController.GetProducts)
	router.GET("/products/:id", productController.GetProduct)
	router.PUT("/products/:id", productController.UpdateProduct)
	router.DELETE("/products/:id", productController.DeleteProduct)
}
