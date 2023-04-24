package auth

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Tus1688/openmerce-backend/auth"
	"github.com/Tus1688/openmerce-backend/database"
	"github.com/Tus1688/openmerce-backend/models"
	"github.com/gin-gonic/gin"
)

func LoginCustomer(c *gin.Context) {
	var request models.ReqLoginCustomer
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Status(400)
		return
	}
	var customer models.CustomerAuth
	err := database.MysqlInstance.
		QueryRow("select id, hashed_password, first_name, last_name from customers where email = ?", request.Email).
		Scan(&customer.ID, &customer.HashedPassword, &customer.FirstName, &customer.LastName)
	if err != nil {
		c.Status(401)
		return
	}
	if !customer.CheckPassword(request.Password) {
		c.Status(401)
		return
	}
	userAgent := c.GetHeader("User-Agent")
	jti := auth.GenerateRandomString(16)
	refreshToken := auth.GenerateRandomString(32)
	// make a json string that contains "user-agent": userAgent, "id": customer.ID, "jti": jti
	jsonString := strings.Join([]string{"{\"user-agent\":\"", userAgent, "\",\"id\":\"", customer.ID.String(), "\",\"jti\":\"", jti, "\"}"}, "")
	// insert into redis
	err = database.RedisInstance[1].Set(context.Background(), refreshToken, jsonString, 14*24*time.Hour).Err()
	if err != nil {
		c.Status(500)
		return
	}
	token, err := auth.GenerateJWTAccessTokenCustomer(customer.ID.String(), jti)
	if err != nil {
		c.Status(500)
		return
	}
	c.SetSameSite(http.SameSiteStrictMode)
	// 3 minutes expiration access token and 14 days expiration refresh token
	c.SetCookie("ac_cus", token, 60*3, "/", "", false, true)
	c.SetCookie("ref_cus", refreshToken, 60*60*24*14, "/", "", false, true)
	c.JSON(200, gin.H{
		"first_name": customer.FirstName,
		"last_name":  customer.LastName,
	})
}

func LoginStaff(c *gin.Context) {
	var request models.ReqLoginStaff
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Status(400)
		return
	}
	var staff models.StaffAuth
	err := database.MysqlInstance.
		QueryRow("SELECT id, username, hashed_password, fin_user, inv_user, sys_admin FROM staffs WHERE username = ? and deleted_at is null", request.Username).
		Scan(&staff.ID, &staff.Username, &staff.HashedPassword, &staff.FinUser, &staff.InvUser, &staff.SysAdmin)
	if err != nil {
		c.Status(401)
		return
	}
	if !staff.CheckPassword(request.Password) {
		c.Status(401)
		return
	}
	userAgent := c.GetHeader("User-Agent")
	jti := auth.GenerateRandomString(16)
	refreshToken := auth.GenerateRandomString(32)
	// make a json string that contains "user-agent": userAgent, "id": staff.ID, "username": staff.Username, "FinUser": staff.FinUser, "InvUser": staff.InvUser, "SysAdmin": staff.SysAdmin, "jti": jti
	jsonString := strings.Join([]string{"{\"user-agent\":\"", userAgent, "\",\"id\":",
		strconv.FormatUint(uint64(staff.ID), 10), ",\"username\":\"", staff.Username, "\",\"FinUser\":",
		strconv.FormatBool(staff.FinUser), ",\"InvUser\":", strconv.FormatBool(staff.InvUser), ",\"SysAdmin\":",
		strconv.FormatBool(staff.SysAdmin), ",\"jti\":\"", jti, "\"}"}, "")
	err = database.RedisInstance[2].Set(context.Background(), refreshToken, jsonString, 14*24*time.Hour).Err()
	if err != nil {
		c.Status(500)
		return
	}
	token, err := auth.GenerateJWTAccessTokenStaff(staff.ID, staff.Username, staff.FinUser, staff.InvUser, staff.SysAdmin, jti)
	if err != nil {
		c.Status(500)
		return
	}
	c.SetSameSite(http.SameSiteStrictMode)
	// 3 minutes expiration access token and 14 days expiration refresh token
	c.SetCookie("ac_stf", token, 60*3, "/", "", false, true)
	c.SetCookie("ref_stf", refreshToken, 60*60*24*14, "/", "", false, true)
	c.JSON(200, gin.H{
		"username":  staff.Username,
		"fin_user":  staff.FinUser,
		"inv_user":  staff.InvUser,
		"sys_admin": staff.SysAdmin,
	})
}
