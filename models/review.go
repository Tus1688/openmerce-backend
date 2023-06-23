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

type CreateReview struct {
	OrderID uint64 `json:"order_id" binding:"required"`
	Rating  uint8  `json:"rating" binding:"required"`
	Review  string `json:"review"`
}

// ReviewResponseCustomer is a specific struct for review response in customer page
type ReviewResponseCustomer struct {
	ID           string `json:"id"`
	ProductID    string `json:"product_id"`
	ProductName  string `json:"product_name"`
	ProductImage string `json:"product_image"`
	Rating       uint8  `json:"rating"`
	Review       string `json:"review"`
	CreatedAt    string `json:"created_at"`
}

type ReviewResponseGlobal struct {
	ID        string `json:"id"`
	Rating    uint8  `json:"rating"`
	Review    string `json:"review"`
	CreatedAt string `json:"created_at"`
	Customer  string `json:"customer"`
}
