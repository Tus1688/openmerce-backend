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

type OrderDetailResponse struct {
	ID                uint64               `json:"id"`
	Status            string               `json:"status"`
	StatusDescription string               `json:"status_description"`
	CreatedAt         string               `json:"created_at"`
	ItemList          []PreCheckoutItem    `json:"item_list"`
	Courier           string               `json:"courier"`
	TrackingCode      string               `json:"tracking_code"`
	AddressDetail     AddressOrderResponse `json:"address_detail"`
	ItemCost          uint                 `json:"item_cost"`
	ShippingCost      uint                 `json:"shipping_cost"`
	TotalCost         uint                 `json:"total_cost"`
}

// AddressOrderResponse is a specific struct for order detail response
type AddressOrderResponse struct {
	RecipientName string `json:"recipient_name"`
	PhoneNumber   string `json:"phone_number"`
	FullAddress   string `json:"full_address"`
	ShippingArea  string `json:"shipping_area"`
}
