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
