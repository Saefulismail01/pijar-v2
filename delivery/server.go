package delivery

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"pijar/config"
	"pijar/delivery/controller"
	"pijar/middleware"
	"pijar/repository"
	"pijar/usecase"
	"pijar/utils/service"
	"syscall"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"google.golang.org/api/option"
)

type Server struct {
	// Fitur existing tim
	coachUC        usecase.SessionUsecase
	journalUC      usecase.JournalUsecase
	topicUC        usecase.TopicUsecase
	articleUC      usecase.ArticleUsecase
	dailyGoalUC    usecase.DailyGoalUseCase
	notifUC        usecase.NotificationUseCase
	userRepo       repository.UserRepoInterface
	userUsecase    usecase.UserUsecase
	authUsecase    *usecase.AuthUsecase
	paymentUsecase usecase.PaymentUsecase
	jwtService     service.JwtService
	authMiddleware *middleware.AuthMiddleware
	engine         *gin.Engine
	host           string
	db             *sql.DB
	server         *http.Server

	// Fitur tambahan dari versi Anda
	notificationUC *usecase.NotificationUseCase
	fcmClient      *service.FCMClient
	cron           *cron.Cron
	coachUC        usecase.SessionUsecase
	journalUC      usecase.JournalUsecase
	topicUC        usecase.TopicUsecase
	articleUC      usecase.ArticleUsecase
	dailyGoalUC    usecase.DailyGoalUseCase
	userRepo       repository.UserRepoInterface
	userUsecase    *usecase.UserUsecase
	authUsecase    *usecase.AuthUsecase
	paymentUsecase usecase.PaymentUsecase
	jwtService     service.JwtService
	authMiddleware *middleware.AuthMiddleware
	engine         *gin.Engine
	host           string
	db             *sql.DB
	server         *http.Server
}

func (s *Server) initRoute() {
	rg := s.engine.Group("/pijar")

	// Initialize controllers and setup routes
	controller.NewUserController(rg, s.userUsecase, s.userRepo, s.jwtService, s.authMiddleware).Route()
	controller.NewAuthController(rg, s.jwtService, s.authUsecase).Route()
	// Payment controllers
	controller.NewPaymentController(rg, s.paymentUsecase).Route()
	controller.NewMidtransCallbackHandler(rg, s.paymentUsecase).Route()

	// Feature Coach
	controller.NewSessionHandler(s.coachUC, rg, *s.authMiddleware).Route()

	// feature journal
	controller.NewJournalController(s.journalUC, rg).Route()

	// feature topic
	controller.NewTopicController(s.topicUC, rg).Route()

	// feature articles
	controller.NewArticleController(s.articleUC, rg).Route()

	// feature daily goals
	controller.NewGoalController(s.dailyGoalUC, rg, *s.authMiddleware).Route()
}

func (s *Server) Run() {

	// s.initRoute()
	// if err := s.engine.Run(s.host); err != nil {
	// 	panic(err)
	// }

	s.initRoute()

	s.server = &http.Server{
		Addr:    s.host,
		Handler: s.engine,
	}

	// channel for signal interrupt
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// run server in goroutine
	go func() {
		fmt.Printf("Server running on %s\n", s.host)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(fmt.Errorf("failed to start server: %v", err))
		}
	}()

	// blocking main goroutine until signal received
	<-quit
	fmt.Println("\nShutting down server...")

	// timeout 5 seconds for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// shutdown server
	if err := s.server.Shutdown(ctx); err != nil {
		fmt.Printf("Server forced to shutdown: %v\n", err)
	}

	// close db connection
	if err := s.db.Close(); err != nil {
		fmt.Printf("Error closing database: %v\n", err)
	}

	fmt.Println("Server gracefully stopped 󱠡")

}

func NewServer() *Server {
	// Load environment variables first
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Warning: Error loading .env file: %v\n", err)
	}

	db, cfg, err := config.ConnectDB()
	if err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		return nil
	}

	// Initialize Firebase
	firebaseApp, err := initializeFirebase()
	if err != nil {
		panic(fmt.Errorf("failed to initialize firebase: %v", err))
	}

	// Buat FCM Client
	fcmClient, err := service.NewFCMClient(firebaseApp)
	if err != nil {
		panic(fmt.Errorf("failed to create FCM client: %v", err))
	}

	// Initialize repositories
	userRepo := repository.NewUserRepo(db)
	productRepo := repository.NewProductRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	dailyGoalRepo := repository.NewDailyGoalsRepository(db)
	notifRepo := repository.NewNotificationRepo(db)

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

	// Tambahan usecase notifikasi
	notificationUC := usecase.NewNotificationUseCase(
		notifRepo,
		dailyGoalRepo,
		fcmClient,
	)

	// Setup cron job untuk reminder
	c := cron.New()
	c.AddFunc("0 9,13,17,20 * * *", notificationUC.SendScheduledReminders)
	c.Start()

	// Initialize session repository
	sessionRepo := repository.NewSession(db)

	// Initialize AI coach
	deepseek := service.NewDeepSeekClient(os.Getenv("AI_API"))
	deepseek.SystemPrompt = "You are a professional mental health coach. Your role is to provide empathetic support and guidance. When users need help with decision-making, use the cost-benefit analysis framework to help them think through their options. Maintain a cheerful and supportive tone, but use emoticons sparingly. Keep your responses concise and focused. Avoid repeating yourself. Your goal is to help users gain clarity and make informed decisions about their mental well-being."
	deepseek.Temperature = 0.7
	deepseek.MaxTokens = 500

	coachUsecase := usecase.NewSessionUsecase(sessionRepo, deepseek)

	// Initialize journal
	journalRepo := repository.NewJournalRepository(db)
	journalUsecase := usecase.NewJournalUsecase(journalRepo)

	// Initialize topic and article
	topicRepo := repository.NewTopicRepository(db)
	topicUsecase := usecase.NewTopicUsecase(topicRepo)

	articleRepo := repository.NewArticleRepository(db)
	articleUsecase := usecase.NewArticleUsecase(articleRepo)

	dailyGoalUC := usecase.NewGoalUseCase(dailyGoalRepo, *userRepo, notifRepo, fcmClient)

	engine := gin.Default()
	host := fmt.Sprintf(":%s", cfg.ApiPort)

	return &Server{
		// Fitur existing
		coachUC:        coachUsecase,
		journalUC:      journalUsecase,
		topicUC:        topicUsecase,
		articleUC:      articleUsecase,
		dailyGoalUC:    dailyGoalUC,
		userRepo:       userRepo,
		userUsecase:    userUsecase,
		authUsecase:    authUsecase,
		paymentUsecase: paymentUsecase,
		jwtService:     jwtService,
		authMiddleware: authMiddleware,
		engine:         engine,
		host:           host,
		db:             db,

		notificationUC: notificationUC,
		fcmClient:      fcmClient,
		cron:           c,
	}
}

func (s *Server) initRoute() {
	rg := s.engine.Group("/pijar")

	// Existing routes
	controller.NewUserController(rg, s.userUsecase, s.jwtService, s.authMiddleware).Route()
	controller.NewAuthController(rg, s.jwtService, s.authUsecase).Route()
	controller.NewPaymentController(rg, s.paymentUsecase).Route()
	controller.NewMidtransCallbackHandler(rg, s.paymentUsecase).Route()
	controller.NewSessionHandler(s.coachUC, rg, *s.authMiddleware).Route()
	controller.NewJournalController(s.journalUC, rg, *s.authMiddleware).Route()
	controller.NewTopicController(s.topicUC, rg, *s.authMiddleware).Route()
	controller.NewArticleController(s.articleUC, rg, *s.authMiddleware).Route()
	controller.NewGoalController(s.dailyGoalUC, rg, *s.authMiddleware).Route()

	// Tambahan routes notifikasi
	controller.NewNotificationController(
		*s.notificationUC,
		rg,
		*s.authMiddleware,
	).Route()
}

func initializeFirebase() (*firebase.App, error) {
	opt := option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing app: %v", err)
	}
	return app, nil
}

// Fungsi Run tetap sama dengan versi tim
func (s *Server) Run() {
	s.initRoute()

	s.server = &http.Server{
		Addr:    s.host,
		Handler: s.engine,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Printf("Server running on %s\n", s.host)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(fmt.Errorf("failed to start server: %v", err))
		}
	}()

	<-quit
	fmt.Println("\nShutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		fmt.Printf("Server forced to shutdown: %v\n", err)
	}

	if err := s.db.Close(); err != nil {
		fmt.Printf("Error closing database: %v\n", err)
	}

	fmt.Println("Server gracefully stopped 󱠡")
}
