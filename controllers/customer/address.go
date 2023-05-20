package customer

import (
	"github.com/Tus1688/openmerce-backend/database"
	"github.com/Tus1688/openmerce-backend/models"
	"github.com/gin-gonic/gin"
)

// GetSuggestArea returns a list of suggested provinces based on the query
func GetSuggestArea(c *gin.Context) {
	var request models.APICommonQuerySearch
	if err := c.ShouldBindQuery(&request); err != nil {
		c.Status(400)
		return
	}
	rows, err := database.MysqlInstance.
		Query("SELECT id, full_name FROM shipping_areas WHERE MATCH(full_name) AGAINST(? IN NATURAL LANGUAGE MODE) LIMIT 30", request.Search)
	if err != nil {
		c.Status(500)
		return
	}
	defer rows.Close()
	var res []models.AreaResponse
	for rows.Next() {
		var area models.AreaResponse
		if err := rows.Scan(&area.ID, &area.Name); err != nil {
			c.Status(500)
			return
		}
		res = append(res, area)
	}
	if len(res) == 0 {
		c.Status(404)
		return
	}
	// this endpoint is not going to change on a daily basis, so we can cache it for a day
	c.Header("Cache-Control", "public, max-age=86400, immutable")
	c.JSON(200, res)
}
