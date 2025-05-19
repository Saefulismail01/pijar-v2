package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"pijar/model"
	"pijar/repository"
	"pijar/utils/service"
)

type JournalAIUsecase interface {
	AnalyzeJournal(ctx context.Context, req *model.AnalysisRequest) (*model.AnalysisResponse, error)
	GetJournalAnalysis(ctx context.Context, journalID int, userID int) (*model.AnalysisResponse, error)
	ReanalyzeJournal(ctx context.Context, journalID int, userID int) (*model.AnalysisResponse, error)
	GetUserAnalyses(ctx context.Context, userID int, limit int) ([]*model.JournalAnalysis, error)
	GetAnalysisWithJournal(ctx context.Context, userID int, limit int) ([]map[string]interface{}, error)
	GenerateTrendAnalysis(ctx context.Context, userID int, periodType string, days int) (*model.TrendResponse, error)
	GetTrendHistory(ctx context.Context, userID int, periodType string) ([]*model.TrendAnalysis, error)
	GetSentimentChart(ctx context.Context, userID int, days int) ([]map[string]interface{}, error)
}

type journalAIUsecase struct {
	repo         repository.JournalAnalysisRepository
	journalRepo  repository.JournalRepository
	aiService    *service.JournalAnalysisService
}

func NewJournalAIUsecase(repo repository.JournalAnalysisRepository, journalRepo repository.JournalRepository, aiService *service.JournalAnalysisService) JournalAIUsecase {
	return &journalAIUsecase{
		repo:        repo,
		journalRepo: journalRepo,
		aiService:   aiService,
	}
}

func (u *journalAIUsecase) AnalyzeJournal(ctx context.Context, req *model.AnalysisRequest) (*model.AnalysisResponse, error) {
	// Use the journal AI service to analyze the journal entry
	return u.aiService.AnalyzeJournalEntry(req)
}

func (u *journalAIUsecase) GetJournalAnalysis(ctx context.Context, journalID int, userID int) (*model.AnalysisResponse, error) {
	// First, check if analysis already exists
	analysis, err := u.repo.GetByJournalID(journalID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to check for existing analysis: %w", err)
	}

	// If analysis exists, return it
	if analysis != nil {
		return &model.AnalysisResponse{
			JournalAnalysis: analysis,
		}, nil
	}

	// Get the journal content
	journal, err := u.journalRepo.FindByID(ctx, journalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get journal: %w", err)
	}

	// Verify that the journal belongs to the user
	if journal.UserID != userID {
		return nil, errors.New("unauthorized: journal does not belong to user")
	}

	// Create analysis request
	req := &model.AnalysisRequest{
		JournalID: journalID,
		UserID:    userID,
		Title:     journal.Judul,
		Content:   journal.Isi,
		Feeling:   journal.Perasaan,
	}

	return u.AnalyzeJournal(ctx, req)
}

func (u *journalAIUsecase) ReanalyzeJournal(ctx context.Context, journalID int, userID int) (*model.AnalysisResponse, error) {
	analysis, err := u.repo.GetByJournalID(journalID)
	if err != nil {
		return nil, err
	}

	// Verify ownership - convert userID to int for comparison
	if int(analysis.UserID) != userID {
		return nil, errors.New("unauthorized access to journal analysis")
	}

	// TODO: Implement reanalysis logic using AI client
	return &model.AnalysisResponse{JournalAnalysis: analysis}, nil
}

func (u *journalAIUsecase) GetUserAnalyses(ctx context.Context, userID int, limit int) ([]*model.JournalAnalysis, error) {
	// First, get all journals for the user
	journals, err := u.journalRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user journals: %w", err)
	}

	var analyses []*model.JournalAnalysis

	// For each journal, get its analysis
	for _, journal := range journals {
		// Get the analysis (this will create one if it doesn't exist)
		resp, err := u.GetJournalAnalysis(ctx, journal.ID, userID)
		if err != nil {
			// Log the error but continue with other journals
			log.Printf("Warning: failed to get/analyze journal %d: %v", journal.ID, err)
			continue
		}

		analyses = append(analyses, resp.JournalAnalysis)

		// If we've reached the limit, break
		if limit > 0 && len(analyses) >= limit {
			break
		}
	}

	return analyses, nil
}

func (u *journalAIUsecase) GetAnalysisWithJournal(ctx context.Context, userID int, limit int) ([]map[string]interface{}, error) {
	return u.repo.GetAnalysisWithJournal(userID, limit)
}

func (u *journalAIUsecase) GenerateTrendAnalysis(ctx context.Context, userID int, periodType string, days int) (*model.TrendResponse, error) {
	// TODO: Implement trend analysis logic
	return nil, errors.New("not implemented")
}

func (u *journalAIUsecase) GetTrendHistory(ctx context.Context, userID int, periodType string) ([]*model.TrendAnalysis, error) {
	return u.repo.GetTrendsByUserID(userID, periodType)
}

func (u *journalAIUsecase) GetSentimentChart(ctx context.Context, userID int, days int) ([]map[string]interface{}, error) {
	return u.repo.GetSentimentTrend(userID, days)
}
