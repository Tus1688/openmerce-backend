// Copyright (c) 2023. Tus1688
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
		Query(
			`
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
		`, request.ID,
		)
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
