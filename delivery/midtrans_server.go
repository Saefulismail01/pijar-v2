package delivery

import (
	"log"
	"time"

	"pijar/config"
	"pijar/delivery/controller"
	"pijar/middleware"
	"pijar/repository"
	"pijar/usecase"
	"pijar/utils/service"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

func NewMidtransServer() {
	// Connect to DB
	db, _, err := config.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	// Setup Gin router
	r := gin.Default()

	// Log all routes
	r.Use(func(c *gin.Context) {
		log.Printf("Request: %s %s", c.Request.Method, c.Request.URL.Path)
		c.Next()
	})

	// Create a RouterGroup
	api := r.Group("/pijar")
	
	// Log startup routes
	log.Println("Starting server with prefix: /pijar")

	// Initialize repositories
	userRepo := repository.NewUserRepo(db)
	productRepo := repository.NewProductRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)

	// Initialize services
	jwtService := service.NewJwtService("SECRETKU", "PIJAR-APP", time.Hour*2)
	restyClient := resty.New()
	midtransService := service.NewMidtransService(restyClient)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	// Initialize usecases
	userUsecase := usecase.NewUserUsecase(userRepo)
	authUsecase := usecase.NewAuthUsecase(userRepo, jwtService)
	paymentUsecase := usecase.NewPaymentUsecase(midtransService, productRepo, transactionRepo)

	// Initialize controllers and setup routes
	userController := controller.NewUserController(api, userUsecase, jwtService, authMiddleware)
	userController.Route()

	authController := controller.NewAuthController(api, jwtService, authUsecase)
	authController.Route()

	paymentController := controller.NewPaymentController(api, paymentUsecase)
	paymentController.Route()

	callbackHandler := controller.NewMidtransCallbackHandler(api, paymentUsecase)
	callbackHandler.Route()

	// Log server start
	log.Println("Server running on port 8080")

	// Run server
	r.Run(":8080")
}
