package controller

import (
	"fmt"
	"net/http"
	"pijar/middleware"
	"pijar/model/dto"
	"pijar/usecase"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ArticleControllerImpl struct {
	articleUsecase usecase.ArticleUsecase
	rg             *gin.RouterGroup
	aM             middleware.AuthMiddleware
}

func NewArticleController(au usecase.ArticleUsecase, rg *gin.RouterGroup, aM middleware.AuthMiddleware) *ArticleControllerImpl {
	return &ArticleControllerImpl{
		articleUsecase: au,
		rg:             rg,
		aM:             aM,
	}
}

func (ac *ArticleControllerImpl) Route() {
	articlesGroup := ac.rg.Group("/articles")

	//admin-specific endpoint
	adminRoutes := articlesGroup.Group("")
	adminRoutes.Use(ac.aM.RequireToken("ADMIN"))
	{
		adminRoutes.GET("/:id", ac.GetArticleByID)
		adminRoutes.DELETE("/:id", ac.DeleteArticle)
	}

	//user endpoint
	userRoutes := articlesGroup.Group("")
	userRoutes.Use(ac.aM.RequireToken("USER", "ADMIN"))
	{
		userRoutes.GET("", ac.GetAllArticles)
		userRoutes.GET("/all", ac.GetAllArticlesWithoutPagination)
		userRoutes.POST("/generate", ac.GenerateArticle)
		userRoutes.POST("/search", ac.SearchArticleByTitle)
	}
}

func (ac *ArticleControllerImpl) SearchArticleByTitle(c *gin.Context) {
	var searchReq dto.ArticleSearchRequest
	if err := c.ShouldBindJSON(&searchReq); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Bad Request",
			Error:   "Invalid request body",
		})
		return
	}

	if searchReq.Title == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Bad Request",
			Error:   "Title is required in the request body",
		})
		return
	}

	// Search for articles with similar titles
	articles, err := ac.articleUsecase.SearchArticlesByTitle(c.Request.Context(), searchReq.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Internal Server Error",
			Error:   "Failed to search articles",
		})
		return
	}

	// Check if any articles were found
	if len(articles) == 0 {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Message: "Not Found",
			Error:   "No articles found matching the search criteria",
		})
		return
	}

	// Prepare response
	response := dto.ArticleSearchResponse{
		Found:   true,
		Article: articles[0],
		Message: "Article found",
	}

	// If there are more than one result, include them as suggestions
	if len(articles) > 1 {
		suggestions := make([]string, 0, len(articles)-1)
		for _, article := range articles[1:] {
			suggestions = append(suggestions, article.Title)
		}
		response.Suggestions = suggestions
	}

	c.JSON(http.StatusOK, dto.Response{
		Message: "Articles retrieved successfully",
		Data:    response,
	})
}

// GenerateArticle handles article generation from topic ID
func (ac *ArticleControllerImpl) GenerateArticle(c *gin.Context) {
	var input struct {
		TopicID int `json:"topic_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Bad Request",
			Error:   err.Error(),
		})
		return
	}

	// Generate article
	article, err := ac.articleUsecase.GenerateArticle(c.Request.Context(), input.TopicID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Internal Server Error",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Message: "Article generated successfully",
		Data:    article,
	})
}

func (ac *ArticleControllerImpl) GetAllArticles(c *gin.Context) {
	// Get query parameters
	page, _ := strconv.Atoi(c.Query("page"))
	if page == 0 {
		page = 1
	}

	// Get articles with pagination
	articles, err := ac.articleUsecase.GetAllArticles(c.Request.Context(), page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Internal Server Error",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Message: "Articles retrieved successfully",
		Data:    articles,
	})
}

func (ac *ArticleControllerImpl) GetAllArticlesWithoutPagination(c *gin.Context) {
	// Get all articles without pagination
	articles, err := ac.articleUsecase.GetAllArticlesWithoutPagination(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Get all articles without pagination successful",
		"data":    articles,
	})
}

func (ac *ArticleControllerImpl) GetArticleByID(c *gin.Context) {
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

// func (ac *ArticleControllerImpl) GetArticleByTitle(c *gin.Context) {
// 	var input struct {
// 		Title string `json:"title" binding:"required"`
// 	}

// 	if err := c.ShouldBindJSON(&input); err != nil {
// 		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
// 			"status":  http.StatusBadRequest,
// 			"message": "Bad Request",
// 			"errors":  "Title is required in the request body",
// 		})
// 		return
// 	}

// 	if input.Title == "" {
// 		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
// 			"status":  http.StatusBadRequest,
// 			"message": "Bad Request",
// 			"errors":  "Title cannot be empty",
// 		})
// 		return
// 	}

// 	article, err := ac.articleUsecase.GetArticleByTitle(c.Request.Context(), input.Title)
// 	if err != nil {
// 		statusCode := http.StatusInternalServerError
// 		if err.Error() == fmt.Sprintf("article with title '%s' not found", input.Title) {
// 			statusCode = http.StatusNotFound
// 		}
// 		c.AbortWithStatusJSON(statusCode, gin.H{
// 			"status":  statusCode,
// 			"message": http.StatusText(statusCode),
// 			"errors":  err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"status":  http.StatusOK,
// 		"message": "Article retrieved successfully",
// 		"data":    article,
// 	})
// }

// func (ac *ArticleControllerImpl) UpdateArticle(c *gin.Context) {
// 	idStr := c.Param("id")
// 	id, err := strconv.Atoi(idStr)
// 	if err != nil {
// 		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
// 		return
// 	}

// 	var updateDTO dto.ArticleDto
// 	if err := c.ShouldBindJSON(&updateDTO); err != nil {
// 		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	err = ac.articleUsecase.UpdateArticle(c.Request.Context(), &updateDTO, id)
// 	if err != nil {
// 		if err.Error() == fmt.Sprintf("article with ID %d not found", id) {
// 			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
// 			return
// 		}
// 		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "Article update successful",
// 	})
// }

func (ac *ArticleControllerImpl) DeleteArticle(c *gin.Context) {
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
