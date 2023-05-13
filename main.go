package main

import (
	"log"
	"os"

	authControllers "github.com/Tus1688/openmerce-backend/controllers/auth"
	customerControllers "github.com/Tus1688/openmerce-backend/controllers/customer"
	globalControllers "github.com/Tus1688/openmerce-backend/controllers/global"
	staffControllers "github.com/Tus1688/openmerce-backend/controllers/staff"
	"github.com/Tus1688/openmerce-backend/middlewares"

	"github.com/Tus1688/openmerce-backend/auth"
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
	staffControllers.NginxFSBaseUrl = os.Getenv("NGINX_FS_BASE_URL")
	staffControllers.NginxFSAuthorization = os.Getenv("NGINX_FS_AUTHORIZATION")
	log.Print("Loaded env!")
}

func initRouter() *gin.Engine {
	router := gin.Default()
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	customerAuth := router.Group("/api/v1/auth") // customer authentication are unprotected by any middleware
	{
		// user is unauthenticated
		customerAuth.POST("/register-1", authControllers.RegisterEmail)        // user get a verification code and retrieve httpOnly cookie with jwt token of the inputted email
		customerAuth.POST("/register-2", authControllers.RegisterEmailConfirm) // user input the verification code and the jwt token to confirm the email
		customerAuth.POST("/register-3", authControllers.CreateAccount)        // user input everything else to create an account
		customerAuth.POST("/login", authControllers.LoginCustomer)             // user login with email and password
		customerAuth.GET("/refresh", authControllers.RefreshTokenCustomer)     // user refresh the token
	}

	staffAuth := router.Group("/api/v1/staff/auth")
	{
		staffAuth.POST("/login", authControllers.LoginStaff)
		staffAuth.GET("/refresh", authControllers.RefreshTokenStaff)
	}

	// handle internal staff issue which won't be exposed to the public
	staffConsole := router.Group("/api/v1/staff/console").
		Use(middlewares.TokenExpiredStaff(1)).
		Use(middlewares.TokenIsSysAdmin())
	{
		staffConsole.GET("/staff", authControllers.GetStaff)
		staffConsole.POST("/staff", authControllers.AddNewStaff)
		staffConsole.PATCH("/staff", authControllers.UpdateStaff)
		staffConsole.DELETE("/staff", authControllers.DeleteStaff)
	}

	// staff dashboard is protected by token expired middleware with 3 minutes (default)
	// every staff can access the dashboard
	staffDashboard := router.Group("/api/v1/staff/dashboard")
	staffDashboard.Use(middlewares.TokenExpiredStaff(3))
	{
		// inventory only accessible by inventory user
		inventory := staffDashboard.Group("/inventory")
		inventory.Use(middlewares.TokenIsInvUser())
		{
			inventory.GET("/category", staffControllers.GetCategories)
			inventory.POST("/category", staffControllers.AddNewCategory)
			inventory.DELETE("/category", staffControllers.DeleteCategory)
			inventory.PATCH("/category", staffControllers.UpdateCategory)

			inventory.POST("/product-1", staffControllers.AddNewProduct)  // handle product meta creation
			inventory.POST("/product-2", staffControllers.AddImage)       // handle image upload
			inventory.DELETE("/product", staffControllers.DeleteProduct)  // delete product and its images
			inventory.DELETE("/product-2", staffControllers.DeleteImage)  // delete image only
			inventory.PATCH("/product-1", staffControllers.UpdateProduct) // update product (without image)
		}
	}

	// customer dashboard is protected by token expired middleware with 3 minutes (default)
	// every customer can access the dashboard
	customerDashboard := router.Group("/api/v1/customer")
	customerDashboard.Use(middlewares.TokenExpiredCustomer(3))
	{
		customerDashboard.GET("/cart", customerControllers.GetCart)
		customerDashboard.POST("/cart", customerControllers.AddToCart)
	}

	// global unprotected routes for public access
	router.GET("/api/v1/product", globalControllers.GetProduct)
	router.GET("/api/v1/category", globalControllers.GetCategory)

	return router
}
