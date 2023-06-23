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
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Tus1688/openmerce-backend/database"
	"github.com/Tus1688/openmerce-backend/models"
	"github.com/Tus1688/openmerce-backend/service/freight"
	"github.com/gin-gonic/gin"
)

// GetSuggestArea suggest a list of areas based on the query
func GetSuggestArea(c *gin.Context) {
	var request models.APICommonQuerySearch
	if err := c.ShouldBindQuery(&request); err != nil {
		c.Status(400)
		return
	}
	// check from redis cache first if there is a match
	val, err := database.RedisInstance[4].Get(context.Background(), request.Search).Result()
	if err == nil {
		var res []models.AreaResponse
		if err := json.Unmarshal([]byte(val), &res); err == nil {
			// this endpoint is not going to change on a daily basis for the estimated price, so we can cache it for 1 day
			c.Header("Cache-Control", "public, max-age=86400, immutable")
			c.JSON(200, res)
			return
		}
		// if there is error in unmarshalling, we will rely on the mysql query
		// delete the key from redis
		go func() {
			if err := database.RedisInstance[4].Del(context.Background(), request.Search).Err(); err != nil {
				log.Print(err)
			}
		}()
	}
	rows, err := database.MysqlInstance.
		Query(
			"SELECT id, full_name FROM shipping_areas WHERE MATCH(full_name) AGAINST(? IN BOOLEAN MODE) LIMIT 5",
			request.Search+"*",
		)
	if err != nil {
		c.Status(500)
		return
	}
	defer rows.Close()
	var res []models.AreaResponse
	for rows.Next() {
		var area models.AreaResponse
		if err := rows.Scan(&area.ID, &area.Name); err != nil {
			c.Status(500)
			return
		}
		res = append(res, area)
	}
	if len(res) == 0 {
		c.Status(404)
		return
	}
	go func(searchTerm string, res []models.AreaResponse) {
		jsonString, err := json.Marshal(res)
		if err != nil {
			log.Print(err)
		}
		// set the expiration to 30 days
		err = database.RedisInstance[4].Set(context.Background(), searchTerm, string(jsonString), 30*24*time.Hour).Err()
		if err != nil {
			log.Print(err)
		}
	}(request.Search, res)
	// this endpoint is not going to change on a daily basis for the estimated price, so we can cache it for 1 day
	c.Header("Cache-Control", "public, max-age=86400, immutable")
	c.JSON(200, res)
}

func GetRatesProduct(c *gin.Context) {
	var request models.GetRatesByProductRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		c.Status(400)
		return
	}
	// check from redis cache first if there is a match
	// redisKey := request.productid + "_" + request.areaID
	redisKey := fmt.Sprintf("%s_%d", request.ProductID, request.AreaID)
	val, err := database.RedisInstance[5].Get(context.Background(), redisKey).Result()
	if err == nil {
		var res freight.WholeResult
		if err := json.Unmarshal([]byte(val), &res); err == nil {
			// this endpoint is not going to change on a daily basis for the estimated price, so we can cache it for 1 day
			c.Header("Cache-Control", "public, max-age=86400, immutable")
			c.JSON(200, res)
			return
		}
		// if there is error in unmarshalling, we will rely on manual query
		// delete the key from redis
		go func() {
			if err := database.RedisInstance[5].Del(context.Background(), redisKey).Err(); err != nil {
				log.Print(err)
			}
		}()
	}
	product := freight.CalculateFreightRequest{
		ID: request.AreaID,
	}
	err = database.MysqlInstance.
		QueryRow(
			"SELECT weight, length, width, height FROM products WHERE id = UUID_TO_BIN(?) AND deleted_at IS NULL",
			request.ProductID,
		).
		Scan(&product.Weight, &product.Length, &product.Width, &product.Height)
	if err != nil {
		if err == sql.ErrNoRows {
			c.Status(404)
			return
		}
		c.Status(500)
		return
	}
	res, err := product.CalculateFreight()
	if err != nil {
		if strings.Contains(err.Error(), "there are no rates available for this route") {
			c.Status(404)
			return
		}
		c.Status(500)
		return
	}
	go func(redisKey string, res freight.WholeResult) {
		jsonString, err := json.Marshal(res)
		if err != nil {
			log.Print(err)
		}
		// set the expiration to 1 day
		err = database.RedisInstance[5].Set(context.Background(), redisKey, string(jsonString), 24*time.Hour).Err()
		if err != nil {
			log.Print(err)
		}
	}(redisKey, res)
	// this endpoint is not going to change on a daily basis for the estimated price, so we can cache it for 1 day
	c.Header("Cache-Control", "public, max-age=86400, immutable")
	c.JSON(200, res)
}
