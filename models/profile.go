package models

import (
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

type CustomerProfile struct {
	Email       string    `json:"email"`
	PhoneNumber string    `json:"phone_number"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	BirthDate   time.Time `json:"birth_date"`
	Gender      string    `json:"gender"`
}

type ChangePassword struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func (s *ChangePassword) PasswordIsValid() bool {
	if len(s.NewPassword) < 8 {
		return false
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasSpecial bool
		hasNumber  bool
	)

	for _, char := range s.NewPassword {
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

func (s *ChangePassword) HashPassword() error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(s.NewPassword), 10)
	if err != nil {
		return err
	}
	s.NewPassword = string(bytes)
	return nil
}

func (s *ChangePassword) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(password), []byte(s.OldPassword))
	return err == nil
}
