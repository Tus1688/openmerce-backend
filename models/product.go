package models

// ProductCreate is the model for creating a new product on step 1
type ProductCreate struct {
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description" binding:"required"`
	Price        uint   `json:"price" binding:"required"`
	Weight       uint16 `json:"weight" binding:"required"`
	InitialStock uint   `json:"initial_stock" binding:"required"`
}
