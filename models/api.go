package models

type APICommonQueryID struct {
	ID int `form:"id" binding:"required"`
}

type APICommonQueryUUID struct {
	ID string `form:"id" binding:"required,uuid"`
}

type APICommonQuerySearch struct {
	Search string `form:"search" binding:"required"`
}
