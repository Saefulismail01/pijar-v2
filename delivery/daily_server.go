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
	"github.com/robfig/cron/v3"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/api/option"
)

type Server struct {
	engine         *gin.Engine
	server         *http.Server
	db             *sql.DB
	host           string
	port           string
	dailyGoalUC    usecase.DailyGoalUseCase
	jwtService     service.JwtService
	notificationUC *usecase.NotificationUseCase // Tambahkan ini
	userRepo       *repository.UserRepo
}

func NewServer() *Server {
	cfg, err := config.NewConfig()
	if err != nil {
		panic(fmt.Errorf("failed to read config: %v", err))
	}

	db, err := config.ConnectDB(cfg.DBConfig)
	if err != nil {
		panic(fmt.Errorf("failed to connect db: %v", err))
	}
	// Inisialisasi Firebase
	firebaseApp, err := initializeFirebase()
	if err != nil {
		panic(fmt.Errorf("failed to initialize firebase: %v", err))
	}

	// Buat FCM Client
	fcmClient, err := service.NewFCMClient(firebaseApp)
	if err != nil {
		panic(fmt.Errorf("failed to create FCM client: %v", err))
	}

	// Inisialisasi Repository dan Usecase
	userRepo := repository.NewUserRepo(db)
	dailyGoalRepo := repository.NewDailyGoalsRepository(db)
	dailyGoalUC := usecase.NewGoalUseCase(dailyGoalRepo, *userRepo, fcmClient)
	notificationUC := usecase.NewNotificationUseCase(
		userRepo,
		dailyGoalRepo,
		fcmClient,
	)

	// Setup cron job untuk reminder
	c := cron.New()
	c.AddFunc("0 9,13,17,20 * * *", notificationUC.SendScheduledReminders)
	c.Start()

	host := fmt.Sprintf("%s:%s", cfg.APIHost, cfg.APIPort)

	engine := gin.New()
	engine.Use(gin.Recovery())

	return &Server{
		engine:         engine,
		db:             db,
		host:           host,
		dailyGoalUC:    dailyGoalUC,
		notificationUC: notificationUC,
		userRepo:       userRepo,
	}
}

func (s *Server) initRoute() {
	// add swagger route in /pijar
	pijarGroup := s.engine.Group("/pijar")
	{
		pijarGroup.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

		authMiddleware := middleware.NewAuthMiddleware(s.jwtService)

		dailyGoalController := controller.NewGoalController(
			s.dailyGoalUC,
			pijarGroup,      // Gunakan group yang sama
			*authMiddleware, // Tambahkan auth middleware
		)
		dailyGoalController.Route()
	}
	notificationCtrl := controller.NewNotificationController(
		*s.userRepo,
		*s.notificationUC,
		pijarGroup,
	)
	notificationCtrl.Route()
}

func initializeFirebase() (*firebase.App, error) {
	// Gunakan credentials dari environment variable
	opt := option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing app: %v", err)
	}
	return app, nil
}

func (s *Server) Run() {
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
