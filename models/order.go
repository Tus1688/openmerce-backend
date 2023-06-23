// Copyright (c) 2023. Tus1688
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package models

type OrderResponse struct {
	ID          uint64 `json:"id"`
	CreatedAt   string `json:"created_at"`
	GrossAmount uint   `json:"gross_amount"`
	Status      string `json:"status"`
	ItemCount   uint8  `json:"item_count"`
	Image       string `json:"image"`
	ProductName string `json:"product_name"`
	Reviewed    bool   `json:"reviewed"`
}

type OrderDetailResponse struct {
	ID                uint64                `json:"id"`
	Status            string                `json:"status"`
	StatusDescription string                `json:"status_description"`
	PaymentType       string                `json:"payment_type"`
	PaymentUrl        string                `json:"payment_url,omitempty"`
	CreatedAt         string                `json:"created_at"`
	ItemList          []ItemListOrderDetail `json:"item_list"`
	Courier           string                `json:"courier"`
	TrackingCode      string                `json:"tracking_code"`
	AddressDetail     AddressOrderResponse  `json:"address_detail"`
	ItemCost          uint                  `json:"item_cost"`
	ShippingCost      uint                  `json:"shipping_cost"`
	TotalCost         uint                  `json:"total_cost"`
}

type ItemListOrderDetail struct {
	OrderID uint64 `json:"order_id"`
	PreCheckoutItem
	Reviewed bool `json:"reviewed"`
}

// AddressOrderResponse is a specific struct for order detail response
type AddressOrderResponse struct {
	RecipientName string `json:"recipient_name"`
	PhoneNumber   string `json:"phone_number"`
	FullAddress   string `json:"full_address"`
	ShippingArea  string `json:"shipping_area"`
}

type OrderResponseStaff struct {
	OrderResponse
	Shipped    bool `json:"shipped"`
	Paid       bool `json:"paid,omitempty"`
	NeedRefund bool `json:"need_refund,omitempty"`
}
