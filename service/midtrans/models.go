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

package midtrans

import "github.com/Tus1688/openmerce-backend/models"

// RequestSnap basically is a request to midtrans to create a snap page token for payment
type RequestSnap struct {
	TransactionDetails `json:"transaction_details"`
	ItemDetails        []models.CheckoutItem `json:"item_details"`
	CustomerDetails    `json:"customer_details"`
	Expiry             `json:"expiry"`
}

type ResponseSnap struct {
	Token       string `json:"token"`
	RedirectUrl string `json:"redirect_url"`
}

type ResponseErrorSnap struct {
	ErrorMessages []string `json:"error_messages"`
}

type TransactionDetails struct {
	OrderId     string `json:"order_id"`
	GrossAmount int    `json:"gross_amount"`
}

type CustomerDetails struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone,omitempty"`
}

type Expiry struct {
	StartTime string `json:"start_time"`
	// Unit is in days, hours, minutes
	Unit     string `json:"unit"`
	Duration int    `json:"duration"`
}

type WebhookNotification struct {
	TransactionStatus string `json:"transaction_status" binding:"required"`
	StatusCode        string `json:"status_code" binding:"required"`
	SignatureKey      string `json:"signature_key" binding:"required"`
	OrderId           string `json:"order_id" binding:"required"`
	GrossAmount       string `json:"gross_amount" binding:"required"`
	PaymentType       string `json:"payment_type" binding:"required"`
	// FraudStatus isn't available in OTC payment (indomaret, alfamart, etc)
	FraudStatus string `json:"fraud_status"`
}

type ResponseErrorDeleteOrder struct {
	StatusCode    string `json:"status_code"`
	StatusMessage string `json:"status_message"`
}
