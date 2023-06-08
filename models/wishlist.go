package models

type WishlistItemResponse struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Price    uint    `json:"price"`
	ImageUrl string  `json:"image"`
	Rating   float64 `json:"rating"`
	Sold     uint    `json:"sold"`
}
