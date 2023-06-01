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

type PreCheckoutFreight struct {
	ProductCode string `json:"product_code"`
	CourierName string `json:"courier_name"`
	ProductName string `json:"product_name"`
	Etd         string `json:"etd"`
	Rates       int    `json:"rates"`
}

// CheckoutItemInternal is used to store the product information that also need to put in the db
type CheckoutItemInternal struct {
	CheckoutItem
	Description string  `json:"description"`
	Weight      float64 `json:"weight"`
}

// CheckoutItem is used to store the product information to be purchased into midtrans API
type CheckoutItem struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Price    int    `json:"price"`
	Quantity int    `json:"quantity"`
}

type CheckoutRequest struct {
	CourierCode string `json:"courier_code" binding:"required"`
	AddressCode string `json:"address_code" binding:"required,uuid"`
}
