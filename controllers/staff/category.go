package staff

import (
	"strings"

	"github.com/Tus1688/openmerce-backend/database"
	"github.com/Tus1688/openmerce-backend/models"
	"github.com/gin-gonic/gin"
)

func GetCategories(c *gin.Context) {
	var categories []models.CategoryResponse
	rows, err := database.MysqlInstance.Query("SELECT id, name, description FROM categories")
	if err != nil {
		c.Status(500)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var category models.CategoryResponse
		if err := rows.Scan(&category.ID, &category.Name, &category.Description); err != nil {
			c.Status(500)
			return
		}
		categories = append(categories, category)
	}
	c.JSON(200, categories)
}

func AddNewCategory(c *gin.Context) {
	var request models.CategoryCreate
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Status(400)
		return
	}
	res, err := database.MysqlInstance.Exec("INSERT INTO categories (name, description) VALUES (?, ?)", request.Name, request.Description)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			c.JSON(409, gin.H{"error": "Category name already exists"})
			return
		}
		c.Status(500)
		return
	}
	id, err := res.LastInsertId()
	if err != nil {
		c.Status(500)
		return
	}
	c.JSON(201, gin.H{"id": id})
}

func DeleteCategory(c *gin.Context) {
	var request models.APICommonQueryID
	if err := c.ShouldBindQuery(&request); err != nil {
		c.Status(400)
		return
	}
	_, err := database.MysqlInstance.Exec("DELETE FROM categories WHERE id = ?", request.ID)
	if err != nil {
		if strings.Contains(err.Error(), "foreign key constraint fails") {
			c.JSON(409, gin.H{"error": "Category is being used"})
			return
		}
		c.Status(500)
		return
	}
	c.Status(200)
}
