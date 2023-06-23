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
