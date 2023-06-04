package models

type CreateAddress struct {
	Label         string `json:"label" binding:"required"`
	FullAddress   string `json:"full_address" binding:"required"`
	Note          string `json:"note"`
	RecipientName string `json:"recipient_name" binding:"required"`
	PhoneNumber   string `json:"phone_number" binding:"required"`
	ShippingArea  uint32 `json:"shipping_area" binding:"required"`
	PostalCode    string `json:"postal_code" binding:"required"`
}

type AddressResponse struct {
	ID            string `json:"id"`
	Label         string `json:"label"`
	FullAddress   string `json:"full_address"`
	Note          string `json:"note"`
	RecipientName string `json:"recipient_name"`
	PhoneNumber   string `json:"phone_number"`
}

type AddressResponseDetail struct {
	ID            string `json:"id"`
	Label         string `json:"label"`
	FullAddress   string `json:"full_address"`
	Note          string `json:"note"`
	RecipientName string `json:"recipient_name"`
	PhoneNumber   string `json:"phone_number"`
	ShippingArea  string `json:"shipping_area"`
	AreaID        uint32 `json:"area_id"`
	PostalCode    string `json:"postal_code"`
}

type UpdateAddress struct {
	ID            string `json:"id" binding:"required"`
	Label         string `json:"label"`
	FullAddress   string `json:"full_address"`
	Note          string `json:"note"`
	RecipientName string `json:"recipient_name"`
	PhoneNumber   string `json:"phone_number"`
	ShippingArea  uint32 `json:"shipping_area"`
	PostalCode    string `json:"postal_code"`
}
