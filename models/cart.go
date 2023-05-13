package models

type CartItemResponse struct {
	ProductId    string `json:"id"`
	ProductName  string `json:"name"`
	ProductPrice uint   `json:"price"`
	Quantity     uint16 `json:"quantity"`
	CurrentStock uint16 `json:"curr_stock"`
}

type CartInsert struct {
	ProductId string `json:"id" binding:"required,uuid"`
	Quantity  uint16 `json:"quantity" binding:"required"`
}
