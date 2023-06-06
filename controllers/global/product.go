package global

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/Tus1688/openmerce-backend/database"
	"github.com/Tus1688/openmerce-backend/logging"
	"github.com/Tus1688/openmerce-backend/models"
	"github.com/gin-gonic/gin"
)

func GetProductSold(c *gin.Context) {
	var request models.APICommonQueryUUID
	if err := c.ShouldBindQuery(&request); err != nil {
		c.Status(400)
		return
	}
	// check from redis[6] if this product is cached
	var count uint32
	err := database.RedisInstance[6].Get(context.Background(), request.ID).Scan(&count)
	// if there is no cache, get the count from database
	if err != nil {
		err := database.MysqlInstance.
			QueryRow("SELECT SUM(quantity) FROM order_items oi LEFT JOIN orders o on oi.order_refer = o.id WHERE oi.product_refer = UUID_TO_BIN(?) AND o.is_paid = 1 AND o.transaction_status = 'settlement' OR o.transaction_status = 'capture' GROUP BY product_refer", request.ID).
			Scan(&count)
		if err != nil {
			if err == sql.ErrNoRows {
				c.Status(404)
				return
			}
			c.Status(500)
			return
		}
		// update the redis cache
		go func(curValue uint32, productID string) {
			err = database.RedisInstance[6].Set(context.Background(), productID, curValue, 24*14*time.Hour).Err()
			if err != nil {
				logging.InsertLog(logging.ERROR, "unable to update redis cache for product sold count")
			}
		}(count, request.ID)
	}
	c.JSON(200, gin.H{"count": count})
}

func GetProduct(c *gin.Context) {
	var requestID models.APICommonQueryUUID
	var requestSearch models.APICommonQuerySearch

	if err := c.ShouldBindQuery(&requestID); err == nil {
		var response models.ProductDetail

		err := database.MysqlInstance.QueryRow(`
			SELECT BIN_TO_UUID(p.id), p.name, p.description, p.price, p.weight, c.name, p.cumulative_review, CONCAT(p.length, ' x ', p.width, ' x ', p.height), i.quantity FROM products p, categories c, inventories i
			WHERE p.category_refer = c.id AND p.deleted_at IS NULL AND p.id = UUID_TO_BIN(?) AND p.id = i.product_refer`, requestID.ID).
			Scan(&response.ID, &response.Name, &response.Description, &response.Price, &response.Weight, &response.CategoryName, &response.CumulativeReview, &response.Dimension, &response.Stock)
		if err != nil {
			if err == sql.ErrNoRows {
				c.Status(404)
				return
			}
			c.Status(500)
			return
		}
		rows, err := database.MysqlInstance.Query(`
		SELECT CONCAT(BIN_TO_UUID(id), '.webp') FROM product_images WHERE product_refer = UUID_TO_BIN(?)`, requestID.ID)
		if err != nil {
			c.Status(500)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var image string
			err := rows.Scan(&image)
			if err != nil {
				c.Status(500)
				return
			}
			response.ImageUrls = append(response.ImageUrls, image)
		}
		c.JSON(200, response)
		return
	}

	if err := c.ShouldBindQuery(&requestSearch); err == nil {
		sqlQuery :=
			`SELECT BIN_TO_UUID(p.id) as id, p.name, p.price, COALESCE(CONCAT(BIN_TO_UUID(pi.id), '.webp'), '') as image, p.cumulative_review
			 FROM products p
			 LEFT JOIN (
			 	SELECT product_refer, MIN(created_at) AS min_created_at
			 FROM product_images
			 GROUP BY product_refer
			 ) pi_min ON p.id = pi_min.product_refer
			 LEFT JOIN product_images pi ON pi.product_refer = p.id AND pi.created_at = pi_min.min_created_at
			 WHERE p.deleted_at IS NULL AND MATCH(p.name) AGAINST(? IN BOOLEAN MODE)`
		args := []interface{}{requestSearch.Search + "*"}
		category := c.Query("category")
		if category != "" {
			sqlQuery += " AND category_refer = ?"
			args = append(args, category)
		}
		priceFrom := c.Query("price_from")
		if priceFrom != "" {
			sqlQuery += " AND price >= ?"
			args = append(args, priceFrom)
		}
		priceTo := c.Query("price_to")
		if priceTo != "" {
			sqlQuery += " AND price <= ?"
			args = append(args, priceTo)
		}
		limit := c.Query("limit")
		if limit != "" {
			sqlQuery += " LIMIT ?"
			args = append(args, limit)
		}
		var response []models.HomepageProduct
		rows, err := database.MysqlInstance.Query(sqlQuery, args...)
		if err != nil {
			c.Status(500)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var product models.HomepageProduct
			if err := rows.Scan(&product.ID, &product.Name, &product.Price, &product.ImageUrl, &product.Rating); err != nil {
				c.Status(500)
				return
			}
			response = append(response, product)
		}
		if response == nil {
			c.Status(404)
			return
		}
		c.JSON(200, response)
		return
	}
	// check if there is no query for id because it binds to uuid and will return 400 if the id is not a valid uuid
	if requestID.ID != "" {
		c.Status(400)
		return
	}
	// return everything if no query is provided
	var response []models.HomepageProductResponse

	// get categories which should be included in the home page
	var categories []uint
	rows, err := database.MysqlInstance.
		Query("SELECT id FROM categories WHERE homepage_visibility = 1 AND deleted_at IS NULL LIMIT 5")
	if err != nil {
		c.Status(500)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var category uint
		if err := rows.Scan(&category); err != nil {
			c.Status(500)
			return
		}
		categories = append(categories, category)
	}

	var wg sync.WaitGroup
	errChan := make(chan error)
	mu := sync.Mutex{}
	for _, category := range categories {
		wg.Add(1)
		go func(category uint) {
			defer wg.Done()
			var chunk models.HomepageProductResponse
			//	get category detail
			err := database.MysqlInstance.QueryRow("SELECT id, name, description FROM categories WHERE id =?", category).
				Scan(&chunk.CategoryID, &chunk.CategoryName, &chunk.CategoryDesc)
			if err != nil {
				errChan <- err
				return
			}
			//  get products in the category
			rows, err := database.MysqlInstance.
				Query(`SELECT BIN_TO_UUID(p.id) AS id, p.name, p.price, COALESCE(CONCAT(BIN_TO_UUID(pi.id), '.webp'), '') AS image, p.cumulative_review,
					       COUNT(oi.id) AS sold_count
					FROM products p
					LEFT JOIN (
					  SELECT product_refer, MIN(created_at) AS min_created_at
					  FROM product_images
					  GROUP BY product_refer
					) pi_min ON p.id = pi_min.product_refer
					LEFT JOIN product_images pi ON pi.product_refer = p.id AND pi.created_at = pi_min.min_created_at
					LEFT JOIN order_items oi ON oi.product_refer = p.id
					LEFT JOIN orders o ON oi.order_refer = o.id AND (o.transaction_status = 'settlement' OR o.transaction_status = 'capture')
					WHERE p.deleted_at IS NULL AND p.category_refer = ?
					GROUP BY p.id, p.name, p.price, image, p.cumulative_review
					LIMIT 12;`, category)
			if err != nil {
				errChan <- err
				return
			}
			defer rows.Close()
			for rows.Next() {
				var product models.HomepageProduct
				if err := rows.Scan(&product.ID, &product.Name, &product.Price, &product.ImageUrl, &product.Rating, &product.Sold); err != nil {
					errChan <- err
					return
				}
				chunk.Products = append(chunk.Products, product)
			}
			mu.Lock()
			response = append(response, chunk)
			mu.Unlock()
		}(category)
	}
	go func() {
		wg.Wait()
		close(errChan)
	}()
	for err := range errChan {
		if err != nil {
			c.Status(500)
			return
		}
	}
	c.JSON(200, response)
}
