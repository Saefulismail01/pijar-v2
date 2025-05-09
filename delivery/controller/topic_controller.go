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

type TopicControllerImpl struct {
	topicUsecase usecase.TopicUsecase
	rg           gin.RouterGroup
	aM           middleware.AuthMiddleware
}

func NewTopicController(tu usecase.TopicUsecase, rg *gin.RouterGroup, aM middleware.AuthMiddleware) *TopicControllerImpl {
	return &TopicControllerImpl{
		topicUsecase: tu,
		rg:           *rg,
		aM:           aM,
	}
}

func (tc *TopicControllerImpl) Route() {

	topicsGroup := tc.rg.Group("/topics")
	// User Routes
	userRoutes := topicsGroup.Use(tc.aM.RequireToken("USER", "ADMIN"))
	{
		userRoutes.POST("/", tc.CreateTopic)
		userRoutes.GET("/", tc.GetAllTopics)
		userRoutes.PUT("/:id", tc.UpdateTopic)
		userRoutes.DELETE("/:id", tc.DeleteTopic)
	}

	// Admin Routes
	adminRoutes := topicsGroup.Use(tc.aM.RequireToken("ADMIN"))
	{
		adminRoutes.GET("/:id", tc.GetTopicByID)
	}

	// tc.RouterGroup.POST("/topics", tc.CreateTopic)
	// tc.RouterGroup.GET("/topics", tc.GetAllTopics)
	// tc.RouterGroup.GET("/topics/:id", tc.GetTopicByID)
	// tc.RouterGroup.PUT("/topics/h:id", tc.UpdateTopic)
	// tc.RouterGroup.DELETE("/topics/:id", tc.DeleteTopic)

}

func (tc *TopicControllerImpl) CreateTopic(c *gin.Context) {
	// Parse request body
	var input dto.InputTopic

	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Bad Request",
			"errors":  fmt.Sprintf("Format request tidak valid: %v", err.Error()),
		})
		return
	}

	// Create topic with provided user ID
	topicID, err := tc.topicUsecase.CreateTopic(c.Request.Context(), input.UserID, input.Preference)
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
		"message": "Topic created successfully",
		"data": gin.H{
			"id":         topicID,
			"user_id":    input.UserID,
			"preference": input.Preference,
		},
	})
}

func (tc *TopicControllerImpl) GetTopicByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Bad Request",
			"errors":  "Invalid ID format",
		})
		return
	}

	topic, err := tc.topicUsecase.GetTopicByID(c.Request.Context(), id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Internal Server Error",
			"errors":  err.Error(),
		})
		return
	}

	if topic == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"status":  http.StatusNotFound,
			"message": "Not Found",
			"errors":  fmt.Sprintf("Topic with ID %d not found", id),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Topic retrieved successfully",
		"data":    topic,
	})
}

func (tc *TopicControllerImpl) GetAllTopics(c *gin.Context) {
	topics, err := tc.topicUsecase.GetAllTopics(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Internal Server Error",
			"errors":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Topics retrieved successfully",
		"data":    topics,
	})
}

func (tc *TopicControllerImpl) UpdateTopic(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Bad Request",
			"errors":  "Invalid ID format",
		})
		return
	}

	var input struct {
		Preference string `json:"preference" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Bad Request",
			"errors":  fmt.Sprintf("Format request tidak valid: %v", err.Error()),
		})
		return
	}

	err = tc.topicUsecase.UpdateTopic(c.Request.Context(), id, input.Preference)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Internal Server Error",
			"errors":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Topic updated successfully",
	})
}

func (tc *TopicControllerImpl) DeleteTopic(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Bad Request",
			"errors":  "Invalid ID format",
		})
		return
	}

	err = tc.topicUsecase.DeleteTopic(c.Request.Context(), id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Internal Server Error",
			"errors":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Topic deleted successfully",
	})
}
