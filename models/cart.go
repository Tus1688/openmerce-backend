package models

type CartItemResponse struct {
	ProductId    string `json:"id"`
	ProductName  string `json:"name"`
	ProductPrice uint   `json:"price"`
	Quantity     uint16 `json:"quantity"`
}