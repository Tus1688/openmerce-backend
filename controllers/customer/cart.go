package customer

import (
	"github.com/Tus1688/openmerce-backend/auth"
	"github.com/Tus1688/openmerce-backend/database"
	"github.com/Tus1688/openmerce-backend/models"
	"github.com/gin-gonic/gin"
)

func GetCart(c *gin.Context) {
	// the token should be valid and exist as it is protected by TokenExpiredCustomer middleware
	token, _ := c.Cookie("ac_cus")
	claims, err := auth.ExtractClaimAccessTokenCustomer(token)
	if err != nil {
		c.Status(401)
		return
	}

	customerId := claims.Uid
	var response []models.CartItemResponse
	rows, err := database.MysqlInstance.
		Query("SELECT BIN_TO_UUID(p.id), p.name, p.price, c.quantity FROM cart_items c, products p WHERE c.customer_refer = UUID_TO_BIN(?) AND c.product_refer = p.id",
			customerId)
	if err != nil {
		c.Status(500)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item models.CartItemResponse
		err := rows.Scan(&item.ProductId, &item.ProductName, &item.ProductPrice, &item.Quantity)
		if err != nil {
			c.Status(500)
			return
		}
		response = append(response, item)
	}
	if len(response) == 0 {
		c.Status(404)
		return
	}
	c.JSON(200, response)
}
