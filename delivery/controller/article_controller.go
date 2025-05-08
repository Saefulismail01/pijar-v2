package controller

import (
	"fmt"
	"net/http"
	"pijar/model/dto"
	"pijar/usecase"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ArticleControllerImpl struct {
	articleUsecase usecase.ArticleUsecase
	RouterGroup    *gin.RouterGroup
}

func (ac *ArticleControllerImpl) SearchArticleByTitle(c *gin.Context) {
	var searchReq dto.ArticleSearchRequest
	if err := c.ShouldBindJSON(&searchReq); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	if searchReq.Title == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Title is required in the request body",
		})
		return
	}

	// Search for articles with similar titles
	articles, err := ac.articleUsecase.SearchArticlesByTitle(c.Request.Context(), searchReq.Title)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to search articles",
			"error":   err.Error(),
		})
		return
	}

	// Check if any articles were found
	if len(articles) == 0 {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"status":  http.StatusNotFound,
			"message": "No articles found matching the search criteria",
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

	c.JSON(http.StatusOK, response)
}

// GenerateArticle handles article generation from topic ID
func (ac *ArticleControllerImpl) GenerateArticle(c *gin.Context) {
	// For simplicity, use a default user ID
	// Parse the request body
	var input dto.GenerateArticleRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Bad Request",
			"errors":  fmt.Sprintf("Format request tidak valid: %v", err.Error()),
		})
		return
	}

	// Log the request
	fmt.Printf("GenerateArticle request received for topic ID: %d\n", input.TopicID)

	// Generate articles based on the topic ID
	articles, err := ac.articleUsecase.GenerateArticle(c.Request.Context(), input.TopicID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMsg := err.Error()

		// Check for specific error messages to provide appropriate status codes
		if errorMsg == fmt.Sprintf("topic with ID %d not found in database", input.TopicID) ||
			errorMsg == fmt.Sprintf("topic with ID %d does not exist", input.TopicID) ||
			errorMsg == fmt.Sprintf("topic with ID %d not found", input.TopicID) {
			statusCode = http.StatusNotFound
		}

		// Log the detailed error
		fmt.Printf("Error generating articles for topic ID %d: %v\n", input.TopicID, errorMsg)

		c.AbortWithStatusJSON(statusCode, gin.H{
			"status":  statusCode,
			"message": http.StatusText(statusCode),
			"errors":  errorMsg,
			"details": fmt.Sprintf("Failed to generate articles for topic ID: %d", input.TopicID),
		})
		return
	}

	// Return the generated articles
	c.JSON(http.StatusCreated, gin.H{
		"status":  http.StatusCreated,
		"message": "Articles generated successfully",
		"data":    articles,
	})
}

func (ac *ArticleControllerImpl) GetAllArticles(c *gin.Context) {
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

func (ac *ArticleControllerImpl) Route() {
	ac.RouterGroup.GET("/articles", ac.GetAllArticles)
	ac.RouterGroup.GET("/articles/:id", ac.GetArticleByID)
	ac.RouterGroup.POST("/articles/generate", ac.GenerateArticle)
	//ac.RouterGroup.PUT("/articles/:id", ac.UpdateArticle)
	ac.RouterGroup.DELETE("/articles/:id", ac.DeleteArticle)
	ac.RouterGroup.POST("/articles/search", ac.SearchArticleByTitle)
}

func NewArticleController(au usecase.ArticleUsecase, rg *gin.RouterGroup) *ArticleControllerImpl {
	return &ArticleControllerImpl{
		articleUsecase: au,
		RouterGroup:    rg}
}
