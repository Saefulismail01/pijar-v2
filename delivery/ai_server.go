package delivery

import (
	"fmt"
	"os"
	"pijar/config"
	"pijar/delivery/controller"
	"pijar/repository"
	"pijar/usecase"
	"pijar/utils/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type Server struct {
	coachUC   usecase.SessionUsecase
	journalUC usecase.JournalUsecase
	topicUC   usecase.TopicUsecase
	articleUC usecase.ArticleUsecase
	engine    *gin.Engine
	host      string
}

func (s *Server) initRoute() {
	rg := s.engine.Group("/pijar")

	// Feature Coach
	controller.NewSessionHandler(s.coachUC, rg).Route()

	// feature journal
	controller.NewJournalController(s.journalUC, rg).Route()

	// feature topic
	controller.NewTopicController(s.topicUC, rg).Route()

	// feature articles
	controller.NewArticleController(s.articleUC, rg).Route()
}

func (s *Server) Run() {
	s.initRoute()
	if err := s.engine.Run(s.host); err != nil {
		panic(err)
	}
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

	// Get AI API key from environment
	aiApiKey := os.Getenv("AI_API")
	if aiApiKey == "" {
		fmt.Println("Warning: AI_API environment variable is not set")
	}

	// Initialize AI coach
	sessionRepo := repository.NewSession(db)
	deepseek := service.NewDeepSeekClient(aiApiKey)
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

	engine := gin.Default()
	host := fmt.Sprintf(":%s", cfg.ApiPort)

	return &Server{
		coachUC:   coachUsecase,
		journalUC: journalUsecase,
		topicUC:   topicUsecase,
		articleUC: articleUsecase,
		engine:    engine,
		host:      host,
	}
}
