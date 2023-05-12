package global

import (
	"database/sql"
	"sync"

	"github.com/Tus1688/openmerce-backend/database"
	"github.com/Tus1688/openmerce-backend/models"
	"github.com/gin-gonic/gin"
)

func GetProduct(c *gin.Context) {
	var requestID models.APICommonQueryUUID
	var requestSearch models.APICommonQuerySearch

	if err := c.ShouldBindQuery(&requestID); err == nil {
		var response models.ProductDetail

		err := database.MysqlInstance.QueryRow(`
			SELECT BIN_TO_UUID(p.id), p.name, p.description, p.price, p.weight, c.name, p.cumulative_review FROM products p, categories c
			WHERE p.category_refer = c.id AND p.deleted_at IS NULL AND p.id = UUID_TO_BIN(?)`, requestID.ID).
			Scan(&response.ID, &response.Name, &response.Description, &response.Price, &response.Weight, &response.CategoryName, &response.CumulativeReview)
		if err != nil {
			if err == sql.ErrNoRows {
				c.Status(404)
				return
			}
			// might be triggered too if the id is not a valid uuid
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
	rows, err := database.MysqlInstance.Query("SELECT id FROM categories WHERE homepage_visibility = 1 AND deleted_at IS NULL")
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
				Query(`SELECT BIN_TO_UUID(p.id) as id, p.name, p.price, COALESCE(CONCAT(BIN_TO_UUID(pi.id), '.webp'), '') as image, p.cumulative_review
							 FROM products p
							 LEFT JOIN (
							   SELECT product_refer, MIN(created_at) AS min_created_at
							   FROM product_images
							   GROUP BY product_refer
							 ) pi_min ON p.id = pi_min.product_refer
							 LEFT JOIN product_images pi ON pi.product_refer = p.id AND pi.created_at = pi_min.min_created_at
							 WHERE p.deleted_at IS NULL and p.category_refer = ?`, category)
			if err != nil {
				errChan <- err
				return
			}
			defer rows.Close()
			for rows.Next() {
				var product models.HomepageProduct
				if err := rows.Scan(&product.ID, &product.Name, &product.Price, &product.ImageUrl, &product.Rating); err != nil {
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
