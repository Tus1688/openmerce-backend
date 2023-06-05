package models

type CreateReview struct {
	OrderID uint64 `json:"order_id" binding:"required"`
	Rating  uint8  `json:"rating" binding:"required"`
	Review  string `json:"review"`
}
