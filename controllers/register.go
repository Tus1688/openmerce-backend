package controllers

import (
	"github.com/Tus1688/openmerce-backend/models"
	"github.com/gin-gonic/gin"
)

// user is unauthenticated and wants to register an account
func RegisterEmail(c *gin.Context) {
	var request models.ReqEmailVerification
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Status(400)
		return
	}

}
