package models

import "mime/multipart"

type InsertBanner struct {
	Picture *multipart.FileHeader `form:"picture" binding:"required"`
	Href    string                `form:"href" binding:"required"`
}

type GetBanner struct {
	Id       uint   `json:"id"`
	ImageUrl string `json:"image_url"`
	Href     string `json:"href"`
}
