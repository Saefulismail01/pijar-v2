package mock

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Models
type Article struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Preferences []string  `json:"preferences"`
	CreatedAt   time.Time `json:"created_at"`
}

type GenerateArticleInput struct {
	Preferences []string `json:"preferences" binding:"required"`
}

// Mock Storage for Articles
type articleStorage struct {
	mu       sync.RWMutex
	articles map[int]*Article
	lastID   int
}

var articleStore = &articleStorage{
	articles: make(map[int]*Article),
	lastID:   0,
}

// Mock Handler
type ArticleMockHandler struct{}

func NewArticleMockHandler() *ArticleMockHandler {
	return &ArticleMockHandler{}
}

func (h *ArticleMockHandler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.POST("/articles/generate", h.GenerateArticle)
		api.GET("/articles", h.GetAllArticles)
		api.GET("/articles/:id", h.GetArticleByID)
		api.GET("/articles/title/:title", h.GetArticleByTitle)
		api.PUT("/articles/:id", h.UpdateArticle)
		api.DELETE("/articles/:id", h.DeleteArticle)
	}
}

func (h *ArticleMockHandler) GenerateArticle(c *gin.Context) {
	var input GenerateArticleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Format request tidak valid: " + err.Error(),
		})
		return
	}

	if len(input.Preferences) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Preferences tidak boleh kosong",
		})
		return
	}

	articleStore.mu.Lock()
	defer articleStore.mu.Unlock()

	articleStore.lastID++
	now := time.Now().UTC()

	// Create mock article
	article := &Article{
		ID:          articleStore.lastID,
		Title:       fmt.Sprintf("Generated Article %d about %v", articleStore.lastID, input.Preferences),
		Content:     fmt.Sprintf("This is a mock article content generated for preferences: %v", input.Preferences),
		Preferences: input.Preferences,
		CreatedAt:   now,
	}

	articleStore.articles[article.ID] = article

	c.JSON(http.StatusCreated, gin.H{
		"message": "Article generation successful",
		"data": gin.H{
			"id": article.ID,
		},
	})
}

func (h *ArticleMockHandler) GetAllArticles(c *gin.Context) {
	articleStore.mu.RLock()
	defer articleStore.mu.RUnlock()

	articles := make([]*Article, 0, len(articleStore.articles))
	for _, article := range articleStore.articles {
		articles = append(articles, article)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Get all articles successful",
		"data":    articles,
	})
}

func (h *ArticleMockHandler) GetArticleByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ID format",
		})
		return
	}

	articleStore.mu.RLock()
	defer articleStore.mu.RUnlock()

	article, exists := articleStore.articles[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("article with ID %d not found", id),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Get article by ID successful",
		"data":    article,
	})
}

func (h *ArticleMockHandler) GetArticleByTitle(c *gin.Context) {
	title := c.Param("title")
	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Title parameter is required",
		})
		return
	}

	articleStore.mu.RLock()
	defer articleStore.mu.RUnlock()

	var foundArticle *Article
	for _, article := range articleStore.articles {
		if article.Title == title {
			foundArticle = article
			break
		}
	}

	if foundArticle == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("article with title '%s' not found", title),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Get article by title successful",
		"data":    foundArticle,
	})
}

func (h *ArticleMockHandler) UpdateArticle(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ID format",
		})
		return
	}

	var updateData Article
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	articleStore.mu.Lock()
	defer articleStore.mu.Unlock()

	article, exists := articleStore.articles[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("article with ID %d not found", id),
		})
		return
	}

	// Update allowed fields
	article.Title = updateData.Title
	article.Content = updateData.Content
	article.Preferences = updateData.Preferences

	c.JSON(http.StatusOK, gin.H{
		"message": "Article update successful",
	})
}

func (h *ArticleMockHandler) DeleteArticle(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ID format",
		})
		return
	}

	articleStore.mu.Lock()
	defer articleStore.mu.Unlock()

	if _, exists := articleStore.articles[id]; !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("article with ID %d not found", id),
		})
		return
	}

	delete(articleStore.articles, id)

	c.JSON(http.StatusOK, gin.H{
		"message": "Article deletion successful",
	})
}
