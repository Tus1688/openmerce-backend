package models

import (
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

type ListStaff struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	FinUser  bool   `json:"fin_user"`
	InvUser  bool   `json:"inv_user"`
	SysAdmin bool   `json:"sys_admin"`
}

type UpdateStaff struct {
	ID       uint   `json:"id" binding:"required"`
	Password string `json:"password"`
	Name     string `json:"name"`
	FinUser  *bool  `json:"fin_user" binding:"required"`
	InvUser  *bool  `json:"inv_user" binding:"required"`
	SysAdmin *bool  `json:"sys_admin" binding:"required"`
}

func (s *UpdateStaff) PasswordIsValid() bool {
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
func (s *UpdateStaff) HashPassword() error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(s.Password), 10)
	if err != nil {
		return err
	}
	s.Password = string(bytes)
	return nil
}
