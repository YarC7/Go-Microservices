package controller

import (
	"database/sql"
	"net/http"
	"strconv"

	"go-microservices/inventory-service/model"

	"github.com/gin-gonic/gin"
)

// InventoryController handles inventory-related requests
type InventoryController struct {
	DB *sql.DB
}

// NewInventoryController creates a new inventory controller
func NewInventoryController(db *sql.DB) *InventoryController {
	return &InventoryController{DB: db}
}

// CreateInventory handles creation of a new inventory item
func (ic *InventoryController) CreateInventory(c *gin.Context) {
	var inventory model.Inventory
	if err := c.BindJSON(&inventory); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var id int
	err := ic.DB.QueryRow(
		"INSERT INTO inventory (product_id, quantity, sku, location) VALUES ($1, $2, $3, $4) RETURNING id",
		inventory.ProductID, inventory.Quantity, inventory.SKU, inventory.Location).Scan(&id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	inventory.ID = id
	c.JSON(http.StatusCreated, inventory)
}

// GetInventories returns all inventory items
func (ic *InventoryController) GetInventories(c *gin.Context) {
	rows, err := ic.DB.Query("SELECT id, product_id, quantity, sku, location FROM inventory")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var inventories []model.Inventory
	for rows.Next() {
		var i model.Inventory
		if err := rows.Scan(&i.ID, &i.ProductID, &i.Quantity, &i.SKU, &i.Location); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		inventories = append(inventories, i)
	}

	c.JSON(http.StatusOK, inventories)
}

// GetInventory returns a specific inventory item by ID
func (ic *InventoryController) GetInventory(c *gin.Context) {
	id := c.Param("id")
	var inventory model.Inventory

	err := ic.DB.QueryRow("SELECT id, product_id, quantity, sku, location FROM inventory WHERE id = $1", id).
		Scan(&inventory.ID, &inventory.ProductID, &inventory.Quantity, &inventory.SKU, &inventory.Location)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inventory item not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, inventory)
}

// UpdateInventory updates an inventory item
func (ic *InventoryController) UpdateInventory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var inventory model.Inventory
	if err := c.BindJSON(&inventory); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := ic.DB.Exec(
		"UPDATE inventory SET product_id = $1, quantity = $2, sku = $3, location = $4 WHERE id = $5",
		inventory.ProductID, inventory.Quantity, inventory.SKU, inventory.Location, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inventory item not found"})
		return
	}

	inventory.ID = id
	c.JSON(http.StatusOK, inventory)
}

// DeleteInventory deletes an inventory item
func (ic *InventoryController) DeleteInventory(c *gin.Context) {
	id := c.Param("id")

	result, err := ic.DB.Exec("DELETE FROM inventory WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inventory item not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Inventory item deleted successfully"})
}

// CheckInventory checks if there's enough inventory for a product
func (ic *InventoryController) CheckInventory(c *gin.Context) {
	var check model.InventoryCheck
	if err := c.BindJSON(&check); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var quantity int
	err := ic.DB.QueryRow("SELECT quantity FROM inventory WHERE product_id = $1", check.ProductID).Scan(&quantity)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusOK, model.InventoryResponse{
			Available: false,
			Message:   "Product not found in inventory",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	available := quantity >= check.Quantity
	message := ""
	if !available {
		message = "Not enough inventory"
	}

	c.JSON(http.StatusOK, model.InventoryResponse{
		Available: available,
		Message:   message,
	})
}
