package customer

import (
	"github.com/Tus1688/openmerce-backend/auth"
	"github.com/Tus1688/openmerce-backend/database"
	"github.com/Tus1688/openmerce-backend/models"
	"github.com/gin-gonic/gin"
)

func GetOrder(c *gin.Context) {
	// the token should be valid and exist as it is protected by TokenExpiredCustomer middleware
	token, _ := c.Cookie("ac_cus")
	claims, err := auth.ExtractClaimAccessTokenCustomer(token)
	if err != nil {
		c.Status(401)
		return
	}
	customerId := claims.Uid
	var response []models.OrderResponse
	// get the order list from database
	rows, err := database.MysqlInstance.
		Query(`
			SELECT o.id,
			       DATE_FORMAT(o.created_at, '%d %M %Y') AS created_at,
			       o.gross_amount,
			       o.transaction_status,
			       oi.item_count,
			       COALESCE(pi.image, ''),
			       p.name
			FROM orders o
			         LEFT JOIN (
			    SELECT order_refer, COUNT(*) AS item_count
			    FROM order_items
			    GROUP BY order_refer
			) oi ON oi.order_refer = o.id
			         LEFT JOIN (
			    SELECT pi.product_refer, CONCAT(BIN_TO_UUID(pi.id), '.webp') AS image
			    FROM (
			             SELECT product_refer, MIN(id) AS id
			             FROM product_images
			             GROUP BY product_refer
			         ) pi
			) pi ON pi.product_refer = (SELECT product_refer FROM order_items WHERE order_refer = o.id LIMIT 1)
			         LEFT JOIN (
			    SELECT id, name
			    FROM products
			) p ON p.id = (SELECT product_refer FROM order_items WHERE order_refer = o.id LIMIT 1)
			WHERE o.customer_refer = UUID_TO_BIN(?);
			`, customerId)
	if err != nil {
		c.Status(500)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item models.OrderResponse
		err := rows.Scan(&item.ID, &item.CreatedAt, &item.GrossAmount, &item.Status, &item.ItemCount, &item.Image, &item.ProductName)
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
