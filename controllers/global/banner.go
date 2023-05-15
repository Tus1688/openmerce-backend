package global

import (
	"github.com/Tus1688/openmerce-backend/database"
	"github.com/Tus1688/openmerce-backend/models"
	"github.com/gin-gonic/gin"
)

func GetHomeBanner(c *gin.Context) {
	var response []models.GetBanner
	rows, err := database.MysqlInstance.Query("SELECT id, file_name, href FROM homepage_banner")
	if err != nil {
		c.Status(500)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var banner models.GetBanner
		if err := rows.Scan(&banner.Id, &banner.ImageUrl, &banner.Href); err != nil {
			c.Status(500)
			return
		}
		response = append(response, banner)
	}
	if len(response) == 0 {
		c.Status(404)
		return
	}
	c.JSON(200, response)
}
