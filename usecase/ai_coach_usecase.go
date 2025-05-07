package usecase

import (
	"fmt"
	"pijar/model"
	"pijar/repository"
	"pijar/utils/service"
)

type SessionUsecase interface {
	StartSession(userID int, userInput string) (string, string, error)
	ContinueSession(userID int, sessionID string, userInput string) (string, error)
	GetSessionHistory(userID int, sessionID string, limit int) ([]model.Message, error)
	GetUserSessions(userID int) ([]model.CoachSession, error)
}

type sessionUsecase struct {
	repo repository.CouchRepository
	ai   *service.DeepSeekClient
}

func (u *sessionUsecase) StartSession(userID int, userInput string) (string, string, error) {
	// Buat sesi baru
	sessionID, err := u.repo.CreateSession(userID, userInput)
	if err != nil {
		return "", "", fmt.Errorf("gagal membuat sesi: %w", err)
	}

	// Dapatkan konteks percakapan
	ctx, err := u.repo.GetOrCreateConversationContext(userID, sessionID)
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
	if err := u.repo.SaveConversationContext(ctx); err != nil {
		return "", "", fmt.Errorf("gagal menyimpan konteks: %w", err)
	}

	// Update respons ke DB
	if err := u.repo.UpdateSessionResponse(sessionID, aiResp); err != nil {
		return "", "", fmt.Errorf("gagal memperbarui respons: %w", err)
	}

	return sessionID, aiResp, nil
}

func (u *sessionUsecase) ContinueSession(userID int, sessionID string, userInput string) (string, error) {
	// Dapatkan konteks percakapan
	ctx, err := u.repo.GetOrCreateConversationContext(userID, sessionID)
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
	if err := u.repo.SaveConversation(userID, sessionID, userInput, aiResp); err != nil {
		return "", fmt.Errorf("gagal menyimpan percakapan: %w", err)
	}

	// Simpan konteks yang diperbarui
	if err := u.repo.SaveConversationContext(ctx); err != nil {
		return "", fmt.Errorf("gagal menyimpan konteks: %w", err)
	}

	return aiResp, nil
}

func (u *sessionUsecase) GetSessionHistory(userID int, sessionID string, limit int) ([]model.Message, error) {
	return u.repo.GetSessionHistory(userID, sessionID, limit)
}

func (u *sessionUsecase) GetUserSessions(userID int) ([]model.CoachSession, error) {
	return u.repo.GetUserSessions(userID)
}

func NewSessionUsecase(repo repository.CouchRepository, aiClient *service.DeepSeekClient) SessionUsecase {
	return &sessionUsecase{
		repo: repo,
		ai:   aiClient,
	}
}
