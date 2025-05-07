// package delivery

// import (
// 	"context"
// 	"database/sql"
// 	"fmt"
// 	"net/http"
// 	"os"
// 	"os/signal"
// 	"pijar/delivery/controller"
// 	"pijar/middleware"
// 	"pijar/repository"
// 	"pijar/usecase"
// 	"syscall"
// 	"time"

// 	"github.com/gin-gonic/gin"
// )

// type Server struct {
// 	engine *gin.Engine
// 	server *http.Server
// }

// func NewServer(db *sql.DB) (*Server, error) {
// 	// Initialize Gin engine with logger and recovery middleware
// 	engine := gin.New()
// 	engine.Use(gin.Logger(), gin.Recovery())

// 	// Get port from environment variable or use default
// 	port := os.Getenv("API_PORT")
// 	if port == "" {
// 		port = "8080"
// 	}
// 	host := fmt.Sprintf(":%s", port)

// 	// Initialize repositories
// 	articleRepo := repository.NewArticleRepository(db)
// 	topicRepo := repository.NewTopicRepository(db)

// 	// Initialize usecases
// 	articleUC := usecase.NewArticleUsecase(articleRepo)
// 	topicUC := usecase.NewTopicUsecase(topicRepo)

// 	// Initialize controllers
// 	articleCtrl := controller.NewArticleController(articleUC)
// 	topicCtrl := controller.NewTopicController(topicUC)

// 	// Create router groups
// 	publicRouter := engine.Group("/api")
// 	protectedRouter := engine.Group("/api")

// 	// Apply middleware to protected routes
// 	protectedRouter.Use(middleware.DeepseekAuthMiddleware())

// 	// Register routes
// 	articleCtrl.RegisterRoutes(publicRouter, protectedRouter)
// 	topicCtrl.RegisterRoutes(publicRouter, protectedRouter)

// 	return &Server{
// 		engine: engine,
// 		server: &http.Server{
// 			Addr:         host,
// 			Handler:      engine,
// 			ReadTimeout:  15 * time.Second,
// 			WriteTimeout: 15 * time.Second,
// 			IdleTimeout:  60 * time.Second,
// 		},
// 	}, nil
// }

// func (s *Server) Run() error {
// 	// Channel to capture OS signals
// 	quit := make(chan os.Signal, 1)
// 	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

// 	// Run server in a goroutine
// 	go func() {
// 		fmt.Printf("Server running on http://localhost%s\n", s.server.Addr)
// 		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
// 			fmt.Printf("Failed to start server: %v\n", err)
// 			quit <- os.Interrupt // Trigger shutdown on error
// 		}
// 	}()

// 	// Wait for shutdown signal
// 	<-quit
// 	fmt.Println("\nShutting down server...")

// 	// Context with timeout for shutdown process
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	// Attempt to gracefully shut down the server
// 	if err := s.server.Shutdown(ctx); err != nil {
// 		return fmt.Errorf("server forced to shutdown: %v", err)
// 	}

// 	fmt.Println("Server gracefully stopped")
// 	return nil
// }

package delivery

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"pijar/delivery/controller"
	"pijar/middleware"
	"pijar/mock"
	"pijar/repository"
	"pijar/usecase"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	engine *gin.Engine
	server *http.Server
}

func NewServer(db *sql.DB) (*Server, error) {
	// Initialize Gin engine with logger and recovery middleware
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	// Get port from environment variable or use default
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}
	host := fmt.Sprintf(":%s", port)

	// Initialize repositories
	topicRepo := repository.NewTopicRepository(db)

	// Initialize usecases
	topicUC := usecase.NewTopicUsecase(topicRepo)

	// Initialize controllers
	topicCtrl := controller.NewTopicController(topicUC)

	// Create router groups
	publicRouter := engine.Group("/api")
	protectedRouter := engine.Group("/api")

	// Apply middleware to protected routes
	protectedRouter.Use(middleware.DeepseekAuthMiddleware())

	// Register topic routes (using real database)
	topicCtrl.RegisterRoutes(publicRouter, protectedRouter)

	// Register mock article routes
	mockArticleHandler := mock.NewArticleMockHandler()
	mockArticleHandler.RegisterRoutes(engine)
	fmt.Println("Using mock implementation for Article endpoints")

	return &Server{
		engine: engine,
		server: &http.Server{
			Addr:         host,
			Handler:      engine,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}, nil
}

func (s *Server) Run() error {
	// Channel to capture OS signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Run server in a goroutine
	go func() {
		fmt.Printf("Server running on http://localhost%s\n", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Failed to start server: %v\n", err)
			quit <- os.Interrupt // Trigger shutdown on error
		}
	}()

	// Wait for shutdown signal
	<-quit
	fmt.Println("\nShutting down server...")

	// Context with timeout for shutdown process
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Attempt to gracefully shut down the server
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %v", err)
	}

	fmt.Println("Server gracefully stopped")
	return nil
}
