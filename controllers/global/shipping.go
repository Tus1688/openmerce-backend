package global

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/Tus1688/openmerce-backend/database"
	"github.com/Tus1688/openmerce-backend/models"
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
		Query("SELECT id, full_name FROM shipping_areas WHERE MATCH(full_name) AGAINST(? IN BOOLEAN MODE) LIMIT 5", request.Search+"*")
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
