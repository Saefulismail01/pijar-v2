package usecase

import (
	"context"
	"fmt"
	"pijar/model"
	"pijar/model/dto"
	"pijar/repository"
	"pijar/utils"

	"github.com/gin-gonic/gin"
)

type ArticleUsecase interface {
	CreateArticle(c *gin.Context, preferences []string) error
	GenerateArticle(ctx context.Context, topicID int) ([]model.Article, error)
	GetAllArticles(ctx context.Context) ([]model.Article, error)
	GetArticleByID(ctx context.Context, id int) (*model.Article, error)
	GetArticleByTitle(ctx context.Context, title string) (*model.Article, error)
	UpdateArticle(ctx context.Context, articleDto *dto.ArticleDto, id int) error
	DeleteArticle(ctx context.Context, id int) error
}

type articleUsecase struct {
	articleRepo repository.ArticleRepository
}

func NewArticleUsecase(articleRepo repository.ArticleRepository) ArticleUsecase {
	return &articleUsecase{
		articleRepo: articleRepo,
	}
}

// CreateArticle menangani generate dan create multiple articles dari preferences
func (u *articleUsecase) CreateArticle(c *gin.Context, preferences []string) error {
	// Generate articles menggunakan Deepseek
	generatedArticles, err := utils.GenerateArticles(c, preferences)
	if err != nil {
		return fmt.Errorf("failed to generate articles: %w", err)
	}

	// Start transaction
	ctx := context.Background()
	tx, err := u.articleRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer u.articleRepo.RollbackTx(tx)

	// Save setiap artikel dalam transaction
	for _, genArticle := range generatedArticles {
		article := &model.Article{
			Title:   genArticle.Title,
			Content: genArticle.Content,
			Source:  genArticle.Source,
			IDTopic: genArticle.TopicID,
		}

		err = u.articleRepo.CreateArticle(ctx, tx, article)
		if err != nil {
			return fmt.Errorf("failed to create article: %w", err)
		}
	}

	// Commit transaction
	if err = u.articleRepo.CommitTx(tx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GenerateArticle menangani generate dan create single article dari topicID
func (u *articleUsecase) GenerateArticle(ctx context.Context, topicID int) ([]model.Article, error) {
	tx, err := u.articleRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	articles, err := u.articleRepo.GenerateArticle(ctx, tx, topicID)
	if err != nil {
		return nil, err
	}

	if err := u.articleRepo.CommitTx(tx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return articles, nil
}

func (u *articleUsecase) GetAllArticles(ctx context.Context) ([]model.Article, error) {
	articles, err := u.articleRepo.GetAllArticles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get articles: %w", err)
	}
	return articles, nil
}

func (u *articleUsecase) GetArticleByID(ctx context.Context, id int) (*model.Article, error) {
	article, err := u.articleRepo.GetArticleByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get article: %w", err)
	}
	return article, nil
}

func (u *articleUsecase) GetArticleByTitle(ctx context.Context, title string) (*model.Article, error) {
	article, err := u.articleRepo.GetArticleByTitle(ctx, title)
	if err != nil {
		return nil, fmt.Errorf("failed to get article by title: %w", err)
	}
	return article, nil
}

func (u *articleUsecase) UpdateArticle(ctx context.Context, articleDto *dto.ArticleDto, id int) error {
	// Check if article exists
	existingArticle, err := u.articleRepo.GetArticleByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get article: %w", err)
	}

	// Update article fields
	existingArticle.Title = articleDto.Title
	existingArticle.Content = articleDto.Content
	existingArticle.Source = articleDto.Source
	existingArticle.IDTopic = articleDto.IDTopic

	// Save updates
	err = u.articleRepo.UpdateArticle(ctx, existingArticle)
	if err != nil {
		return fmt.Errorf("failed to update article: %w", err)
	}

	return nil
}

func (u *articleUsecase) DeleteArticle(ctx context.Context, id int) error {
	err := u.articleRepo.DeleteArticle(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete article: %w", err)
	}
	return nil
}
