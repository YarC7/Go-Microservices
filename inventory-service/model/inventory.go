package model

// Inventory represents an inventory item for a product
type Inventory struct {
	ID        int    `json:"id"`
	ProductID int    `json:"product_id"`
	Quantity  int    `json:"quantity"`
	SKU       string `json:"sku"`
	Location  string `json:"location"`
}

// InventoryCheck is used for checking if an order can be fulfilled
type InventoryCheck struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

// InventoryResponse is the response when checking inventory
type InventoryResponse struct {
	Available bool   `json:"available"`
	Message   string `json:"message,omitempty"`
}
