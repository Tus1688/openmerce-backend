package models

type PreCheckoutFreightItems struct {
	Weight float64 `json:"weight"`
	Volume float64 `json:"volume"`
}

type PreCheckoutItem struct {
	ProductId    string `json:"id"`
	ProductName  string `json:"name"`
	ProductPrice uint   `json:"price"`
	ProductImage string `json:"image"`
	Quantity     uint16 `json:"quantity"`
}
