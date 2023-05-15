package customer

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

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
		Query(`SELECT BIN_TO_UUID(p.id) as id, p.name, p.price, COALESCE(CONCAT(BIN_TO_UUID(pi.id), '.webp'), '') as image, c.quantity, i.quantity
			FROM cart_items c
			JOIN products p ON c.product_refer = p.id
			JOIN inventories i ON p.id = i.product_refer
			LEFT JOIN (SELECT product_refer, MIN(created_at) AS min_created_at
			FROM product_images
			GROUP BY product_refer) pi_min ON p.id = pi_min.product_refer
			LEFT JOIN product_images pi ON pi.product_refer = p.id AND pi.created_at = pi_min.min_created_at
			WHERE c.customer_refer = UUID_TO_BIN(?)
			AND p.deleted_at IS NULL;
			`, customerId)
	if err != nil {
		c.Status(500)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item models.CartItemResponse
		err := rows.Scan(&item.ProductId, &item.ProductName, &item.ProductPrice, &item.ProductImage, &item.Quantity, &item.CurrentStock)
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

// AddToCart adds a product to cart and may update the quantity if the product already exists in cart
func AddToCart(c *gin.Context) {
	var request models.CartInsert
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Status(400)
		return
	}
	wg := sync.WaitGroup{}
	errChan := make(chan error, 1)
	stockChan := make(chan uint16, 1)
	wg.Add(1)
	// this goroutine check if the product exists and the quantity is enough
	go func(productId string) {
		defer wg.Done()
		var quantity uint16
		err := database.MysqlInstance.
			QueryRow("SELECT i.quantity FROM inventories i, products p WHERE p.id = UUID_TO_BIN(?) AND i.product_refer = p.id AND p.deleted_at IS NULL", productId).
			Scan(&quantity)
		if err != nil {
			errChan <- err
			return
		}
		if quantity < request.Quantity {
			errChan <- fmt.Errorf("quantity is not enough")
			stockChan <- quantity
			return
		}
	}(request.ProductId)

	// the token should be valid and exist as it is protected by TokenExpiredCustomer middleware
	token, _ := c.Cookie("ac_cus")
	claims, err := auth.ExtractClaimAccessTokenCustomer(token)
	if err != nil {
		c.Status(401)
		return
	}
	customerId := claims.Uid

	wg.Wait()
	close(errChan)
	close(stockChan)
	if err := <-errChan; err != nil {
		if err.Error() == sql.ErrNoRows.Error() {
			c.JSON(404, gin.H{"error": "product not found"})
			return
		}
		if strings.Contains(err.Error(), "quantity") {
			stock := <-stockChan
			c.JSON(409, gin.H{
				"error":      err.Error(),
				"curr_stock": stock,
			})
			return
		}
		c.Status(500)
		return
	}

	_, err = database.MysqlInstance.Exec(`
		INSERT INTO cart_items (product_refer, customer_refer, quantity) VALUES
		(UUID_TO_BIN(?), UUID_TO_BIN(?), ?)
		ON DUPLICATE KEY UPDATE quantity = ?
	`, request.ProductId, customerId, request.Quantity, request.Quantity)
	if err != nil {
		c.Status(500)
		return
	}
	c.Status(200)
}

func DeleteCart(c *gin.Context) {
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

	res, err := database.MysqlInstance.Exec(`
		DELETE FROM cart_items WHERE customer_refer = UUID_TO_BIN(?) AND product_refer = UUID_TO_BIN(?)
	`, customerId, request.ID)
	if err != nil {
		c.Status(500)
		return
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		c.Status(404)
		return
	}
	c.Status(200)
}
