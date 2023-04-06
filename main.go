package main

import (
	"log"
	"os"

	"github.com/Tus1688/openmerce-backend/auth"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	loadEnv()
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
	log.Print("Loaded env!")
}

func initRouter() *gin.Engine {
	router := gin.Default()
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	auth := router.Group("/api/v1/auth")
	{
		auth.POST("/register")
	}
	return router
}
