package models

// ProductCreate1 is the model for creating a new product on step 1
type ProductCreate1 struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
	Price       uint   `json:"price" binding:"required"`
	Weight      uint16 `json:"weight" binding:"required"`
}
