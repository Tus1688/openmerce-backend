package models

type GetRatesByProductRequest struct {
	ProductID string `form:"product_id" binding:"uuid"`
	AreaID    uint32 `form:"area_id" binding:"required"`
}

type ShipOrderToCustomer struct {
	OrderId      uint64 `json:"order_id" binding:"required"`
	TrackingCode string `json:"tracking_code" binding:"required"`
}
