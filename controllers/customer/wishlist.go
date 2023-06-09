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
	_, err = database.MysqlInstance.Exec(
		"INSERT INTO wishlists (product_refer, customer_refer) VALUES (UUID_TO_BIN(?), UUID_TO_BIN(?))", request.ID,
		customerId,
	)
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
	res, err := database.MysqlInstance.Exec(
		"DELETE FROM wishlists WHERE product_refer = UUID_TO_BIN(?) AND customer_refer = UUID_TO_BIN(?)", request.ID,
		customerId,
	)
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
			QueryRow(
				"SELECT 1 FROM wishlists WHERE product_refer = UUID_TO_BIN(?) AND customer_refer = UUID_TO_BIN(?)",
				request.ID, customerId,
			).
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
		Query(
			`
			SELECT
			    BIN_TO_UUID(p.id) AS id,
			    p.name,
			    p.price,
			    COALESCE(CONCAT(BIN_TO_UUID(pi.id), '.webp'), '') AS image,
			    p.cumulative_review,
			    COUNT(oi.id) AS sold_count
			FROM
			    wishlists w
			    left join products p on p.id = w.product_refer
			        LEFT JOIN (
			        SELECT
			            id,
			            product_refer,
			            ROW_NUMBER() OVER (PARTITION BY product_refer ORDER BY created_at) AS rn
			        FROM
			            product_images
			    ) pi ON p.id = pi.product_refer AND pi.rn = 1
			        LEFT JOIN
			    order_items oi ON oi.product_refer = p.id
			        LEFT JOIN
			    orders o ON oi.order_refer = o.id
			        AND (o.transaction_status = 'settlement'
			            OR o.transaction_status = 'capture')
			WHERE
			    p.deleted_at IS NULL
			AND w.customer_refer = UUID_TO_BIN(?)
			GROUP BY
			    p.id,
			    image;
			`, customerId,
		)
	if err != nil {
		c.Status(500)
		return
	}
	var response []models.WishlistItemResponse
	for rows.Next() {
		var item models.WishlistItemResponse
		err := rows.Scan(&item.ID, &item.Name, &item.Price, &item.ImageUrl, &item.Rating, &item.Sold)
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
