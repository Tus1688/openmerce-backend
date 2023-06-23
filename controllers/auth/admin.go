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
	"database/sql"
	"os"

	"github.com/Tus1688/openmerce-backend/database"
	"github.com/Tus1688/openmerce-backend/models"
	"github.com/gin-gonic/gin"
)

func AddNewStaff(c *gin.Context) {
	var request models.NewStaff
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Status(400)
		return
	}
	// check if there is a staff with the same username
	var username string
	err := database.MysqlInstance.QueryRow(
		"select username from staffs where username = ?", request.Username,
	).Scan(&username)
	if err != nil && err != sql.ErrNoRows {
		c.Status(500)
		return
	}
	if username != "" {
		c.JSON(409, gin.H{"error": "Username already exists"})
		return
	}
	// hash the password
	if err := request.HashPassword(); err != nil {
		c.Status(500)
		return
	}
	// insert the new staff
	_, err = database.MysqlInstance.Exec(
		"insert into staffs (username, hashed_password, name, fin_user, inv_user, sys_admin) values (?, ?, ?, ?, ?, ?)",
		request.Username, request.Password, request.Name, request.FinUser, request.InvUser, request.SysAdmin,
	)
	if err != nil {
		c.Status(500)
		return
	}
	c.Status(201)
}

func GetStaff(c *gin.Context) {
	var request models.APICommonQueryID
	if err := c.ShouldBindQuery(&request); err != nil {
		//	send all staffs if there is no id in the request parameters
		rows, err := database.MysqlInstance.Query("select id, username, name, fin_user, inv_user, sys_admin from staffs")
		if err != nil {
			c.Status(500)
			return
		}
		defer rows.Close()
		var staffs []models.ListStaff
		for rows.Next() {
			var staff models.ListStaff
			if err := rows.Scan(
				&staff.ID, &staff.Username, &staff.Name, &staff.FinUser, &staff.InvUser, &staff.SysAdmin,
			); err != nil {
				c.Status(500)
				return
			}
			staffs = append(staffs, staff)
		}
		c.JSON(200, staffs)
		return
	}
	var staff models.ListStaff
	err := database.MysqlInstance.
		QueryRow("select id, username, name, fin_user, inv_user, sys_admin from staffs where id = ?", request.ID).
		Scan(&staff.ID, &staff.Username, &staff.Name, &staff.FinUser, &staff.InvUser, &staff.SysAdmin)
	if err != nil {
		if err == sql.ErrNoRows {
			c.Status(404)
			return
		}
		c.Status(500)
		return
	}
	c.JSON(200, staff)
}

func UpdateStaff(c *gin.Context) {
	var request models.UpdateStaff
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Status(400)
		return
	}
	username := os.Getenv("ADMIN_USERNAME")
	if username == "" {
		username = "admin"
	}
	var superAdminID uint
	err := database.MysqlInstance.QueryRow("select id from staffs where username = ?", username).Scan(&superAdminID)
	if err != nil {
		c.Status(500)
		return
	}
	if request.ID == superAdminID {
		c.Status(403)
		return
	}
	if request.Password != "" {
		if !request.PasswordIsValid() {
			c.JSON(
				400,
				gin.H{"error": "Password must contain at least 8 characters, 1 uppercase, 1 lowercase, 1 special character and 1 number"},
			)
			return
		}
		if err := request.HashPassword(); err != nil {
			c.Status(500)
			return
		}
	}
	query := "UPDATE staffs SET updated_at = CURRENT_TIMESTAMP"
	var args []interface{}
	if request.Name != "" {
		query += ", name = ?"
		args = append(args, request.Name)
	}
	if request.Password != "" {
		query += ", hashed_password = ?"
		args = append(args, request.Password)
	}
	query += ", fin_user = ?, inv_user = ?, sys_admin = ? WHERE id = ?"
	args = append(args, request.FinUser, request.InvUser, request.SysAdmin, request.ID)

	res, err := database.MysqlInstance.Exec(query, args...)
	if err != nil {
		c.Status(500)
		return
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		c.Status(500)
		return
	}
	if rowsAffected == 0 {
		c.Status(404)
		return
	}
	c.Status(200)
}

func DeleteStaff(c *gin.Context) {
	var request models.APICommonQueryID
	if err := c.ShouldBindQuery(&request); err != nil {
		c.Status(400)
		return
	}
	username := os.Getenv("ADMIN_USERNAME")
	if username == "" {
		username = "admin"
	}
	var superAdminID uint
	err := database.MysqlInstance.QueryRow("select id from staffs where username = ?", username).Scan(&superAdminID)
	if err != nil {
		c.Status(500)
		return
	}
	// the super admin id won't be higher than int
	if request.ID == int(superAdminID) {
		c.Status(403)
		return
	}
	//	update the deleted_at column to the current timestamp
	res, err := database.MysqlInstance.Exec(
		"UPDATE staffs SET deleted_at = CURRENT_TIMESTAMP WHERE id = ? AND deleted_at is null", request.ID,
	)
	if err != nil {
		c.Status(500)
		return
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		c.Status(500)
		return
	}
	if rowsAffected == 0 {
		c.Status(404)
		return
	}
	c.Status(200)
}
