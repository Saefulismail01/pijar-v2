package usecase

import (
	"fmt"
	"context"
	"pijar/model"
	"pijar/repository"
	"pijar/utils/service"
)

type SessionUsecase interface {
	StartSession(c context.Context, userID int, userInput string) (string, string, error)
	ContinueSession(c context.Context, userID int, sessionID string, userInput string) (string, error)
	GetSessionHistory(c context.Context, userID int, sessionID string, limit int) ([]model.Message, error)
	GetUserSessions(c context.Context, userID int) ([]model.CoachSession, error)
	DeleteSession(c context.Context, userID int, sessionID string) error
}

type sessionUsecase struct {
	repo repository.CoachSessionRepository
	ai   *service.DeepSeekClient
}

func (u *sessionUsecase) StartSession(c context.Context, userID int, userInput string) (string, string, error) {
	// Buat sesi baru
	sessionID, err := u.repo.CreateSession(c, userID, userInput)
	if err != nil {
		return "", "", fmt.Errorf("gagal membuat sesi: %w", err)
	}

	// Dapatkan konteks percakapan
	ctx, err := u.repo.GetOrCreateConversationContext(c, userID, sessionID)
	if err != nil {
		return "", "", fmt.Errorf("gagal mendapatkan konteks: %w", err)
	}

	// Tambahkan pesan user ke konteks
	ctx.Messages = append(ctx.Messages, model.Message{
		Role:    "user",
		Content: userInput,
	})

	// Dapatkan respons AI dengan konteks
	aiResp, err := u.ai.GetAIResponseWithContext(ctx.Messages)
	if err != nil {
		return "", "", fmt.Errorf("gagal mendapatkan respons AI: %w", err)
	}

	// Tambahkan respons AI ke konteks
	ctx.Messages = append(ctx.Messages, model.Message{
		Role:    "assistant",
		Content: aiResp,
	})

	// Simpan konteks yang diperbarui
	if err := u.repo.SaveConversationContext(c, ctx); err != nil {
		return "", "", fmt.Errorf("gagal menyimpan konteks: %w", err)
	}

	// Update respons ke DB
	if err := u.repo.UpdateSessionResponse(c, sessionID, aiResp); err != nil {
		return "", "", fmt.Errorf("gagal memperbarui respons: %w", err)
	}

	return sessionID, aiResp, nil
}

func (u *sessionUsecase) ContinueSession(c context.Context, userID int, sessionID string, userInput string) (string, error) {
	// Dapatkan konteks percakapan
	ctx, err := u.repo.GetOrCreateConversationContext(c, userID, sessionID)
	if err != nil {
		return "", fmt.Errorf("gagal mendapatkan konteks: %w", err)
	}

	// Tambahkan pesan user ke konteks
	userMessage := model.Message{
		Role:    "user",
		Content: userInput,
	}
	ctx.Messages = append(ctx.Messages, userMessage)

	// Dapatkan respons AI dengan konteks
	aiResp, err := u.ai.GetAIResponseWithContext(ctx.Messages)
	if err != nil {
		return "", fmt.Errorf("gagal mendapatkan respons AI: %w", err)
	}

	// Tambahkan respons AI ke konteks
	aiMessage := model.Message{
		Role:    "assistant",
		Content: aiResp,
	}
	ctx.Messages = append(ctx.Messages, aiMessage)

	// Simpan pesan user dan respons AI ke database
	if err := u.repo.SaveConversation(c, userID, sessionID, userInput, aiResp); err != nil {
		return "", fmt.Errorf("gagal menyimpan percakapan: %w", err)
	}

	// Simpan konteks yang diperbarui
	if err := u.repo.SaveConversationContext(c, ctx); err != nil {
		return "", fmt.Errorf("gagal menyimpan konteks: %w", err)
	}

	return aiResp, nil
}

func (u *sessionUsecase) GetSessionHistory(c context.Context, userID int, sessionID string, limit int) ([]model.Message, error) {
	return u.repo.GetSessionHistory(c, userID, sessionID, limit)
}

func (u *sessionUsecase) GetUserSessions(c context.Context, userID int) ([]model.CoachSession, error) {
	return u.repo.GetUserSessions(c, userID)
}

func (u *sessionUsecase) DeleteSession(c context.Context, userID int, sessionID string) error {
	return u.repo.DeleteSession(c, userID, sessionID)
}

func NewSessionUsecase(repo repository.CoachSessionRepository, aiClient *service.DeepSeekClient) SessionUsecase {
	return &sessionUsecase{
		repo: repo,
		ai:   aiClient,
	}
}
