package controllers

import (
	"context"
	"crypto/rand"
	"database/sql"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/Tus1688/openmerce-backend/auth"
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
	// create a JWT token with the email and send it to the user as a httpOnly cookie
	tokenstring, err := auth.GenerateJWTEmailVerification(request.Email, false)
	if err != nil {
		c.Status(500)
		return
	}
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("email", tokenstring, 300, "/", "", false, true)
	c.Status(200)
}

func RegisterEmailConfirm(c *gin.Context) {
	var request models.ReqEmailVerificationConfirmation
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Status(400)
		return
	}
	// get cookie with name of "email"
	cookie, err := c.Cookie("email")
	if err != nil {
		// User must have deleted the cookie
		c.Status(403)
		return
	}
	// verify the cookie
	claims, err := auth.ExtractClaimEmailVerification(cookie)
	if err != nil {
		c.Status(401)
		return
	}
	if claims.Status {
		// the user has already verified their email
		c.Status(403)
		return
	}
	if claims.IssuedAt.Time.Add(time.Minute * 5).Before(time.Now()) {
		// the reason we give 403 is because the token should be deleted from the user's browser after 5 minutes
		// same as the token in redis
		c.Status(403)
		return
	}
	if request.Email != claims.Email {
		// attempt to change email this is not allowed and should be blocked
		c.Status(403)
		return
	}
	// check if the code is correct
	code, err := database.RedisInstance[0].Get(context.Background(), request.Email).Result()
	if err != nil {
		c.Status(500)
		return
	}
	if code != strconv.Itoa(request.Code) {
		c.JSON(401, gin.H{"error": "Invalid verification code"})
		return
	}
	// generate a JWT token with the email and send it to the user as a httpOnly cookie
	tokenstring, err := auth.GenerateJWTEmailVerification(request.Email, true)
	if err != nil {
		c.Status(500)
		return
	}
	c.SetSameSite(http.SameSiteStrictMode)
	// set the cookie to expire in 10 minutes so that the user can register an account
	c.SetCookie("email", tokenstring, 600, "/", "", false, true)
	c.Status(200)
}
