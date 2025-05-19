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
	rg           *gin.RouterGroup
	aM           middleware.AuthMiddleware
}

func NewTopicController(tu usecase.TopicUsecase, rg *gin.RouterGroup, aM middleware.AuthMiddleware) *TopicControllerImpl {
	return &TopicControllerImpl{
		topicUsecase: tu,
		rg:           rg,
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

}

func (tc *TopicControllerImpl) CreateTopic(c *gin.Context) {
	// get user ID from jwt body
	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Message: "Authentication required",
		})
		return
	}
	userID, ok := val.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Message: "Invalid user identity in context",
		})
		return
	}

	var input struct {
		Preference string `json:"preference" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Bad Request",
			Error:   "Invalid request body",
		})
		return
	}

	// Create topic with provided user ID
	topicID, err := tc.topicUsecase.CreateTopic(c.Request.Context(), userID, input.Preference)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Internal Server Error",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.Response{
		Message: "Topic created successfully",
		Data: gin.H{
			"id":         topicID,
			"user_id":    userID,
			"preference": input.Preference,
		},
	})
}

func (tc *TopicControllerImpl) GetTopicByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Bad Request",
			Error:   "Invalid ID format",
		})
		return
	}

	topic, err := tc.topicUsecase.GetTopicByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Message: "Not Found",
			Error:   "Topic not found",
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Message: "Topic retrieved successfully",
		Data:    topic,
	})

}

func (tc *TopicControllerImpl) GetAllTopics(c *gin.Context) {
	topics, err := tc.topicUsecase.GetAllTopics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Internal Server Error",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Message: "Topics retrieved successfully",
		Data:    topics,
	})
}

func (tc *TopicControllerImpl) UpdateTopic(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Bad Request",
			Error:   "Invalid ID format",
		})
		return
	}

	var input struct {
		Preference string `json:"preference" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Bad Request",
			Error:   "Invalid request body format",
		})
		return
	}

	// Get user ID from JWT token
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Message: "Authentication required",
		})
		return
	}
	userIDInt, ok := userID.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Invalid user identity",
			Error:   "Failed to parse user ID from token",
		})
		return
	}

	// Check if the topic belongs to the user
	topic, err := tc.topicUsecase.GetTopicByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Message: "Not Found",
			Error:   "Topic not found",
		})
		return
	}

	// Verify topic ownership
	if topic.UserID != userIDInt {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Message: "Forbidden",
			Error:   "Cannot update topic that doesn't belong to you",
		})
		return
	}

	err = tc.topicUsecase.UpdateTopic(c.Request.Context(), id, input.Preference)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Internal Server Error",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Message: "Topic updated successfully",
	})
}

func (tc *TopicControllerImpl) DeleteTopic(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Bad Request",
			Error:   "Invalid ID format",
		})
		return
	}

	if _, exists := c.Get("userID"); !exists {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Message: "Authentication required",
		})
		return
	}

	if err := tc.topicUsecase.DeleteTopic(c.Request.Context(), id); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == fmt.Sprintf("topic with ID %d not found", id) {
			status = http.StatusNotFound
		}
		c.JSON(status, dto.ErrorResponse{
			Message: http.StatusText(status),
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Message: "Topic deleted successfully",
	})
}
