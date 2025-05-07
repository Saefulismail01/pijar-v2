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
)

type Server struct {
	coachUC   usecase.SessionUsecase
	journalUC usecase.JournalUsecase
	engine    *gin.Engine
	host      string
}

func (s *Server) initRoute() {
	rg := s.engine.Group("/pijar")
	controller.NewSessionHandler(s.coachUC, rg).Route()
	controller.NewJournalController(s.journalUC, rg).Route()
}

func (s *Server) Run() {
	s.initRoute()
	if err := s.engine.Run(s.host); err != nil {
		panic(err)
	}
}

func NewServer() *Server {

	db, cfg, err := config.ConnectDB()
	if err != nil {
		fmt.Printf("err: %v", err)
		return nil
	}

	//ini fitur ai-couch
	sessionRepo := repository.NewSession(db)
	deepseek := service.NewDeepSeekClient(os.Getenv("AI_API"))
	deepseek.SystemPrompt = "You are a professional mental health coach. Your role is to provide empathetic support and guidance. When users need help with decision-making, use the cost-benefit analysis framework to help them think through their options. Maintain a cheerful and supportive tone, but use emoticons sparingly. Keep your responses concise and focused. Avoid repeating yourself. Your goal is to help users gain clarity and make informed decisions about their mental well-being."

	deepseek.Temperature = 0.7
	deepseek.MaxTokens = 500

	coachUsecase := usecase.NewSessionUsecase(sessionRepo, deepseek)

	// Initialize journal repository and usecase
	journalRepo := repository.NewJournalRepository(db)
	journalUsecase := usecase.NewJournalUsecase(journalRepo)

	engine := gin.Default()
	host := fmt.Sprintf(":%s", cfg.ApiPort)

	return &Server{
		coachUC:   coachUsecase,
		journalUC: journalUsecase,
		engine:    engine,
		host:      host,
	}
}
