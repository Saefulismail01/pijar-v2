package delivery

import (
	"fmt"
	"os"
	"pijar/config"
	"pijar/delivery/controller"
	"pijar/usecase"
	"pijar/repository"
	"pijar/utils/service"

	"github.com/gin-gonic/gin"
)

type Server struct {
	coachUC usecase.SessionUsecase
	engine  *gin.Engine
	host    string
}

func (s *Server) initRoute() {
	rg := s.engine.Group("/pijar")
	controller.NewSessionHandler(s.coachUC, rg).Route()
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
	coachUsecase := usecase.NewSessionUsecase(sessionRepo, deepseek)

	engine := gin.Default()
	host := fmt.Sprintf(":%s", cfg.ApiPort)

	return &Server{
		coachUC: coachUsecase,
		engine:  engine,
		host:    host,
	}
}