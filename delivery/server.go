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

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
)

type Server struct {
	coachUC        usecase.SessionUsecase
	journalUC      usecase.JournalUsecase
	topicUC        usecase.TopicUsecase
	articleUC      usecase.ArticleUsecase
	dailyGoalUC    usecase.DailyGoalUseCase
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
}

func (s *Server) initRoute() {
	rg := s.engine.Group("/pijar")

	controller.NewUserController(rg, s.userUsecase, s.jwtService, s.authMiddleware).Route()
	controller.NewAuthController(rg, s.jwtService, *s.authUsecase).Route()
	controller.NewPaymentController(rg, s.paymentUsecase, *s.authMiddleware).Route()
	controller.NewMidtransCallbackHandler(rg, s.paymentUsecase).Route()
	controller.NewSessionHandler(s.coachUC, rg, *s.authMiddleware).Route()
	controller.NewJournalController(s.journalUC, rg, *s.authMiddleware).Route()
	controller.NewTopicController(s.topicUC, rg, *s.authMiddleware).Route()
	controller.NewArticleController(s.articleUC, rg, *s.authMiddleware).Route()
	controller.NewGoalController(s.dailyGoalUC, rg, *s.authMiddleware).Route()
}

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

func NewServer() *Server {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Warning: Error loading .env file: %v\n", err)
	}

	db, cfg, err := config.ConnectDB()
	if err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		return nil
	}

	// Initialize database repositories
	userRepo := repository.NewUserRepo(db)
	productRepo := repository.NewProductRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)

	// Initialize service dependencies
	jwtSecret := os.Getenv("JWT_SECRET")
	appName := os.Getenv("APP_NAME")
	jwtExpiryStr := os.Getenv("JWT_EXPIRY")
	jwtExpiry, err := time.ParseDuration(jwtExpiryStr)
	if err != nil {
		jwtExpiry = 24 * time.Hour
		fmt.Printf("Warning: Could not parse JWT_EXPIRY value '%s', using default of 24h: %v\n", jwtExpiryStr, err)
	}
	jwtService := service.NewJwtService(jwtSecret, appName, jwtExpiry)
	restyClient := resty.New()
	midtransService := service.NewMidtransService(restyClient)

	// Initialize middleware components
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	// Initialize usecase layer components
	userUsecase := usecase.NewUserUsecase(userRepo)
	authUsecase := usecase.NewAuthUsecase(userRepo, jwtService)
	paymentUsecase := usecase.NewPaymentUsecase(midtransService, productRepo, transactionRepo, userRepo)

	// Initialize session management components
	sessionRepo := repository.NewSession(db)

	// Initialize AI coach service with custom prompt and settings
	deepseek := service.NewDeepSeekClient(os.Getenv("AI_API"))
	deepseek.SystemPrompt = "You are a professional mental health coach. Your role is to provide empathetic support and guidance. When users need help with decision-making, use the cost-benefit analysis framework to help them think through their options. Maintain a cheerful and supportive tone, but use emoticons sparingly. Keep your responses concise and focused. Avoid repeating yourself. Your goal is to help users gain clarity and make informed decisions about their mental well-being."
	deepseek.Temperature = 0.7
	deepseek.MaxTokens = 500

	coachUsecase := usecase.NewSessionUsecase(sessionRepo, deepseek)

	// Initialize journal management components
	journalRepo := repository.NewJournalRepository(db)
	journalUsecase := usecase.NewJournalUsecase(journalRepo)

	// Initialize topic management components
	topicRepo := repository.NewTopicRepository(db)
	topicUsecase := usecase.NewTopicUsecase(topicRepo)

	// Initialize article management components
	articleRepo := repository.NewArticleRepository(db)
	articleUsecase := usecase.NewArticleUsecase(articleRepo)

	// Initialize daily goals management components
	dailyGoalRepo := repository.NewDailyGoalsRepository(db)
	dailyGoalUC := usecase.NewGoalUseCase(dailyGoalRepo)

	engine := gin.Default()
	host := fmt.Sprintf(":%s", cfg.ApiPort)

	return &Server{
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
	}
}
