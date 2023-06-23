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

package auth

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
	Remember  bool   `json:"remember_me"`
}

type redisValueStaff struct {
	UserAgent string `json:"user-agent"`
	Id        uint   `json:"id"`
	Username  string `json:"username"`
	FinUser   bool   `json:"FinUser"`
	InvUser   bool   `json:"InvUser"`
	SysAdmin  bool   `json:"SysAdmin"`
	Jti       string `json:"jti"`
	Remember  bool   `json:"remember_me"`
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
	if redisValue.Remember {
		c.SetCookie("ac_cus", newAccessToken, 60*3, "/", "", false, true)
		c.SetCookie("ref_cus", newRefreshToken, int(ttl.Seconds()), "/", "", false, true)
	} else {
		c.SetCookie("ac_cus", newAccessToken, 0, "/", "", false, true)
		c.SetCookie("ref_cus", newRefreshToken, 0, "/", "", false, true)
	}
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
	newAccessToken, err := auth.GenerateJWTAccessTokenStaff(
		redisValue.Id, redisValue.Username, redisValue.FinUser, redisValue.InvUser, redisValue.SysAdmin, redisValue.Jti,
	)
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
	if redisValue.Remember {
		c.SetCookie("ac_stf", newAccessToken, 60*3, "/", "", false, true)
		c.SetCookie("ref_stf", newRefreshToken, int(ttl.Seconds()), "/", "", false, true)
	} else {
		c.SetCookie("ac_stf", newAccessToken, 0, "/", "", false, true)
		c.SetCookie("ref_stf", newRefreshToken, 0, "/", "", false, true)
	}
	c.Status(200)
}
