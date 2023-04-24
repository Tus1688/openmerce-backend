package middlewares

import (
	"github.com/Tus1688/openmerce-backend/auth"
	"github.com/gin-gonic/gin"
)

func TokenIsSysAdmin() gin.HandlerFunc {
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
		if !claims.SysAdmin {
			c.AbortWithStatus(403)
			return
		}
		c.Next()
	}
}

func TokenIsInvUser() gin.HandlerFunc {
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
		if !claims.InvUser {
			c.AbortWithStatus(403)
			return
		}
		c.Next()
	}
}
