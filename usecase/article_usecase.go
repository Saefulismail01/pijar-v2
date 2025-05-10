package usecase

import (
	"context"
	"fmt"
	"pijar/model"
	"pijar/repository"
	"pijar/utils/service"

	"github.com/gin-gonic/gin"
)

type ArticleUsecase interface {
	CreateArticle(c *gin.Context, preferences []string) error
	GenerateArticle(ctx context.Context, topicID int) ([]model.Article, error)
	GetAllArticles(ctx context.Context, page int) (*model.ArticleResponse, error)
	GetAllArticlesWithoutPagination(ctx context.Context) ([]model.Article, error)
	GetArticleByID(ctx context.Context, id int) (*model.Article, error)
	// GetArticleByTitle(ctx context.Context, title string) (*model.Article, error)
	SearchArticlesByTitle(ctx context.Context, title string) ([]model.Article, error)
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

// CreateArticle handles generating and creating multiple articles from preferences
func (u *articleUsecase) CreateArticle(c *gin.Context, preferences []string) error {
	// Generate articles using Deepseek
	generatedArticles, err := service.GenerateArticles(c, preferences)
	if err != nil {
		return fmt.Errorf("failed to generate articles: %w", err)
	}

	// Start transaction and save articles
	ctx := context.Background()
	tx, err := u.articleRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer u.articleRepo.RollbackTx(tx)

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

// GenerateArticle handles the generation and creation of articles based on a topic ID
func (u *articleUsecase) GenerateArticle(ctx context.Context, topicID int) ([]model.Article, error) {
	// Log the article generation request
	fmt.Printf("Article generation requested for topic ID: %d\n", topicID)

	// Get the gin context if available
	ginCtx, isGinCtx := ctx.(*gin.Context)

	// First, check if the topic exists using the repository in context
	var topicExists bool
	var topicPreference string

	// Try to get topicRepo from context if available
	topicRepo, ok := ctx.Value("topicRepo").(repository.TopicUserRepository)
	if ok {
		// Get topic by ID to validate it exists
		topics, err := topicRepo.GetTopicByID(ctx, topicID)
		if err != nil {
			return nil, fmt.Errorf("failed to get topic: %w", err)
		}

		if len(topics) > 0 {
			topicExists = true
			topicPreference = topics[0].Preference

			// If we have a gin context, set the user ID
			if isGinCtx {
				ginCtx.Set("user_id", topics[0].UserID)
			}
		}
	}

	// If we couldn't verify the topic through the context, we'll rely on the repository check
	if !topicExists {
		// We'll let the repository handle the topic existence check
		// Start transaction
		tx, err := u.articleRepo.BeginTx(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer u.articleRepo.RollbackTx(tx)

		// Generate articles - the repository will check if the topic exists
		articles, err := u.articleRepo.GenerateArticle(ctx, tx, topicID)
		if err != nil {
			return nil, err
		}

		// Commit the transaction
		if err := u.articleRepo.CommitTx(tx); err != nil {
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}

		return articles, nil
	}

	// If we get here, we've confirmed the topic exists and have its preference
	fmt.Printf("Topic found with preference: %s\n", topicPreference)

	// Start transaction
	tx, err := u.articleRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer u.articleRepo.RollbackTx(tx)

	// Generate articles using the topic's preference
	articles, err := u.articleRepo.GenerateArticle(ctx, tx, topicID)
	if err != nil {
		return nil, err
	}

	// Commit the transaction
	if err := u.articleRepo.CommitTx(tx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return articles, nil
}

func (u *articleUsecase) GetAllArticles(ctx context.Context, page int) (*model.ArticleResponse, error) {
	const limit = 3

	if page < 1 {
		page = 1
	}

	articles, totalItems, err := u.articleRepo.GetPaginatedArticles(ctx, page, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get articles: %w", err)
	}

	totalPages := int(totalItems) / limit
	if int(totalItems)%limit != 0 {
		totalPages++
	}

	response := &model.ArticleResponse{
		Articles: articles,
		Pagination: model.Pagination{
			CurrentPage: page,
			TotalPages:  totalPages,
			TotalItems:  totalItems,
			Limit:       limit,
		},
	}

	return response, nil
}

func (u *articleUsecase) GetAllArticlesWithoutPagination(ctx context.Context) ([]model.Article, error) {
	articles, err := u.articleRepo.GetAllArticles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all articles: %w", err)
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

// func (u *articleUsecase) GetArticleByTitle(ctx context.Context, title string) (*model.Article, error) {
// 	return u.articleRepo.GetArticleByTitle(ctx, title)
// }

// SearchArticlesByTitle finds articles with titles similar to the given title
func (u *articleUsecase) SearchArticlesByTitle(ctx context.Context, title string) ([]model.Article, error) {
	// First try to get exact match (case insensitive)
	article, err := u.articleRepo.GetArticleByTitle(ctx, title)
	if err == nil && article != nil {
		return []model.Article{*article}, nil
	}

	// If no exact match, search for similar titles
	return u.articleRepo.SearchArticlesByTitle(ctx, title)
}

func (u *articleUsecase) DeleteArticle(ctx context.Context, id int) error {
	err := u.articleRepo.DeleteArticle(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete article: %w", err)
	}
	return nil
}
