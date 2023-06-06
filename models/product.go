package models

import (
	"mime/multipart"
)

// ProductCreate is the model for creating a new product on step 1
type ProductCreate struct {
	CategoryID   uint    `json:"category_id" binding:"required"`
	Name         string  `json:"name" binding:"required"`
	Description  string  `json:"description" binding:"required"`
	Price        uint    `json:"price" binding:"required"`
	Weight       float64 `json:"weight" binding:"required"`
	InitialStock uint    `json:"initial_stock" binding:"required"`
	Length       uint16  `json:"length" binding:"required"`
	Width        uint16  `json:"width" binding:"required"`
	Height       uint16  `json:"height" binding:"required"`
}

// ProductUpdate is the model for updating a product (also considered as step 1)
// there is no product update for step 2 (product images)
type ProductUpdate struct {
	ID          string  `json:"id" binding:"required,uuid"`
	CategoryID  uint    `json:"category_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       uint    `json:"price"`
	Weight      float64 `json:"weight"`
	Stock       uint    `json:"stock"`
	Length      uint16  `json:"length"`
	Width       uint16  `json:"width"`
	Height      uint16  `json:"height"`
}

type ProductImage struct {
	ProductID string                `form:"product_id" binding:"required,uuid"`
	Picture   *multipart.FileHeader `form:"picture" binding:"required"`
}

type ProductImageDelete struct {
	ProductID string `json:"product_id" binding:"required,uuid"`
	FileName  string `json:"file_name" binding:"required"`
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

type CategoryResponseCompact struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type CategoryUpdate struct {
	ID                 uint   `json:"id" binding:"required"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	HomePageVisibility *bool  `json:"homepage_visibility"`
}

// HomepageProductResponse is the model for the homepage product response
type HomepageProductResponse struct {
	CategoryID   uint              `json:"category_id"`
	CategoryName string            `json:"category_name"`
	CategoryDesc string            `json:"category_desc"`
	Products     []HomepageProduct `json:"products"`
}

// HomepageProduct is the sub model for HomepageProductResponse, and it also used for the product search response
type HomepageProduct struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Price    uint    `json:"price"`
	ImageUrl string  `json:"image"`
	Rating   float64 `json:"rating"`
	Sold     uint    `json:"sold"`
}

// ProductDetail is the model for product detail response (query by id)
type ProductDetail struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	Price            uint     `json:"price"`
	Weight           float64  `json:"weight"`
	CategoryName     string   `json:"category_name"`
	CumulativeReview float64  `json:"cumulative_review"`
	ImageUrls        []string `json:"image_urls"`
	Dimension        string   `json:"dimension"`
	Stock            uint     `json:"stock"`
}
