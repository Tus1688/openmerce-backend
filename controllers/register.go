package controllers

import (
	"context"
	"crypto/rand"
	"database/sql"
	"math/big"
	"time"

	"github.com/Tus1688/openmerce-backend/database"
	"github.com/Tus1688/openmerce-backend/models"
	"github.com/Tus1688/openmerce-backend/service/mailgun"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// user is unauthenticated and wants to register an account
func RegisterEmail(c *gin.Context) {
	var request models.ReqEmailVerification
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Status(400)
		return
	}
	row := database.MysqlInstance.QueryRow("select email from customers where email = ?", request.Email)
	var email string
	err := row.Scan(&email)
	if err != sql.ErrNoRows {
		c.JSON(409, gin.H{"error": "Email already registered"})
		return
	}
	// generate a random 6-digit number
	big := big.NewInt(999999)
	randNumber, err := rand.Int(rand.Reader, big)
	if err != nil {
		c.Status(500)
		return
	}

	_, err = database.RedisInstance[0].Get(context.Background(), request.Email).Result()
	if err != redis.Nil {
		c.JSON(409, gin.H{"error": "You have already requested a verification code. Please wait 5 minutes before requesting another one."})
		return
	}

	_, err = database.RedisInstance[0].Set(context.Background(), request.Email, randNumber.String(), time.Minute*5).Result()
	if err != nil {
		c.Status(500)
		return
	}
	err = mailgun.SendEmail(mailgun.MailgunSend{
		FromName:    "Openmerce Auth Service",
		FromAddress: "noreply",
		To:          request.Email,
		Subject:     "Openmerce Email Verification",
		Body:        "Your verification code is: " + randNumber.String(),
	})
	if err != nil {
		c.Status(500)
		return
	}
	c.Status(200)
}
