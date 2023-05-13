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
