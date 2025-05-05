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
	// Inisialisasi DeepSeekClient dengan personalisasi
	deepseek := service.NewDeepSeekClient(os.Getenv("AI_API")).
		WithSystemPrompt(`Kamu adalah AI coach profesional dan sahabat yang membantu generasi muda dan akademisi dalam pengembangan karir dan skill. 
		Berikan saran yang spesifik, praktis, dan dapat ditindaklanjuti. 
		Gunakan bahasa yang ramah, penuh empati, dan mudah dipahami. 
		Selalu berikan contoh konkret dan relevan dengan konteks pengguna. 
		Jika dia bertanya terkait opsi maka jawaban anda dengan framewrok Cost Benefit Analysis secara mendalam.`).
		WithTemperature(0.7).
		WithMaxTokens(2000)

	coachUsecase := usecase.NewSessionUsecase(sessionRepo, deepseek)

	engine := gin.Default()
	host := fmt.Sprintf(":%s", cfg.ApiPort)

	return &Server{
		coachUC: coachUsecase,
		engine:  engine,
		host:    host,
	}
}
