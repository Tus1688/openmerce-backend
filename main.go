package main

import (
	"log"
	"os"

	"github.com/Tus1688/openmerce-backend/auth"
	"github.com/Tus1688/openmerce-backend/controllers"
	"github.com/Tus1688/openmerce-backend/database"
	"github.com/Tus1688/openmerce-backend/service/mailgun"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	loadEnv()
	err := database.NewMysql()
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Connected to mysql!")
	err = database.NewRedis()
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Connected to redis!")
	router := initRouter()
	router.Run(":6000")
}

func loadEnv() {
	if os.Getenv("GIN_MODE") != "release" {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}
	auth.JwtKey = []byte(os.Getenv("JWT_KEY"))
	mailgun.ReadEnv()
	log.Print("Loaded env!")
}

func initRouter() *gin.Engine {
	router := gin.Default()
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	auth := router.Group("/api/v1/auth")
	{
		// user is unauthenticated
		auth.POST("/register-1", controllers.RegisterEmail)        // user get a verification code and retrieve httpOnly cookie with jwt token of the inputted email
		auth.POST("/register-2", controllers.RegisterEmailConfirm) // user input the verification code and the jwt token to confirm the email
		auth.POST("/register-3", controllers.CreateAccount)        // user input everything else to create an account
		auth.POST("/login", controllers.LoginCustomer)             // user login with email and password
	}
	return router
}
