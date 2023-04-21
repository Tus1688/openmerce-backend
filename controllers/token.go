package controllers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Tus1688/openmerce-backend/auth"
	"github.com/Tus1688/openmerce-backend/database"
	"github.com/gin-gonic/gin"
)

type redisValueCustomer struct {
	UserAgent string `json:"user-agent"`
	Id        string `json:"id"`
	Jti       string `json:"jti"`
}

type redisValueStaff struct {
	UserAgent string `json:"user-agent"`
	Id        uint   `json:"id"`
	Username  string `json:"username"`
	FinUser   bool   `json:"FinUser"`
	InvUser   bool   `json:"InvUser"`
	SysAdmin  bool   `json:"SysAdmin"`
	Jti       string `json:"jti"`
}

func RefreshTokenCustomer(c *gin.Context) {
	refreshToken, err := c.Cookie("ref_cus")
	if err != nil {
		c.Status(401)
		return
	}
	if refreshToken == "" {
		c.Status(401)
		return
	}
	// get from redis and check if there is a key with the value of refreshToken
	res, err := database.RedisInstance[1].Get(context.Background(), refreshToken).Result()
	if err != nil {
		c.Status(401)
		return
	}
	// parse json from res and get the user-agent
	// check if the user-agent is the same as the one in the request header
	// if not, return 401
	var redisValue redisValueCustomer
	err = json.Unmarshal([]byte(res), &redisValue)
	if err != nil {
		c.Status(500)
		return
	}
	userAgent := c.GetHeader("User-Agent")
	if redisValue.UserAgent != userAgent {
		c.Status(401)
		return
	}
	// generate new access token and new refresh token
	newAccessToken, err := auth.GenerateJWTAccessTokenCustomer(redisValue.Id, redisValue.Jti)
	if err != nil {
		c.Status(500)
		return
	}
	newRefreshToken := auth.GenerateRandomString(32)
	// rename the old refresh token key to the new refresh token
	err = database.RedisInstance[1].Rename(context.Background(), refreshToken, newRefreshToken).Err()
	if err != nil {
		c.Status(500)
		return
	}
	ttl, err := database.RedisInstance[1].TTL(context.Background(), newRefreshToken).Result()
	if err != nil {
		c.Status(500)
		return
	}
	c.SetSameSite(http.SameSiteStrictMode)
	// set the new access token and new refresh token to the cookie
	c.SetCookie("ac_cus", newAccessToken, 60*3, "/", "", false, true)
	c.SetCookie("ref_cus", newRefreshToken, int(ttl.Seconds()), "/", "", false, true)
	c.Status(200)
}

func RefreshTokenStaff(c *gin.Context) {
	refreshToken, err := c.Cookie("ref_stf")
	if err != nil {
		c.Status(401)
		return
	}
	if refreshToken == "" {
		c.Status(401)
		return
	}
	// get from redis and check if there is a key with the value of refreshToken
	res, err := database.RedisInstance[2].Get(context.Background(), refreshToken).Result()
	if err != nil {
		c.Status(401)
		return
	}
	// parse json from res and get the user-agent
	// check if the user-agent is the same as the one in the request header
	// if not, return 401
	var redisValue redisValueStaff
	err = json.Unmarshal([]byte(res), &redisValue)
	if err != nil {
		c.Status(500)
		log.Print(err)
		return
	}
	userAgent := c.GetHeader("User-Agent")
	if redisValue.UserAgent != userAgent {
		c.Status(401)
		return
	}
	// generate new access token and new refresh token
	newAccessToken, err := auth.GenerateJWTAccessTokenStaff(redisValue.Id, redisValue.Username, redisValue.FinUser, redisValue.InvUser, redisValue.SysAdmin, redisValue.Jti)
	if err != nil {
		c.Status(500)
		return
	}
	newRefreshToken := auth.GenerateRandomString(32)
	// rename the old refresh token key to the new refresh token
	err = database.RedisInstance[2].Rename(context.Background(), refreshToken, newRefreshToken).Err()
	if err != nil {
		c.Status(500)
		return
	}
	ttl, err := database.RedisInstance[2].TTL(context.Background(), newRefreshToken).Result()
	if err != nil {
		c.Status(500)
		return
	}
	c.SetSameSite(http.SameSiteStrictMode)
	// set the new access token and new refresh token to the cookie
	c.SetCookie("ac_stf", newAccessToken, 60*3, "/", "", false, true)
	c.SetCookie("ref_stf", newRefreshToken, int(ttl.Seconds()), "/", "", false, true)
	c.Status(200)
}