package main

import (
	"log"
	"os"

	"github.com/Tus1688/openmerce-backend/middlewares"

	"github.com/Tus1688/openmerce-backend/auth"
	"github.com/Tus1688/openmerce-backend/controllers"
	"github.com/Tus1688/openmerce-backend/database"
	"github.com/Tus1688/openmerce-backend/service/mailgun"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
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
	err = database.InitAdminAccount()
	if err != nil {
		log.Fatal(err)
	}
	router := initRouter()
	err = router.Run(":6000")
	if err != nil {
		log.Fatal(err)
	}
}

func loadEnv() {
	auth.JwtKeyCustomer = []byte(os.Getenv("JWT_KEY_CUSTOMER"))
	auth.JwtKeyStaff = []byte(os.Getenv("JWT_KEY_STAFF"))
	mailgun.ReadEnv()
	log.Print("Loaded env!")
}

func initRouter() *gin.Engine {
	router := gin.Default()
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	customerAuth := router.Group("/api/v1/auth") // customer authentication are unprotected by any middleware
	{
		// user is unauthenticated
		customerAuth.POST("/register-1", controllers.RegisterEmail)        // user get a verification code and retrieve httpOnly cookie with jwt token of the inputted email
		customerAuth.POST("/register-2", controllers.RegisterEmailConfirm) // user input the verification code and the jwt token to confirm the email
		customerAuth.POST("/register-3", controllers.CreateAccount)        // user input everything else to create an account
		customerAuth.POST("/login", controllers.LoginCustomer)             // user login with email and password
		customerAuth.GET("/refresh", controllers.RefreshTokenCustomer)     // user refresh the token
	}

	staffAuth := router.Group("/api/v1/staff/auth")
	{
		staffAuth.POST("/login", controllers.LoginStaff)
		staffAuth.GET("/refresh", controllers.RefreshTokenStaff)
	}

	// handle internal staff issue which won't be exposed to the public
	staffConsole := router.Group("/api/v1/staff/console").
		Use(middlewares.TokenExpiredStaff(1)).
		Use(middlewares.TokenIsSysAdmin())
	{
		staffConsole.GET("/staff", controllers.GetStaff)
		staffConsole.POST("/staff", controllers.AddNewStaff)
		staffConsole.PATCH("/staff", controllers.UpdateStaff)
		staffConsole.DELETE("/staff", controllers.DeleteStaff)
	}

	staffDashboard := router.Group("/api/v1/staff/dashboard").
		// staff dashboard is protected by token expired middleware with 3 minutes (default)
		// every staff can access the dashboard
		Use(middlewares.TokenExpiredStaff(3)).
		Use(middlewares.TokenIsSysAdmin())
	{
		staffDashboard.GET("/test", func(context *gin.Context) {
			context.JSON(200, gin.H{
				"message": "hello world",
			})
		})
	}

	return router
}
