package global

import (
	"github.com/Tus1688/openmerce-backend/database"
	"github.com/Tus1688/openmerce-backend/models"
	"github.com/gin-gonic/gin"
)

func GetCategory(c *gin.Context) {
	var response []models.CategoryResponseCompact
	rows, err := database.MysqlInstance.
		Query("SELECT id, name FROM categories WHERE deleted_at IS NULL")
	if err != nil {
		c.Status(500)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var category models.CategoryResponseCompact
		if err := rows.Scan(&category.ID, &category.Name); err != nil {
			c.Status(500)
			return
		}
		response = append(response, category)
	}
	if len(response) == 0 {
		c.Status(404)
		return
	}
	c.JSON(200, response)
}
