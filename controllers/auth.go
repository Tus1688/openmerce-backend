package controllers

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Tus1688/openmerce-backend/auth"
	"github.com/Tus1688/openmerce-backend/database"
	"github.com/Tus1688/openmerce-backend/models"
	"github.com/gin-gonic/gin"
)

func LoginCustomer(c *gin.Context) {
	var request models.ReqLogin
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Status(400)
		return
	}
	var customer models.CustomerAuth
	err := database.MysqlInstance.QueryRow("select id, hashed_password from customers where email = ?", request.Email).Scan(&customer.ID, &customer.HashedPassword)
	if err != nil {
		c.Status(401)
		return
	}
	if !customer.CheckPassword(request.Password) {
		c.Status(401)
		return
	}
	userAgent := c.GetHeader("User-Agent")
	jti := auth.GenerateRandomString(16)
	refreshToken := auth.GenerateRandomString(32)
	// make a json string that contains "user-agent": userAgent, "id": customer.ID, "jti": jti
	jsonString := strings.Join([]string{"{\"user-agent\":\"", userAgent, "\",\"id\":\"", customer.ID.String(), "\",\"jti\":\"", jti, "\"}"}, "")
	// insert into redis
	err = database.RedisInstance[1].Set(context.Background(), refreshToken, jsonString, 14*24*time.Hour).Err()
	if err != nil {
		log.Print(err)
		c.Status(500)
		return
	}
	token, err := auth.GenerateJWTAccessToken(customer.ID.String(), jti)
	if err != nil {
		c.Status(500)
		return
	}
	c.SetSameSite(http.SameSiteStrictMode)
	// 3 minutes expiration for each token
	c.SetCookie("access_token", token, 60*3, "/", "", false, true)
	c.SetCookie("refresh_token", refreshToken, 60*60*24*14, "/", "", false, true)
	c.Status(200)
}
