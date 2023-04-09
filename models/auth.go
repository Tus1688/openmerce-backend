package models

type ReqEmailVerification struct {
	Email string `json:"email" binding:"required"`
}
