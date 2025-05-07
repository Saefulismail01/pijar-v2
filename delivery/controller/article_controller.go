package controller

import (
	"fmt"
	"net/http"
	"pijar/model/dto"
	"pijar/usecase"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ArticleController interface {
	GenerateArticle(c *gin.Context)
	GetAllArticles(c *gin.Context)
	GetArticleByID(c *gin.Context)
	GetArticleByTitle(c *gin.Context)
	UpdateArticle(c *gin.Context)
	DeleteArticle(c *gin.Context)
	RegisterRoutes(rg *gin.RouterGroup, protected *gin.RouterGroup)
}

type articleControllerImpl struct {
	articleUsecase usecase.ArticleUsecase
}

func NewArticleController(au usecase.ArticleUsecase) ArticleController {
	return &articleControllerImpl{articleUsecase: au}
}

func (ac *articleControllerImpl) RegisterRoutes(rg *gin.RouterGroup, protected *gin.RouterGroup) {
	// Public routes
	rg.GET("/articles", ac.GetAllArticles)
	rg.GET("/articles/:id", ac.GetArticleByID)
	rg.GET("/articles/title/:title", ac.GetArticleByTitle)

	// Protected routes
	protected.POST("/articles/generate", ac.GenerateArticle)
	protected.PUT("/articles/:id", ac.UpdateArticle)
	protected.DELETE("/articles/:id", ac.DeleteArticle)
}

// GenerateArticle handles article generation from topic ID
func (ac *articleControllerImpl) GenerateArticle(c *gin.Context) {
	var input dto.ArticleDto
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Bad Request",
			"errors":  fmt.Sprintf("Format request tidak valid: %v", err.Error()),
		})
		return
	}

	articles, err := ac.articleUsecase.GenerateArticle(c.Request.Context(), input.TopicID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Internal Server Error",
			"errors":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  http.StatusCreated,
		"message": "Success generate article",
		"data":    articles,
	})
}

func (ac *articleControllerImpl) GetAllArticles(c *gin.Context) {
	articles, err := ac.articleUsecase.GetAllArticles(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Get all articles successful",
		"data":    articles,
	})
}

func (ac *articleControllerImpl) GetArticleByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	article, err := ac.articleUsecase.GetArticleByID(c.Request.Context(), id)
	if err != nil {
		if err.Error() == fmt.Sprintf("article with ID %d not found", id) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Get article by ID successful",
		"data":    article,
	})
}

func (ac *articleControllerImpl) GetArticleByTitle(c *gin.Context) {
	title := c.Param("title")
	if title == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Title parameter is required"})
		return
	}

	article, err := ac.articleUsecase.GetArticleByTitle(c.Request.Context(), title)
	if err != nil {
		if err.Error() == fmt.Sprintf("article with title '%s' not found", title) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Get article by title successful",
		"data":    article,
	})
}

func (ac *articleControllerImpl) UpdateArticle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var updateDTO dto.ArticleDto
	if err := c.ShouldBindJSON(&updateDTO); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = ac.articleUsecase.UpdateArticle(c.Request.Context(), &updateDTO, id)
	if err != nil {
		if err.Error() == fmt.Sprintf("article with ID %d not found", id) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Article update successful",
	})
}

func (ac *articleControllerImpl) DeleteArticle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	err = ac.articleUsecase.DeleteArticle(c.Request.Context(), id)
	if err != nil {
		if err.Error() == fmt.Sprintf("article with ID %d not found", id) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Article deletion successful",
	})
}
