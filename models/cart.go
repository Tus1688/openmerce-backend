package models

type CartItemResponse struct {
	ProductId    string `json:"id"`
	ProductName  string `json:"name"`
	ProductPrice uint   `json:"price"`
	ProductImage string `json:"image"`
	Quantity     uint16 `json:"quantity"`
	CurrentStock uint16 `json:"curr_stock"`
	Checked      bool   `json:"checked"`
}

type CartInsert struct {
	ProductId string `json:"id" binding:"required,uuid"`
	Quantity  uint16 `json:"quantity" binding:"required"`
}

// CartCheck is used to tick or un-tick a product in cart (to be purchased)
type CartCheck struct {
	ProductID string `json:"id" binding:"required,uuid"`
	State     *bool  `json:"state" binding:"required"`
}

type CheckAll struct {
	State *bool `json:"state" binding:"required"`
}
