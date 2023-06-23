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
	"fmt"
	"strings"
	"sync"

	"github.com/Tus1688/openmerce-backend/auth"
	"github.com/Tus1688/openmerce-backend/database"
	"github.com/Tus1688/openmerce-backend/models"
	"github.com/Tus1688/openmerce-backend/service/freight"
	"github.com/gin-gonic/gin"
)

// PreCheckoutFreight is the handler for getting freight choice before the checkout process
func PreCheckoutFreight(c *gin.Context) {
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
	req := freight.PrecalculateFreightRequest{}
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}
	errChan := make(chan error, 2)
	wg.Add(2)

	//	get the shipping area id from the address
	go func() {
		defer wg.Done()
		var id uint32
		err = database.MysqlInstance.
			QueryRow(
				"SELECT shipping_area_refer FROM customer_addresses WHERE id = UUID_TO_BIN(?) AND customer_refer = UUID_TO_BIN(?)",
				request.ID, customerId,
			).
			Scan(&id)
		if err != nil {
			if err == sql.ErrNoRows {
				errChan <- fmt.Errorf("address not found")
				return
			}
			errChan <- err
			return
		}
		mu.Lock()
		req.ID = id
		mu.Unlock()
	}()

	//	count all volume and weight of the items in the cart
	go func() {
		defer wg.Done()
		var weight, volume float64
		err = database.MysqlInstance.
			QueryRow(
				`
				select sum(p.weight * c.quantity) as weight, sum((p.length * p.height * p.width) * c.quantity) as volume
				from products p, cart_items c
				left join inventories i on i.product_refer = c.product_refer
				where p.id = c.product_refer and c.checked = 1 and c.customer_refer = uuid_to_bin(?) and
				      c.quantity <= i.quantity and p.deleted_at is null
				group by c.customer_refer;
				`, customerId,
			).Scan(&weight, &volume)
		if err != nil {
			if err == sql.ErrNoRows {
				errChan <- fmt.Errorf("you haven't selected any item in the cart")
				return
			}
			errChan <- err
			return
		}
		mu.Lock()
		req.Weight = weight
		req.Volume = volume
		mu.Unlock()
	}()

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {
			if strings.Contains(err.Error(), "address not found") {
				c.JSON(404, gin.H{"error": err.Error()})
				return
			} else if strings.Contains(err.Error(), "you haven't selected any item in the cart") {
				c.JSON(409, gin.H{"error": err.Error()})
				return
			} else {
				c.Status(500)
				return
			}
		}
	}

	res, err := req.CalculateFreight()
	if err != nil {
		if strings.Contains(err.Error(), "there are no rates available for this route") {
			c.JSON(404, gin.H{"error": err.Error()})
			return
		}
		c.Status(500)
		return
	}
	var response []models.PreCheckoutFreight
	for _, value := range res.Anteraja {
		response = append(
			response, models.PreCheckoutFreight{
				ProductCode: "anteraja-" + value.ProductCode,
				CourierName: "anteraja",
				ProductName: value.ProductName,
				Etd:         value.Etd,
				Rates:       value.Rates,
			},
		)
	}
	for _, value := range res.Sicepat {
		response = append(
			response, models.PreCheckoutFreight{
				ProductCode: "sicepat-" + value.ProductCode,
				CourierName: "sicepat",
				ProductName: value.ProductName,
				Etd:         value.Etd,
				Rates:       value.Rates,
			},
		)
	}
	c.JSON(200, response)
}

// PreCheckoutItems is the handler for getting all ticked items in the cart before the checkout process
func PreCheckoutItems(c *gin.Context) {
	// the token should be valid and exist as it is protected by TokenExpiredCustomer middleware
	token, _ := c.Cookie("ac_cus")
	claims, err := auth.ExtractClaimAccessTokenCustomer(token)
	if err != nil {
		c.Status(401)
		return
	}
	customerId := claims.Uid
	var response []models.PreCheckoutItem
	rows, err := database.MysqlInstance.
		Query(
			`
			SELECT
			    BIN_TO_UUID(p.id) AS id,
			    p.name,
			    p.price,
			    COALESCE(CONCAT(BIN_TO_UUID(pi.id), '.webp'), '') AS image,
			    c.quantity
			FROM cart_items c
			        left join products p on p.id = c.product_refer
			        LEFT JOIN (
			        SELECT
			            id,
			            product_refer,
			            ROW_NUMBER() OVER (PARTITION BY product_refer ORDER BY created_at) AS rn
			        FROM
			            product_images
			    ) pi ON p.id = pi.product_refer AND pi.rn = 1
				LEFT JOIN inventories i ON i.product_refer = c.product_refer
			WHERE
			    p.deleted_at IS NULL
			  AND c.customer_refer = UUID_TO_BIN(?) AND c.checked = 1
						AND p.deleted_at IS NULL AND c.quantity <= i.quantity
		`, customerId,
		)
	if err != nil {
		c.Status(500)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item models.PreCheckoutItem
		if err := rows.Scan(
			&item.ProductId, &item.ProductName, &item.ProductPrice, &item.ProductImage, &item.Quantity,
		); err != nil {
			c.Status(500)
			return
		}
		response = append(response, item)
	}
	if err := rows.Err(); err != nil {
		c.Status(500)
		return
	}
	if len(response) == 0 {
		c.JSON(409, gin.H{"error": "you haven't selected any item in the cart"})
		return
	}
	c.JSON(200, response)
}
