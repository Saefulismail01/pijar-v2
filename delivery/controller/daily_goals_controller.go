package controller

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"pijar/middleware"
	"pijar/model/dto"
	"pijar/usecase"
	"strconv"

	"github.com/gin-gonic/gin"
)

type dailyGoalsController struct {
	uc usecase.DailyGoalUseCase
	rg *gin.RouterGroup
	aM middleware.AuthMiddleware
}

func NewGoalController(
	uc usecase.DailyGoalUseCase,
	rg *gin.RouterGroup,
	aM middleware.AuthMiddleware,
) *dailyGoalsController {
	return &dailyGoalsController{uc: uc, rg: rg, aM: aM}
}

func (c *dailyGoalsController) Route() {
	goalsGroup := c.rg.Group("/goals")

	// Admin-specific endpoint
	adminRoutes := goalsGroup.Group("")
	adminRoutes.Use(c.aM.RequireToken("ADMIN"))
	{
		adminRoutes.GET("/:user_id", c.GetUserGoals) 
	}

	// Endpoint for regular users
	userRoutes := goalsGroup.Group("")
	userRoutes.Use(c.aM.RequireToken("USER", "ADMIN"))
	{
		userRoutes.POST("/:user_id", c.CreateGoal)                  
		userRoutes.PUT("/:user_id/:id", c.UpdateGoal)               
		userRoutes.PUT("/complete-article", c.CompleteGoalProgress) 
		userRoutes.DELETE("/:user_id/:id", c.DeleteGoal)            
	}
}

func (c *dailyGoalsController) CreateGoal(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Invalid user ID",
			Error:   err.Error(),
		})
		return
	}

	var req dto.CreateGoalRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	log.Printf("ArticlesToRead %v", req.ArticlesToRead)

	createdGoal, err := c.uc.CreateGoal(ctx.Request.Context(), userID, req.Title, req.Task, req.ArticlesToRead)
	if err != nil {
		if strings.Contains(err.Error(), "invalid article IDs") {
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Message: "Invalid article IDs",
				Error:   err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Failed to create goal",
			Error:   err.Error(),
		})
		return
	}

	response := dto.GoalResponse{
		ID:             createdGoal.ID,
		Title:          createdGoal.Title,
		Task:           createdGoal.Task,
		ArticlesToRead: createdGoal.ArticlesToRead,
		Completed:      createdGoal.Completed,
		CreatedAt:      createdGoal.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Message: "Goal created successfully",
		Data:    response,
	})
}

func (c *dailyGoalsController) CompleteGoalProgress(ctx *gin.Context) {
	// Parse request body
	var req dto.CompleteArticleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Validate IDs
	if req.UserID <= 0 {

		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Invalid user ID",
		})
		return
	}

	if req.GoalID <= 0 {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Invalid goal ID",
		})
		return
	}

	if req.ArticleID <= 0 {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Invalid article ID",
		})
		return
	}

	// Call usecase to complete article progress
	result, err := c.uc.CompleteArticleProgress(context.Background(), req.GoalID, req.ArticleID, req.UserID)
	if err != nil {
		// Check if error is due to article not found in goal
		if err.Error() == "artikel tidak termasuk dalam goal ini" {
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Failed to complete article progress",
			Error:   err.Error(),
		})
		return
	}

	// Convert article progress to response format
	var articles []dto.ArticleProgress
	for _, p := range result.Progress {
		articles = append(articles, dto.ArticleProgress{
			ArticleID:     p.ArticleID,
			Completed:     p.Completed,
			DateCompleted: p.DateCompleted,
		})
	}

	// Count completed articles
	completedCount := 0
	for _, a := range articles {
		if a.Completed {
			completedCount++
		}
	}

	// Build response
	response := dto.GoalProgressResponse{
		ID:             result.Goal.ID,
		Title:          result.Goal.Title,
		Task:           result.Goal.Task,
		Articles:       articles,
		Completed:      result.Goal.Completed,
		CreatedAt:      result.Goal.CreatedAt.Format("2006-01-02 15:04:05"),
		TotalCompleted: completedCount,
		TotalArticles:  len(articles),
	}

	message := fmt.Sprintf("Article %v is finish", req.ArticleID)

	ctx.JSON(http.StatusOK, dto.Response{
		Message: message,
		Data:    response,
	})
}

func (c *dailyGoalsController) GetUserGoals(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Invalid user ID",
		})
		return
	}

	goals, err := c.uc.GetUserGoals(ctx.Request.Context(), userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Message: "Get user goals successful",
		Data:    goals,
	})
}

func (c *dailyGoalsController) UpdateGoal(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Invalid user ID",
		})
		return
	}

	// Get goal ID from URL
	goalID := ctx.Param("id")
	gID, err := strconv.Atoi(goalID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Invalid goal ID"})
		return
	}

	// Bind request body
	var req dto.UpdateGoalRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Call usecase
	result, err := c.uc.UpdateGoal(
		context.Background(),
		userID,
		gID,
		req.Title,
		req.Task,
		req.Completed,
		req.ArticlesToRead,
	)
	if err != nil {
		// Handle article IDs error from usecase
		if strings.Contains(err.Error(), "invalid article IDs") {
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Message: err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Failed to update goal",
			Error:   err.Error(),
		})
		return
	}

	// Convert article progress to response format
	var articles []dto.ArticleProgress
	for _, p := range result.Progress {
		articles = append(articles, dto.ArticleProgress{
			ArticleID:     p.ArticleID,
			Completed:     p.Completed,
			DateCompleted: p.DateCompleted,
		})
	}

	// Count completed articles
	completedCount := 0
	for _, a := range articles {
		if a.Completed {
			completedCount++
		}
	}

	// Build response
	response := dto.GoalProgressResponse{
		ID:             result.Goal.ID,
		Title:          result.Goal.Title,
		Task:           result.Goal.Task,
		Articles:       articles,
		Completed:      result.Goal.Completed,
		CreatedAt:      result.Goal.CreatedAt.Format("2006-01-02 15:04:05"),
		TotalCompleted: completedCount,
		TotalArticles:  len(articles),
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Message: "Update user goals successful",
		Data:    response,
	})
}

func (c *dailyGoalsController) DeleteGoal(ctx *gin.Context) {
	// Get user_id from URL
	userID, err := strconv.Atoi(ctx.Param("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Invalid user ID",
		})
		return
	}

	// Get goal_id from URL
	goalID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Invalid goal ID",
		})
		return
	}

	// Call usecase
	err = c.uc.DeleteGoal(context.Background(), userID, goalID)
	if err != nil {
		// Handle specific errors
		if strings.Contains(err.Error(), "goal not found") {
			ctx.JSON(http.StatusNotFound, gin.H{
				"message": "Goal is not found",
			})
		} else {
			ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Message: "Failed to delete goal",
				Error:   err.Error(),
			})
		}
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Message: "Goal deleted successfully",
	})
}
