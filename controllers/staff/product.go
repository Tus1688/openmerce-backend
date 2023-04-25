package staff

import (
	"strings"

	"github.com/Tus1688/openmerce-backend/database"
	"github.com/Tus1688/openmerce-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AddNewProduct(c *gin.Context) {
	var request models.ProductCreate
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Status(400)
		return
	}
	id := uuid.New()
	// insert the new product
	_, err := database.MysqlInstance.
		Exec("INSERT INTO products (id, name, description, price, weight) VALUES (UUID_TO_BIN(?), ?, ?, ?, ?)",
			id, request.Name, request.Description, request.Price, request.Weight)
	if err != nil {
		//	check if the product name already exists
		if strings.Contains(err.Error(), "Duplicate entry") {
			c.JSON(409, gin.H{"error": "Product name already exists"})
			return
		}
		c.Status(500)
		return
	}
	// insert the new product into inventories
	_, err = database.MysqlInstance.
		Exec("INSERT INTO inventories (product_refer, quantity, updated_at) VALUE (UUID_TO_BIN(?), ?, CURRENT_TIMESTAMP)",
			id, request.InitialStock)
	if err != nil {
		c.Status(500)
		return
	}
	c.JSON(201, gin.H{"id": id})
}
