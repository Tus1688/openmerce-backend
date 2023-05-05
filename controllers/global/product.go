package global

import (
	"sync"

	"github.com/Tus1688/openmerce-backend/database"
	"github.com/Tus1688/openmerce-backend/models"
	"github.com/gin-gonic/gin"
)

func GetProduct(c *gin.Context) {
	var requestID models.APICommonQueryUUID
	var requestSearch models.APICommonQuerySearch

	if err := c.ShouldBindQuery(&requestID); err == nil {
		return
	}

	if err := c.ShouldBindQuery(&requestSearch); err == nil {
		return
	}
	// return everything if no query is provided
	var response []models.HomepageProductResponse

	// get categories which should be included in the home page
	var categories []uint
	rows, err := database.MysqlInstance.Query("SELECT id FROM categories WHERE homepage_visibility = 1")
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
				Query(`SELECT BIN_TO_UUID(p.id) as id, p.name, p.price, CONCAT(BIN_TO_UUID(pi.id), '.webp') as image, p.cumulative_review
					FROM product_images pi
					INNER JOIN (
					  SELECT product_refer, MIN(created_at) AS min_created_at
					  FROM product_images
					  GROUP BY product_refer
					) pi_min ON pi.product_refer = pi_min.product_refer AND pi.created_at = pi_min.min_created_at
					INNER JOIN products p on pi.product_refer = p.id and p.deleted_at is null and p.category_refer = ?`, category)
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
			response = append(response, chunk)
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
