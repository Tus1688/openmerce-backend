package models

import (
	"time"

	"github.com/google/uuid"
)

type ReqEmailVerification struct {
	Email string `json:"email" binding:"required"`
}

type Customer struct {
	ID             uuid.UUID `json:"id"`
	Email          string    `json:"email"`
	HashedPassword string    `json:"hashed_password"`
	PhoneNumber    string    `json:"phone_number"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	BirthDate      time.Time `json:"birth_date"`
	Gender         string    `json:"gender"`
}
