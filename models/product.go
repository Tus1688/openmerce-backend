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
	Name               string `json:"name" binding:"required"`
	Description        string `json:"description" binding:"required"`
	HomePageVisibility *bool  `json:"homepage_visibility" binding:"required"`
}

type CategoryResponse struct {
	ID                 uint   `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	HomePageVisibility bool   `json:"homepage_visibility"`
}

type CategoryUpdate struct {
	ID                 uint   `json:"id" binding:"required"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	HomePageVisibility *bool  `json:"homepage_visibility"`
}

// HomepageProductResponse is the model for the homepage product response
type HomepageProductResponse struct {
	CategoryID   uint   `json:"category_id"`
	CategoryName string `json:"category_name"`
	CategoryDesc string `json:"category_desc"`
	Products     []struct {
		ProductID     string  `json:"id"`
		ProductName   string  `json:"name"`
		ProductPrice  uint    `json:"price"`
		ProductImage  string  `json:"image"`
		ProductRating float64 `json:"rating"`
	} `json:"products"`
}
