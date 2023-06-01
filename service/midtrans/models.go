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
