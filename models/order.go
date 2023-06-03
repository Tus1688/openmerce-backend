package models

type OrderResponse struct {
	ID          uint64 `json:"id"`
	CreatedAt   string `json:"created_at"`
	GrossAmount uint   `json:"gross_amount"`
	Status      string `json:"status"`
	ItemCount   uint8  `json:"item_count"`
	Image       string `json:"image"`
	ProductName string `json:"product_name"`
}
