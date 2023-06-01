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

// CheckoutItem is used to store the product information to be purchased into third party payment gateway
type CheckoutItem struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Price    int    `json:"price"`
	Quantity int    `json:"quantity"`
}
