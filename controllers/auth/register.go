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
	"crypto/rand"
	"database/sql"
	"math/big"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/Tus1688/openmerce-backend/auth"
	"github.com/Tus1688/openmerce-backend/database"
	"github.com/Tus1688/openmerce-backend/models"
	"github.com/Tus1688/openmerce-backend/service/mailgun"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RegisterEmail is for unauthenticated and wants to register an account
func RegisterEmail(c *gin.Context) {
	var request models.ReqEmailVerification
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Status(400)
		return
	}
	var email string
	err := database.MysqlInstance.QueryRow("select email from customers where email = ?", request.Email).Scan(&email)
	if err != nil && err != sql.ErrNoRows {
		c.Status(500)
		return
	}
	// validate the email address
	if _, err := mail.ParseAddress(request.Email); err != nil {
		c.Status(400)
		return
	}

	if email != "" {
		c.JSON(409, gin.H{"error": "Email already registered"})
		return
	}

	_, err = database.RedisInstance[0].Get(context.Background(), request.Email).Result()
	if err != redis.Nil {
		c.JSON(
			409,
			gin.H{"error": "You have already requested a verification code. Please wait 5 minutes before requesting another one."},
		)
		return
	}

	// generate a random 6-digit number
	code := big.NewInt(999999)
	randNumber, err := rand.Int(rand.Reader, code)
	if err != nil {
		c.Status(500)
		return
	}

	_, err = database.RedisInstance[0].Set(
		context.Background(), request.Email, randNumber.String(), time.Minute*5,
	).Result()
	if err != nil {
		c.Status(500)
		return
	}
	err = mailgun.SendEmail(
		mailgun.Send{
			FromName:    "Openmerce Auth Service",
			FromAddress: "noreply",
			To:          request.Email,
			Subject:     "Openmerce Email Verification",
			Body:        "Your verification code is: " + randNumber.String(),
		},
	)
	if err != nil {
		c.Status(500)
		return
	}
	// create a JWT token with the email and send it to the user as a httpOnly cookie
	tokenString, err := auth.GenerateJWTEmailVerification(request.Email, false)
	if err != nil {
		c.Status(500)
		return
	}
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("email", tokenString, 300, "/", "", false, true)
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
		// the reason we give 403 because the token should be deleted from the user's browser after 5 minutes
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
	tokenString, err := auth.GenerateJWTEmailVerification(request.Email, true)
	if err != nil {
		c.Status(500)
		return
	}
	c.SetSameSite(http.SameSiteStrictMode)
	// set the cookie to expire in 10 minutes so that the user can register an account
	c.SetCookie("email", tokenString, 600, "/", "", false, true)
	c.Status(200)
}

func CreateAccount(c *gin.Context) {
	var request models.ReqNewAccount
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
	if !claims.Status {
		// the user has not verified their email and should not be able to send this request
		// therefore 403 will be returned
		c.Status(403)
		return
	}
	if claims.IssuedAt.Time.Add(time.Minute * 10).Before(time.Now()) {
		// the reason we give 401 because the token should be deleted from the user's browser after 10 minutes
		// check register-2 for more info
		c.Status(401)
		return
	}
	// check if the email is the same as the one in the cookie
	if request.Email != claims.Email {
		// attempt to change email this is not allowed and should be blocked
		c.Status(403)
		return
	}
	// validate the password
	if !request.PasswordIsValid() {
		c.JSON(
			400,
			gin.H{"error": "Password must be at least 8 characters long and contain at least one uppercase letter, one lowercase letter, and one number"},
		)
		return
	}
	// validate the gender
	if request.Gender != "male" && request.Gender != "female" {
		c.Status(400)
		return
	}
	// hash the password
	if err := request.HashPassword(); err != nil {
		c.Status(500)
		return
	}
	// try to input the user into the database, if there is a conflict then the email is already registered and should return 409
	_, err = database.MysqlInstance.
		Exec(
			"insert into customers (email, hashed_password, first_name, last_name, birth_date, gender) values (?, ?, ?, ?, ?, ?)",
			request.Email, request.Password, request.FirstName, request.LastName, request.BirthDate, request.Gender,
		)
	if err != nil {
		// check from error if it is a duplicate entry error
		if strings.Contains(err.Error(), "Duplicate entry") {
			c.JSON(409, gin.H{"error": "Email already registered"})
			return
		}
		c.Status(500)
		return
	}
	c.Status(200)
}
