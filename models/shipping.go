package models

type GetRatesByProductRequest struct {
	ProductID string `form:"product_id" binding:"uuid"`
	AreaID    uint32 `form:"area_id" binding:"required"`
}
