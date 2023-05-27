package models

type GetRatesByProductRequest struct {
	ProductID string `json:"product_id" binding:"uuid"`
	AreaID    uint32 `json:"area_id" binding:"required"`
}
