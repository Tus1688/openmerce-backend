package controllers

import (
	"database/sql"

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
	err := database.MysqlInstance.QueryRow("select username from staffs where username = ?", request.Username).Scan(&username)
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
	_, err = database.MysqlInstance.Exec("insert into staffs (username, hashed_password, name, fin_user, inv_user, sys_admin) values (?, ?, ?, ?, ?, ?)",
		request.Username, request.Password, request.Name, request.FinUser, request.InvUser, request.SysAdmin)
	if err != nil {
		c.Status(500)
		return
	}
	c.Status(200)
}
