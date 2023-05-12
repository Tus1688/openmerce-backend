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

// ReqNewAccount happen after user has verified their email
type ReqNewAccount struct {
	Email     string    `json:"email" binding:"required"`
	Password  string    `json:"password" binding:"required"`
	FirstName string    `json:"first_name" binding:"required"`
	LastName  string    `json:"last_name" binding:"required"`
	BirthDate time.Time `json:"birth_date" binding:"required"`
	Gender    string    `json:"gender" binding:"required"`
}

type ReqLoginCustomer struct {
	Email      string `json:"email" binding:"required"`
	Password   string `json:"password" binding:"required"`
	RememberMe bool   `json:"remember_me"`
}

type ReqLoginStaff struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	RememberMe bool   `json:"remember_me"`
}

type CustomerAuth struct {
	ID             uuid.UUID
	HashedPassword string
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
}

type StaffAuth struct {
	ID             uint   `json:"id"`
	Username       string `json:"username"`
	HashedPassword string
	FinUser        bool `json:"fin_user"`
	InvUser        bool `json:"inv_user"`
	SysAdmin       bool `json:"sys_admin"`
}

type NewStaff struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Name     string `json:"name" binding:"required"`
	FinUser  bool   `json:"fin_user"`
	InvUser  bool   `json:"inv_user"`
	SysAdmin bool   `json:"sys_admin"`
}

// PasswordIsValid validate password is at least 8 characters long and contains at least one uppercase letter, one lowercase letter, and one number
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

func (s *NewStaff) PasswordIsValid() bool {
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

func (s *NewStaff) HashPassword() error {
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

func (s *StaffAuth) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(s.HashedPassword), []byte(password))
	return err == nil
}
