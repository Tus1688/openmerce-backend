package customer

import (
	"database/sql"
	"strings"

	"github.com/Tus1688/openmerce-backend/auth"
	"github.com/Tus1688/openmerce-backend/database"
	"github.com/Tus1688/openmerce-backend/models"
	"github.com/gin-gonic/gin"
)

func AddToWishlist(c *gin.Context) {
	var request models.APICommonQueryUUID
	if err := c.ShouldBindQuery(&request); err != nil {
		c.Status(400)
		return
	}
	// the token should be valid and exist as it is protected by TokenExpiredCustomer middleware
	token, _ := c.Cookie("ac_cus")
	claims, err := auth.ExtractClaimAccessTokenCustomer(token)
	if err != nil {
		c.Status(401)
		return
	}
	customerId := claims.Uid
	// check if the product exist
	var exists uint8
	err = database.MysqlInstance.
		QueryRow("SELECT 1 FROM products WHERE id = UUID_TO_BIN(?) AND deleted_at IS NULL", request.ID).
		Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			c.Status(404)
			return
		}
		c.Status(500)
		return
	}
	_, err = database.MysqlInstance.Exec("INSERT INTO wishlists (product_refer, customer_refer) VALUES (UUID_TO_BIN(?), UUID_TO_BIN(?))", request.ID, customerId)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			c.JSON(409, gin.H{"error": "product already in wishlist"})
			return
		}
		c.Status(500)
		return
	}
	c.Status(200)
}

func DeleteWishlist(c *gin.Context) {
	var request models.APICommonQueryUUID
	if err := c.ShouldBindQuery(&request); err != nil {
		c.Status(400)
		return
	}
	// the token should be valid and exist as it is protected by TokenExpiredCustomer middleware
	token, _ := c.Cookie("ac_cus")
	claims, err := auth.ExtractClaimAccessTokenCustomer(token)
	if err != nil {
		c.Status(401)
		return
	}
	customerId := claims.Uid
	res, err := database.MysqlInstance.Exec("DELETE FROM wishlists WHERE product_refer = UUID_TO_BIN(?) AND customer_refer = UUID_TO_BIN(?)", request.ID, customerId)
	if err != nil {
		c.Status(500)
		return
	}
	affected, err := res.RowsAffected()
	if err != nil {
		c.Status(500)
		return
	}
	if affected == 0 {
		c.JSON(404, gin.H{"error": "product not found in the wishlist"})
		return
	}
	c.Status(200)
}

func GetWishlist(c *gin.Context) {
	// the token should be valid and exist as it is protected by TokenExpiredCustomer middleware
	token, _ := c.Cookie("ac_cus")
	claims, err := auth.ExtractClaimAccessTokenCustomer(token)
	if err != nil {
		c.Status(401)
		return
	}
	customerId := claims.Uid

	var request models.APICommonQueryUUID
	if err := c.ShouldBindQuery(&request); err == nil {
		var exists uint8
		err = database.MysqlInstance.
			QueryRow("SELECT 1 FROM wishlists WHERE product_refer = UUID_TO_BIN(?) AND customer_refer = UUID_TO_BIN(?)", request.ID, customerId).
			Scan(&exists)
		if err != nil && err != sql.ErrNoRows {
			c.Status(500)
			return
		}
		if exists != 0 {
			c.JSON(200, gin.H{"state": true})
			return
		}
		c.JSON(200, gin.H{"state": false})
	}
	if request.ID != "" {
		c.Status(400)
		return
	}
	//	return all wishlist if there is no query
	rows, err := database.MysqlInstance.
		Query(`
				SELECT BIN_TO_UUID(p.id) as id, p.name, p.price, COALESCE(CONCAT(BIN_TO_UUID(pi.id), '.webp'), '') as image
				FROM wishlists w
				         JOIN products p ON w.product_refer = p.id
				         LEFT JOIN (SELECT product_refer, MIN(created_at) AS min_created_at
				                    FROM product_images
				                    GROUP BY product_refer) pi_min ON p.id = pi_min.product_refer
				         LEFT JOIN product_images pi ON pi.product_refer = p.id AND pi.created_at = pi_min.min_created_at
				WHERE w.customer_refer = UUID_TO_BIN(?)
				  AND p.deleted_at IS NULL;
			`, customerId)
	if err != nil {
		c.Status(500)
		return
	}
	var response []models.WishlistItemResponse
	for rows.Next() {
		var item models.WishlistItemResponse
		err := rows.Scan(&item.ProductId, &item.ProductName, &item.ProductPrice, &item.ProductImage)
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
