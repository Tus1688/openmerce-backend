package models

import (
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	ID             uuid.UUID `json:"id"`
	Email          string    `json:"email"`
	HashedPassword string
	PhoneNumber    string    `json:"phone_number"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	BirthDate      time.Time `json:"birth_date"`
	Gender         string    `json:"gender"`
}

type Staff struct {
	ID             uint   `json:"id"`
	Username       string `json:"username"`
	HashedPassword string
	Name           string `json:"name"`
	FinUser        bool   `json:"fin_user"`
	InvUser        bool   `json:"inv_user"`
	SysAdmin       bool   `json:"sys_admin"`
}
