package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"pijar/model"
	"pijar/utils/service"
	"time"

	"github.com/gin-gonic/gin"
)

type ArticleRepository interface {
	CreateArticle(ctx context.Context, tx *sql.Tx, article *model.Article) error
	GenerateArticle(ctx context.Context, tx *sql.Tx, topicID int) ([]model.Article, error)
	GetAllArticles(ctx context.Context) ([]model.Article, error)
	GetPaginatedArticles(ctx context.Context, page, limit int) ([]model.Article, int64, error)
	GetArticleByID(ctx context.Context, id int) (*model.Article, error)
	GetArticleByTitle(ctx context.Context, title string) (*model.Article, error)
	SearchArticlesByTitle(ctx context.Context, title string) ([]model.Article, error)
	//UpdateArticle(ctx context.Context, article *model.Article) error
	DeleteArticle(ctx context.Context, id int) error
	BeginTx(ctx context.Context) (*sql.Tx, error)
	CommitTx(tx *sql.Tx) error
	RollbackTx(tx *sql.Tx) error
}

type articleRepository struct {
	db *sql.DB
}

func NewArticleRepository(db *sql.DB) ArticleRepository {
	return &articleRepository{
		db: db,
	}
}

func (r *articleRepository) GenerateArticle(ctx context.Context, tx *sql.Tx, topicID int) ([]model.Article, error) {
	// Get current time for article creation
	currentTime := time.Now()

	// First, get the topic preference from the database
	var preference string
	var userID int
	query := `SELECT preference, user_id FROM topics WHERE id = $1`

	// Log the query being executed
	fmt.Printf("Executing query: %s with topicID: %d\n", query, topicID)

	err := r.db.QueryRowContext(ctx, query, topicID).Scan(&preference, &userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("topic with ID %d not found in database", topicID)
		}
		return nil, fmt.Errorf("database error when getting topic %d: %w", topicID, err)
	}

	// Log successful retrieval
	fmt.Printf("Successfully retrieved topic ID %d - Preference: %s, UserID: %d\n", topicID, preference, userID)

	// Create a new context with topic_id for the article generation service
	ginCtx, ok := ctx.(*gin.Context)
	if !ok {
		// If we can't get the gin context, create a new one
		ginCtx = &gin.Context{}
	}

	// Set topic_id and user_id in the context
	ginCtx.Set("topic_id", topicID)
	ginCtx.Set("user_id", userID)

	// Use the actual preference for article generation
	preferences := []string{preference}

	// Log the article generation process
	fmt.Printf("Generating articles for topic ID %d with preference: %s\n", topicID, preference)

	// Generate articles using the service package
	generatedArticles, err := service.GenerateArticles(ginCtx, preferences)
	if err != nil {
		return nil, fmt.Errorf("failed to generate articles: %w", err)
	}

	// Check if we got any articles
	if len(generatedArticles) == 0 {
		return nil, fmt.Errorf("no articles were generated for topic ID %d", topicID)
	}

	articles := make([]model.Article, len(generatedArticles))

	// Loop for each generated article
	for i, genArticle := range generatedArticles {
		// Create article model
		article := &model.Article{
			Title:     genArticle.Title,
			Content:   genArticle.Content,
			Source:    genArticle.Source,
			IDTopic:   topicID,
			CreatedAt: currentTime,
		}

		// Use CreateArticle to insert into database
		if err := r.CreateArticle(ctx, tx, article); err != nil {
			return nil, err
		}

		articles[i] = *article
	}

	return articles, nil
}

// CreateArticle inserts a new article into the database using the provided transaction
func (r *articleRepository) CreateArticle(ctx context.Context, tx *sql.Tx, article *model.Article) error {
	// Validate article data before insertion
	if article.Title == "" || article.Content == "" {
		return fmt.Errorf("article title and content cannot be empty")
	}

	// Ensure topic ID exists
	var topicExists bool
	query := `SELECT EXISTS(SELECT 1 FROM topics WHERE id = $1)`
	err := r.db.QueryRowContext(ctx, query, article.IDTopic).Scan(&topicExists)
	if err != nil {
		return fmt.Errorf("failed to check if topic exists: %w", err)
	}

	if !topicExists {
		return fmt.Errorf("topic with ID %d does not exist", article.IDTopic)
	}

	// Insert the article
	insertQuery := `
        INSERT INTO articles (title, content, source, topic_id, created_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id`

	err = tx.QueryRowContext(
		ctx,
		insertQuery,
		article.Title,
		article.Content,
		article.Source,
		article.IDTopic,
		article.CreatedAt,
	).Scan(&article.ID)

	if err != nil {
		return fmt.Errorf("failed to create article: %w", err)
	}

	return nil
}

// Implementasi fungsi baru untuk paginasi
func (r *articleRepository) GetPaginatedArticles(ctx context.Context, page, limit int) ([]model.Article, int64, error) {
	offset := (page - 1) * limit

	// Get total count
	var totalItems int64
	countQuery := "SELECT COUNT(*) FROM articles"
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&totalItems)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count articles: %w", err)
	}

	// Get paginated data
	query := `
        SELECT id, title, content, source, topic_id, created_at 
        FROM articles 
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get articles: %w", err)
	}
	defer rows.Close()

	var articles []model.Article
	for rows.Next() {
		var article model.Article
		err := rows.Scan(
			&article.ID,
			&article.Title,
			&article.Content,
			&article.Source,
			&article.IDTopic,
			&article.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan article: %w", err)
		}
		articles = append(articles, article)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating articles: %w", err)
	}

	return articles, totalItems, nil
}

// Fungsi GetAllArticles tetap ada dan tidak berubah untuk kompatibilitas
func (r *articleRepository) GetAllArticles(ctx context.Context) ([]model.Article, error) {
	query := `
        SELECT id, title, content, source, topic_id, created_at 
        FROM articles 
        ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get articles: %w", err)
	}
	defer rows.Close()

	var articles []model.Article
	for rows.Next() {
		var article model.Article
		err := rows.Scan(
			&article.ID,
			&article.Title,
			&article.Content,
			&article.Source,
			&article.IDTopic,
			&article.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan article: %w", err)
		}
		articles = append(articles, article)
	}

	return articles, nil
}

func (r *articleRepository) GetArticleByID(ctx context.Context, id int) (*model.Article, error) {
	query := `
		SELECT id, title, content, source, topic_id, created_at 
		FROM articles 
		WHERE id = $1`

	var article model.Article
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&article.ID,
		&article.Title,
		&article.Content,
		&article.Source,
		&article.IDTopic,
		&article.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("article not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	return &article, nil
}

func (r *articleRepository) GetArticleByTitle(ctx context.Context, title string) (*model.Article, error) {
	query := `
		SELECT id, title, content, source, topic_id, created_at 
		FROM articles 
		WHERE LOWER(title) = LOWER($1)`

	var article model.Article
	err := r.db.QueryRowContext(ctx, query, title).Scan(
		&article.ID,
		&article.Title,
		&article.Content,
		&article.Source,
		&article.IDTopic,
		&article.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("article not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	return &article, nil
}

func (r *articleRepository) SearchArticlesByTitle(ctx context.Context, title string) ([]model.Article, error) {
	query := `
		SELECT id, title, content, source, topic_id, created_at 
		FROM articles 
		WHERE LOWER(title) LIKE LOWER($1)
		ORDER BY title
		LIMIT 5`

	rows, err := r.db.QueryContext(ctx, query, "%"+title+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to search articles: %w", err)
	}
	defer rows.Close()

	var articles []model.Article
	for rows.Next() {
		var article model.Article
		err := rows.Scan(
			&article.ID,
			&article.Title,
			&article.Content,
			&article.Source,
			&article.IDTopic,
			&article.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan article: %w", err)
		}
		articles = append(articles, article)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating articles: %w", err)
	}

	return articles, nil
}

func (r *articleRepository) DeleteArticle(ctx context.Context, id int) error {
	query := `DELETE FROM articles WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete article: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("article not found")
	}

	return nil
}

func (r *articleRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

func (r *articleRepository) CommitTx(tx *sql.Tx) error {
	return tx.Commit()
}

func (r *articleRepository) RollbackTx(tx *sql.Tx) error {
	return tx.Rollback()
}
