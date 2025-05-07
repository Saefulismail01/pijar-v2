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
	"pijar/repository"
	"pijar/usecase"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	engine      *gin.Engine
	server      *http.Server
	db          *sql.DB
	host        string
	port        string
	dailyGoalUC usecase.DailyGoalUseCase
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

	dailyGoalRepo := repository.NewDailyGoalsRepository(db)
	dailyGoalUC := usecase.NewGoalUseCase(dailyGoalRepo)

	host := fmt.Sprintf("%s:%s", cfg.APIHost, cfg.APIPort)

	engine := gin.New()
	engine.Use(gin.Recovery())

	return &Server{
		engine:      engine,
		db:          db,
		host:        host,
		dailyGoalUC: dailyGoalUC,
	}
}

func (s *Server) initRoute() {
	// initialize Daily Goals Controller
	dailyGoalController := controller.NewGoalController(
		s.dailyGoalUC,
		s.engine.Group("/pijar"),
	)
	dailyGoalController.Route()
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
