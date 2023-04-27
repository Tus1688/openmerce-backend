package models

import "mime/multipart"

// ProductCreate is the model for creating a new product on step 1
type ProductCreate struct {
	CategoryID   uint   `json:"category_id" binding:"required"`
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description" binding:"required"`
	Price        uint   `json:"price" binding:"required"`
	Weight       uint16 `json:"weight" binding:"required"`
	InitialStock uint   `json:"initial_stock" binding:"required"`
}

type ProductImage struct {
	ProductID string                `form:"product_id" binding:"required"`
	Picture   *multipart.FileHeader `form:"picture" binding:"required"`
}

type CategoryCreate struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type CategoryResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
