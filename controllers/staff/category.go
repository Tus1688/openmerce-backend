package staff

import (
	"strings"

	"github.com/Tus1688/openmerce-backend/database"
	"github.com/Tus1688/openmerce-backend/models"
	"github.com/gin-gonic/gin"
)

func GetCategories(c *gin.Context) {
	var categories []models.CategoryResponse
	rows, err := database.MysqlInstance.Query("SELECT id, name, description, homepage_visibility FROM categories")
	if err != nil {
		c.Status(500)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var category models.CategoryResponse
		if err := rows.Scan(&category.ID, &category.Name, &category.Description, &category.HomePageVisibility); err != nil {
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
	res, err := database.MysqlInstance.Exec("INSERT INTO categories (name, description, homepage_visibility) VALUES (?, ?, ?)",
		request.Name, request.Description, request.HomePageVisibility)
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
	res, err := database.MysqlInstance.Exec("DELETE FROM categories WHERE id = ?", request.ID)
	if err != nil {
		if strings.Contains(err.Error(), "foreign key constraint fails") {
			c.JSON(409, gin.H{"error": "Category is being used"})
			return
		}
		c.Status(500)
		return
	}
	affected, err := res.RowsAffected()
	if err != nil {
		c.Status(500)
		return
	}
	if affected == 0 {
		c.Status(404)
		return
	}
	c.Status(200)
}

func UpdateCategory(c *gin.Context) {
	var request models.CategoryUpdate
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Status(400)
		return
	}
	query := "UPDATE categories set updated_at = CURRENT_TIMESTAMP"
	var args []interface{}
	if request.Name != "" {
		query += ", name = ?"
		args = append(args, request.Name)
	}
	if request.Description != "" {
		query += ", description = ?"
		args = append(args, request.Description)
	}
	if request.HomePageVisibility != nil {
		query += ", homepage_visibility = ?"
		args = append(args, request.HomePageVisibility)
	}
	query += " WHERE id = ?"
	args = append(args, request.ID)
	res, err := database.MysqlInstance.Exec(query, args...)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			c.JSON(409, gin.H{"error": "Category name already exists"})
			return
		}
		c.Status(500)
		return
	}
	affected, err := res.RowsAffected()
	if err != nil {
		c.Status(500)
		return
	}
	if affected == 0 {
		c.Status(404)
		return
	}
	c.Status(200)
}
