package usecase

import (
	"context"
	"pijar/model"
	"pijar/repository"
	"pijar/utils/service"
)

type SessionUsecase interface {
	StartSession(ctx context.Context, userID int, userInput string) (string, error)
	GetSessionByUserID(ctx context.Context, userID int) ([]model.CoachSession, error)
	DeleteSessionByUserID(ctx context.Context, userID int) error
}

type sessionUsecase struct {
	repo repository.CoachRepository
	ai   *service.DeepSeekClient
}

func (u *sessionUsecase) StartSession(ctx context.Context, userID int, userInput string) (string, error) {

	// Simpan input awal
	sessionID, err := u.repo.CreateSession(ctx, userID, userInput)
	if err != nil {
		return "", err
	}

	// Dapatkan respons AI
	aiResp, err := u.ai.GetAIResponse(userInput)
	if err != nil {
		return "", err
	}

	// Update respons ke DB
	u.repo.UpdateSessionResponse(ctx, sessionID, aiResp)

	return aiResp, nil
}

func (u *sessionUsecase) GetSessionByUserID(ctx context.Context, userID int) ([]model.CoachSession, error) {
	return u.repo.GetSessionByUserID(ctx, userID)
}

func (u *sessionUsecase) DeleteSessionByUserID(ctx context.Context, userID int) error {
	return u.repo.DeleteSessionByUserID(ctx, userID)
}

func NewSessionUsecase(repo repository.CoachRepository, aiClient *service.DeepSeekClient) SessionUsecase {
	return &sessionUsecase{repo: repo, ai: aiClient}
}
