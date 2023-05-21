package customer

import (
	"database/sql"
	"strings"

	"github.com/Tus1688/openmerce-backend/auth"
	"github.com/Tus1688/openmerce-backend/database"
	"github.com/Tus1688/openmerce-backend/models"
	"github.com/gin-gonic/gin"
)

func AddAddress(c *gin.Context) {
	var request models.CreateAddress
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Status(400)
		return
	}
	// token should be valid and exist as it is protected by TokenExpiredCustomer middleware
	token, _ := c.Cookie("ac_cus")
	claims, err := auth.ExtractClaimAccessTokenCustomer(token)
	if err != nil {
		c.Status(401)
		return
	}
	customerId := claims.Uid
	_, err = database.MysqlInstance.
		Exec("INSERT INTO customer_addresses (customer_refer, label, full_address, note, recipient_name, phone_number, shipping_area_refer, postal_code) VALUES (UUID_TO_BIN(?), ?, ?, ?, ?, ?, ?, ?)",
			customerId, request.Label, request.FullAddress, request.Note, request.RecipientName, request.PhoneNumber, request.ShippingArea, request.PostalCode)
	if err != nil {
		if strings.Contains(err.Error(), "shipping_area_refer") {
			c.JSON(409, gin.H{"error": "shipping area not found"})
			return
		}
		if strings.Contains(err.Error(), "Duplicate") {
			c.JSON(409, gin.H{"error": "address already exists"})
			return
		}
		c.Status(500)
		return
	}
	c.Status(200)
}

func GetAddress(c *gin.Context) {
	// token should be valid and exist as it is protected by TokenExpiredCustomer middleware
	token, _ := c.Cookie("ac_cus")
	claims, err := auth.ExtractClaimAccessTokenCustomer(token)
	if err != nil {
		c.Status(401)
		return
	}
	customerId := claims.Uid
	var request models.APICommonQueryUUID
	if err := c.ShouldBindQuery(&request); err == nil {
		var response models.AddressResponseDetail
		err := database.MysqlInstance.QueryRow(`
					select BIN_TO_UUID(c.id), c.label, c.full_address, c.note, c.recipient_name, c.phone_number, s.full_name, c.postal_code
					from customer_addresses c, shipping_areas s
					where c.shipping_area_refer = s.id and c.customer_refer = UUID_TO_BIN(?) and c.id = UUID_TO_BIN(?)
					`, customerId, request.ID).
			Scan(&response.ID, &response.Label, &response.FullAddress, &response.Note, &response.RecipientName, &response.PhoneNumber, &response.ShippingArea, &response.PostalCode)
		if err != nil {
			if err == sql.ErrNoRows {
				c.Status(404)
				return
			}
			c.Status(500)
			return
		}
		c.JSON(200, response)
		return
	}
	if request.ID != "" {
		c.Status(400)
		return
	}
	// if there is no query string, return all addresses
	rows, err := database.MysqlInstance.
		Query("SELECT BIN_TO_UUID(id), label, full_address, note, recipient_name, phone_number FROM customer_addresses WHERE customer_refer = UUID_TO_BIN(?)", customerId)
	if err != nil {
		c.Status(500)
		return
	}
	var response []models.AddressResponse
	for rows.Next() {
		var row models.AddressResponse
		if err := rows.Scan(&row.ID, &row.Label, &row.FullAddress, &row.Note, &row.RecipientName, &row.PhoneNumber); err != nil {
			c.Status(500)
			return
		}
		response = append(response, row)
	}
	if len(response) == 0 {
		c.Status(404)
		return
	}
	c.JSON(200, response)
}

func DeleteAddress(c *gin.Context) {
	var request models.APICommonQueryUUID
	if err := c.ShouldBindQuery(&request); err != nil {
		c.Status(400)
		return
	}
	// token should be valid and exist as it is protected by TokenExpiredCustomer middleware
	token, _ := c.Cookie("ac_cus")
	claims, err := auth.ExtractClaimAccessTokenCustomer(token)
	if err != nil {
		c.Status(401)
		return
	}
	customerId := claims.Uid
	res, err := database.MysqlInstance.
		Exec("DELETE FROM customer_addresses WHERE customer_refer = UUID_TO_BIN(?) AND id = UUID_TO_BIN(?)", customerId, request.ID)
	if err != nil {
		// 1451 = foreign key constraint
		if strings.Contains(err.Error(), "1451") {
			c.JSON(409, gin.H{"error": "address is used in an order"})
			return
		}
		c.Status(500)
		return
	}
	if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
		c.Status(404)
		return
	}
	c.Status(200)
}

func UpdateAddress(c *gin.Context) {
	var request models.UpdateAddress
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Status(400)
		return
	}
	// token should be valid and exist as it is protected by TokenExpiredCustomer middleware
	token, _ := c.Cookie("ac_cus")
	claims, err := auth.ExtractClaimAccessTokenCustomer(token)
	if err != nil {
		c.Status(401)
		return
	}
	customerId := claims.Uid
	query := "UPDATE customer_addresses SET updated_at = CURRENT_TIMESTAMP"
	var args []interface{}
	if request.Label != "" {
		query += ", label = ?"
		args = append(args, request.Label)
	}
	if request.FullAddress != "" {
		query += ", full_address = ?"
		args = append(args, request.FullAddress)
	}
	if request.Note != "" {
		query += ", note = ?"
		args = append(args, request.Note)
	}
	if request.RecipientName != "" {
		query += ", recipient_name = ?"
		args = append(args, request.RecipientName)
	}
	if request.PhoneNumber != "" {
		query += ", phone_number = ?"
		args = append(args, request.PhoneNumber)
	}
	if request.ShippingArea != 0 {
		query += ", shipping_area_refer = ?"
		args = append(args, request.ShippingArea)
	}
	if request.PostalCode != "" {
		query += ", postal_code = ?"
		args = append(args, request.PostalCode)
	}
	query += " WHERE customer_refer = UUID_TO_BIN(?) AND id = UUID_TO_BIN(?)"
	args = append(args, customerId, request.ID)
	res, err := database.MysqlInstance.Exec(query, args...)
	if err != nil {
		// 1452 = foreign key constraint
		if strings.Contains(err.Error(), "1452") {
			c.JSON(409, gin.H{"error": "shipping area not found"})
			return
		}
		if strings.Contains(err.Error(), "Duplicate") {
			c.JSON(409, gin.H{"error": "label already exist"})
			return
		}
		c.Status(500)
		return
	}
	if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
		// address might have been deleted / no change has been made (possible if user send more than 1 request in 1 second)
		c.Status(404)
		return
	}
	c.Status(200)
}
