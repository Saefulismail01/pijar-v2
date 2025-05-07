package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"pijar/model"
	"pijar/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ArticleRepository interface {
	CreateArticle(ctx context.Context, tx *sql.Tx, article *model.Article) error
	GenerateArticle(ctx context.Context, tx *sql.Tx, topicID int) ([]model.Article, error)
	GetAllArticles(ctx context.Context) ([]model.Article, error)
	GetArticleByID(ctx context.Context, id int) (*model.Article, error)
	GetArticleByTitle(ctx context.Context, title string) (*model.Article, error)
	UpdateArticle(ctx context.Context, article *model.Article) error
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
	// Nilai dari context
	currentTime, _ := time.Parse("2006-01-02 15:04:05", "2025-05-07 10:42:41")

	// Langsung convert topic_id ke string untuk preference
	preferences := []string{strconv.Itoa(topicID)}

	// Generate articles menggunakan utils
	generatedArticles, err := utils.GenerateArticles(ctx.(*gin.Context), preferences)
	if err != nil {
		return nil, fmt.Errorf("failed to generate articles: %w", err)
	}

	articles := make([]model.Article, len(generatedArticles))

	// Loop untuk setiap article yang digenerate
	for i, genArticle := range generatedArticles {
		// Buat article model
		article := &model.Article{
			Title:     genArticle.Title,
			Content:   genArticle.Content,
			Source:    genArticle.Source,
			IDTopic:   topicID,
			CreatedAt: currentTime,
		}

		// Gunakan CreateArticle untuk insert ke database
		if err := r.CreateArticle(ctx, tx, article); err != nil {
			return nil, err
		}

		articles[i] = *article
	}

	return articles, nil
}

// CreateArticle juga perlu disesuaikan
func (r *articleRepository) CreateArticle(ctx context.Context, tx *sql.Tx, article *model.Article) error {
	query := `
        INSERT INTO articles (title, content, source, id_topic, created_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id`

	err := tx.QueryRowContext(
		ctx,
		query,
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

func (r *articleRepository) GetAllArticles(ctx context.Context) ([]model.Article, error) {
	query := `
        SELECT id, title, content, source, id_topic, created_at 
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
		SELECT id, title, content, source, id_topic, created_at 
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
		SELECT id, title, content, source, id_topic, created_at 
		FROM articles 
		WHERE title = $1`

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

func (r *articleRepository) UpdateArticle(ctx context.Context, article *model.Article) error {
	query := `
		UPDATE articles 
		SET title = $1, content = $2, source = $3, id_topic = $4
		WHERE id = $5`

	result, err := r.db.ExecContext(ctx, query,
		article.Title,
		article.Content,
		article.Source,
		article.IDTopic,
		article.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update article: %w", err)
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
