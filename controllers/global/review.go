package global

import (
	"github.com/Tus1688/openmerce-backend/database"
	"github.com/Tus1688/openmerce-backend/models"
	"github.com/gin-gonic/gin"
)

// GetReviewGlobal is the handler for getting review for specific product
func GetReviewGlobal(c *gin.Context) {
	var request models.APICommonQueryUUID
	if err := c.ShouldBindQuery(&request); err != nil {
		c.Status(400)
		return
	}
	// get the review
	var reviews []models.ReviewResponseGlobal
	rows, err := database.MysqlInstance.
		Query(`
			SELECT BIN_TO_UUID(r.id),
			       r.rating,
			       COALESCE(r.review, ''),
			       DATE_FORMAT(o.created_at, '%d %M %Y'),
			       CONCAT(c.first_name, ' ', c.last_name)
			FROM reviews r
			         LEFT JOIN order_items oi on r.order_item_refer = oi.id
			         LEFT JOIN orders o on oi.order_refer = o.id
			         LEFT JOIN customers c on o.customer_refer = c.id
			WHERE r.product_refer = UUID_TO_BIN(?)
		`, request.ID)
	if err != nil {
		c.Status(500)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var review models.ReviewResponseGlobal
		err = rows.Scan(&review.ID, &review.Rating, &review.Review, &review.CreatedAt, &review.Customer)
		if err != nil {
			c.Status(500)
			return
		}
		reviews = append(reviews, review)
	}
	if len(reviews) == 0 {
		c.Status(404)
		return
	}
	c.JSON(200, reviews)
}
