package database

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"

	_ "github.com/go-sql-driver/mysql"
)

const (
	maxOpenConns    = 10
	maxIdleConns    = 5
	connMaxLifetime = 5 * time.Minute
)

var MysqlInstance *sql.DB

func NewMysql() error {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=UTC", dbUser, dbPass, dbHost, dbPort, dbName)

	var err error
	MysqlInstance, err = sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	if err = MysqlInstance.Ping(); err != nil {
		return err
	}
	MysqlInstance.SetMaxOpenConns(maxOpenConns)
	MysqlInstance.SetMaxIdleConns(maxIdleConns)
	MysqlInstance.SetConnMaxLifetime(connMaxLifetime)
	return nil
}

func InitAdminAccount() error {
	username := os.Getenv("ADMIN_USERNAME")
	if username == "" {
		username = "admin"
	}
	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		return fmt.Errorf("admin password is empty")
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	//	insert into staff table
	_, err = MysqlInstance.Exec("INSERT INTO staffs (username, hashed_password, name, fin_user, inv_user, sys_admin) VALUES (?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE hashed_password=VALUES(hashed_password), name=VALUES(name), fin_user=VALUES(fin_user), inv_user=VALUES(inv_user), sys_admin=VALUES(sys_admin)",
		username, string(bytes), "superadmin", true, true, true)
	if err != nil {
		return err
	}
	return nil
}
