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

package middlewares

import (
	"time"

	"github.com/Tus1688/openmerce-backend/auth"
	"github.com/gin-gonic/gin"
)

func TokenExpiredStaff(expiredIn int) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("ac_stf")
		if err != nil {
			c.AbortWithStatus(401)
			return
		}
		claims, err := auth.ExtractClaimAccessTokenStaff(tokenString)
		if err != nil {
			c.AbortWithStatus(401)
			return
		}
		if claims.IssuedAt.Time.Add(time.Duration(expiredIn) * time.Minute).Before(time.Now()) {
			c.AbortWithStatus(401)
			return
		}
		c.Next()
	}
}

func TokenExpiredCustomer(expiredIn int) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("ac_cus")
		if err != nil {
			c.AbortWithStatus(401)
			return
		}
		claims, err := auth.ExtractClaimAccessTokenCustomer(tokenString)
		if err != nil {
			c.AbortWithStatus(401)
			return
		}
		// verify the struct of token as email verification token has same secret key
		if claims.IssuedAt.Time.Add(time.Duration(expiredIn)*time.Minute).Before(time.Now()) || claims.Uid == "" {
			c.AbortWithStatus(401)
			return
		}
		c.Next()
	}
}
