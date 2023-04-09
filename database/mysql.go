package database

import (
	"database/sql"
	"fmt"
	"os"
	"time"

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
