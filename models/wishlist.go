package models

type WishlistItemResponse struct {
	ProductId    string `json:"id"`
	ProductName  string `json:"name"`
	ProductPrice uint   `json:"price"`
	ProductImage string `json:"image"`
}
