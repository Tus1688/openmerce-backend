package models

import "time"

type CustomerProfile struct {
	Email       string    `json:"email"`
	PhoneNumber string    `json:"phone_number"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	BirthDate   time.Time `json:"birth_date"`
	Gender      string    `json:"gender"`
}
