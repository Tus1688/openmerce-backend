package models

import (
	"time"
	"unicode"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type ReqEmailVerification struct {
	Email string `json:"email" binding:"required"`
}

type ReqEmailVerificationConfirmation struct {
	Email string `json:"email" binding:"required"`
	Code  int    `json:"code" binding:"required"`
}

// happen after user has verified their email
type ReqNewAccount struct {
	Email     string    `json:"email" binding:"required"`
	Password  string    `json:"password" binding:"required"`
	FirstName string    `json:"first_name" binding:"required"`
	LastName  string    `json:"last_name" binding:"required"`
	BirthDate time.Time `json:"birth_date" binding:"required"`
	Gender    string    `json:"gender" binding:"required"`
}

type ReqLogin struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type CustomerAuth struct {
	ID             uuid.UUID
	HashedPassword string
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

// validate password is at least 8 characters long and contains at least one uppercase letter, one lowercase letter, and one number
func (s *ReqNewAccount) PasswordIsValid() bool {
	if len(s.Password) < 8 {
		return false
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasSpecial bool
		hasNumber  bool
	)

	for _, char := range s.Password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		case unicode.IsNumber(char):
			hasNumber = true
		}
		if hasUpper && hasLower && hasSpecial && hasNumber {
			return true
		}
	}

	return false
}

func (s *ReqNewAccount) HashPassword() error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(s.Password), 10)
	if err != nil {
		return err
	}
	s.Password = string(bytes)
	return nil
}

func (s *CustomerAuth) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(s.HashedPassword), []byte(password))
	return err == nil
}
