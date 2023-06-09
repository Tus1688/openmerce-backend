package models

type CreateReview struct {
	OrderID uint64 `json:"order_id" binding:"required"`
	Rating  uint8  `json:"rating" binding:"required"`
	Review  string `json:"review"`
}

// ReviewResponseCustomer is a specific struct for review response in customer page
type ReviewResponseCustomer struct {
	ID           string `json:"id"`
	ProductID    string `json:"product_id"`
	ProductName  string `json:"product_name"`
	ProductImage string `json:"product_image"`
	Rating       uint8  `json:"rating"`
	Review       string `json:"review"`
	CreatedAt    string `json:"created_at"`
}

type ReviewResponseGlobal struct {
	ID        string `json:"id"`
	Rating    uint8  `json:"rating"`
	Review    string `json:"review"`
	CreatedAt string `json:"created_at"`
	Customer  string `json:"customer"`
}
