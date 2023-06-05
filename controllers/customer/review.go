package customer

import (
	"database/sql"
	"strings"

	"github.com/Tus1688/openmerce-backend/auth"
	"github.com/Tus1688/openmerce-backend/database"
	"github.com/Tus1688/openmerce-backend/models"
	"github.com/gin-gonic/gin"
)

func CreateReview(c *gin.Context) {
	var request models.CreateReview
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Status(400)
		return
	}
	// check if the request.rating is in the range of 1-5
	if request.Rating < 1 || request.Rating > 5 {
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
	// check if the order id belongs to the customer
	var exist int8
	err = database.MysqlInstance.
		QueryRow("SELECT 1 FROM order_items oi LEFT JOIN orders o on oi.order_refer = o.id WHERE o.customer_refer = UUID_TO_BIN(?) AND oi.id = ?",
			customerId, request.OrderID).Scan(&exist)
	if err != nil {
		if err == sql.ErrNoRows {
			c.Status(404)
			return
		}
		c.Status(500)
		return
	}
	// insert the review
	_, err = database.MysqlInstance.
		Exec(`INSERT INTO reviews (order_item_refer, product_refer, rating, review) 
		VALUES (?, (SELECT oi.product_refer FROM order_items oi WHERE oi.id = ?), ?, ?)`,
			request.OrderID, request.OrderID, request.Rating, request.Review)
	if err != nil {
		// check if the review already exists
		if strings.Contains(err.Error(), "Duplicate entry") {
			c.JSON(409, gin.H{"error": "you have already reviewed this item"})
			return
		}
		c.Status(500)
		return
	}
	c.Status(201)
}

// GetReview here is meant to get the review of every review of the customer
func GetReview(c *gin.Context) {
	// the token should be valid and exist as it is protected by TokenExpiredCustomer middleware
	token, _ := c.Cookie("ac_cus")
	claims, err := auth.ExtractClaimAccessTokenCustomer(token)
	if err != nil {
		c.Status(401)
		return
	}
	customerId := claims.Uid
	var response []models.ReviewResponseCustomer
	rows, err := database.MysqlInstance.
		Query(`
			SELECT BIN_TO_UUID(r.id),
			       BIN_TO_UUID(p.id),
			       p.name,
			       COALESCE(pi.image, ''),
			       r.rating,
			       COALESCE(r.review, ''),
			       date_format(o.created_at, '%d %M %Y')
			FROM reviews r
			         LEFT JOIN order_items oi on r.order_item_refer = oi.id
			         LEFT JOIN products p on r.product_refer = p.id
			         LEFT JOIN orders o on oi.order_refer = o.id
			         left join (select pi.product_refer, CONCAT(BIN_TO_UUID(pi.id), '.webp') as image
			                    from (select product_refer, min(id) as id
			                          from product_images
			                          group by product_refer) pi) pi on pi.product_refer = oi.product_refer
			WHERE o.customer_refer = UUID_TO_BIN(?);
			`, customerId)
	if err != nil {
		c.Status(500)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var review models.ReviewResponseCustomer
		err = rows.Scan(&review.ID, &review.ProductID, &review.ProductName, &review.ProductImage, &review.Rating,
			&review.Review, &review.CreatedAt)
		if err != nil {
			c.Status(500)
			return
		}
		response = append(response, review)
	}
	if len(response) == 0 {
		c.Status(404)
		return
	}
	c.JSON(200, response)
}
