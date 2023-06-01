package customer

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Tus1688/openmerce-backend/auth"
	"github.com/Tus1688/openmerce-backend/database"
	"github.com/Tus1688/openmerce-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func GetCartCount(c *gin.Context) {
	// the token should be valid and exist as it is protected by TokenExpiredCustomer middleware
	token, _ := c.Cookie("ac_cus")
	claims, err := auth.ExtractClaimAccessTokenCustomer(token)
	if err != nil {
		c.Status(401)
		return
	}
	//	check from redis[3] if the cart count is cached
	customerId := claims.Uid
	var count uint8
	err = database.RedisInstance[3].Get(context.Background(), customerId).Scan(&count)
	//	if there is no cache, get the count from database
	if err == redis.Nil {
		err := database.MysqlInstance.QueryRow("SELECT COUNT(customer_refer) FROM cart_items WHERE customer_refer = UUID_TO_BIN(?)", customerId).Scan(&count)
		if err != nil {
			// there won't be sql.ErrNoRows as it will return 0
			c.Status(500)
			return
		}
		// update the redis cache
		go func(curValue uint8, userID string) {
			_ = database.RedisInstance[3].Set(context.Background(), userID, curValue, 24*14*time.Hour).Err()
		}(count, customerId)
	}
	c.JSON(200, gin.H{"count": count})
}

func CheckCartItem(c *gin.Context) {
	var request models.CartCheck
	if err := c.ShouldBindJSON(&request); err != nil {
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
	res, err := database.MysqlInstance.
		Exec(`
			UPDATE cart_items c
			LEFT JOIN inventories i on c.product_refer = i.product_refer
			LEFT JOIN products p on c.product_refer = p.id
			SET c.checked = ? WHERE c.customer_refer = UUID_TO_BIN(?) AND c.product_refer = UUID_TO_BIN(?) 
			AND i.quantity >= c.quantity AND p.deleted_at IS NULL
			`,
			request.State, customerId, request.ProductID)
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
		// product not found in cart or there is no change in the state
		c.Status(404)
		return
	}
	// we don't need to update the redis cache as it doesn't affect the cart count
	c.Status(200)
}

func CheckAllCartItem(c *gin.Context) {
	var request models.CheckAll
	if err := c.ShouldBindJSON(&request); err != nil {
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
	res, err := database.MysqlInstance.
		Exec(`
		UPDATE cart_items 
		LEFT JOIN inventories i on cart_items.product_refer = i.product_refer
		LEFT JOIN products p on cart_items.product_refer = p.id
		SET checked = ? WHERE customer_refer = UUID_TO_BIN(?) AND i.quantity >= cart_items.quantity AND p.deleted_at IS NULL
		`, request.State, customerId)
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
		c.Status(409)
		return
	}
	c.Status(200)
}

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
		Query(`SELECT BIN_TO_UUID(p.id) as id, p.name, p.price, COALESCE(CONCAT(BIN_TO_UUID(pi.id), '.webp'), '') as image, c.quantity, i.quantity, c.checked
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
		err := rows.Scan(&item.ProductId, &item.ProductName, &item.ProductPrice, &item.ProductImage, &item.Quantity, &item.CurrentStock, &item.Checked)
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
	go updateCartCache(customerId)
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
	go updateCartCache(customerId)
	c.Status(200)
}

func updateCartCache(customerID string) {
	var count uint16
	err := database.MysqlInstance.
		QueryRow("SELECT COUNT(customer_refer) FROM cart_items WHERE customer_refer = UUID_TO_BIN(?)", customerID).
		Scan(&count)
	if err == nil {
		_ = database.RedisInstance[3].Set(context.Background(), customerID, count, 24*14*time.Hour).Err()
	}
}
