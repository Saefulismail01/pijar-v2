package usecase

import (
	"pijar/model"
	"pijar/repository"
	"pijar/utils/service"
)

type SessionUsecase interface {
	StartSession(userID int, userInput string) (string, error)
	GetSessionByUserID(userID int) ([]model.CoachSession, error)
}

type sessionUsecase struct {
	repo repository.CouchRepository
	ai   *service.DeepSeekClient
}

func (u *sessionUsecase) StartSession(userID int, userInput string) (string, error) {

	// Simpan input awal
	sessionID, err := u.repo.CreateSession(userID, userInput)
	if err != nil {
		return "", err
	}

	// Dapatkan respons AI
	aiResp, err := u.ai.GetAIResponse(userInput)
	if err != nil {
		return "", err
	}

	// Update respons ke DB
	u.repo.UpdateSessionResponse(sessionID, aiResp)

	return aiResp, nil
}

func (u *sessionUsecase) GetSessionByUserID(userID int) ([]model.CoachSession, error) {
	return u.repo.GetSessionByUserID(userID)
}

func NewSessionUsecase(repo repository.CouchRepository, aiClient *service.DeepSeekClient) SessionUsecase {
	return &sessionUsecase{repo: repo, ai: aiClient}
}
