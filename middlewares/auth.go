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
