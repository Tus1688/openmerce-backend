package customer

import (
	"database/sql"
	"net/mail"

	"github.com/Tus1688/openmerce-backend/auth"
	"github.com/Tus1688/openmerce-backend/database"
	"github.com/Tus1688/openmerce-backend/logging"
	"github.com/Tus1688/openmerce-backend/models"
	"github.com/gin-gonic/gin"
)

func GetProfile(c *gin.Context) {
	// the token should be valid and exist as it is protected by TokenExpiredCustomer middleware
	token, _ := c.Cookie("ac_cus")
	claims, err := auth.ExtractClaimAccessTokenCustomer(token)
	if err != nil {
		c.Status(401)
		return
	}
	customerId := claims.Uid
	var response models.CustomerProfile
	err = database.MysqlInstance.
		QueryRow(`
			SELECT email, COALESCE(phone_number, ''), first_name, last_name, birth_date, gender FROM customers
			WHERE id = UUID_TO_BIN(?)
		`, customerId).
		Scan(&response.Email, &response.PhoneNumber, &response.FirstName, &response.LastName, &response.BirthDate, &response.Gender)
	if err != nil {
		if err == sql.ErrNoRows {
			// this shouldn't happen as the customer who has the token should exist
			c.Status(403)
			return
		}
		c.Status(500)
		return
	}
	c.JSON(200, response)
}

func UpdateProfile(c *gin.Context) {
	var request models.CustomerProfile
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Status(400)
		return
	}
	// the token should be valid and exist as it is protected by TokenExpiredCustomer middleware
	token, _ := c.Cookie("ac_cus")
	claims, err := auth.ExtractClaimAccessTokenCustomer(token)
	if err != nil {
		c.Status(401)
		return
	}
	customerId := claims.Uid
	query := "UPDATE customers SET updated_at = CURRENT_TIMESTAMP"
	var args []interface{}
	if request.FirstName != "" {
		query += ", first_name = ?"
		args = append(args, request.FirstName)
	}
	if request.LastName != "" {
		query += ", last_name = ?"
		args = append(args, request.LastName)
	}
	if request.Email != "" {
		// check if the email is valid or not
		if _, err := mail.ParseAddress(request.Email); err != nil {
			c.Status(400)
			return
		}
		query += ", email = ?"
		args = append(args, request.Email)
	}
	if request.PhoneNumber != "" {
		query += ", phone_number = ?"
		args = append(args, request.PhoneNumber)
	}
	if !request.BirthDate.IsZero() {
		query += ", birth_date = ?"
		args = append(args, request.BirthDate)
	}
	if request.Gender != "" {
		if request.Gender != "male" && request.Gender != "female" {
			c.Status(400)
			return
		}
		query += ", gender = ?"
		args = append(args, request.Gender)
	}
	query += " WHERE id = UUID_TO_BIN(?)"
	args = append(args, customerId)
	_, err = database.MysqlInstance.Exec(query, args...)
	if err != nil {
		c.Status(500)
		return
	}
	c.Status(200)
}

func UpdatePassword(c *gin.Context) {
	var request models.ChangePassword
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Status(400)
		return
	}
	// the token should be valid and exist as it is protected by TokenExpiredCustomer middleware
	token, _ := c.Cookie("ac_cus")
	claims, err := auth.ExtractClaimAccessTokenCustomer(token)
	if err != nil {
		c.Status(401)
		return
	}
	customerId := claims.Uid
	var oldPassword string
	err = database.MysqlInstance.
		QueryRow("SELECT hashed_password FROM customers WHERE id = UUID_TO_BIN(?)", customerId).
		Scan(&oldPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			// this shouldn't happen as the customer who has the token should exist
			c.Status(403)
			return
		}
		c.Status(500)
		return
	}
	if !request.CheckPassword(oldPassword) {
		c.Status(401)
		return
	}
	// check the password requirement
	if !request.PasswordIsValid() {
		c.JSON(409, gin.H{"error": "password requirement not met"})
		return
	}
	//	hash the password
	if err := request.HashPassword(); err != nil {
		go logging.InsertLog(logging.ERROR, "UpdatePassword"+err.Error())
		c.Status(500)
		return
	}
	_, err = database.MysqlInstance.Exec("UPDATE customers SET hashed_password = ? WHERE id = UUID_TO_BIN(?)", request.NewPassword, customerId)
	if err != nil {
		c.Status(500)
		return
	}
	c.Status(200)
}
