package controller

import (
	"context"
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

	// Endpoint khusus admin
	adminRoutes := goalsGroup.Group("")
	adminRoutes.Use(c.aM.RequireToken("ADMIN"))
	{
		adminRoutes.GET("/:user_id", c.GetUserGoals) // [Admin] Get goals by user ID
	}

	// Endpoint untuk user biasa
	userRoutes := goalsGroup.Group("")
	userRoutes.Use(c.aM.RequireToken("USER", "ADMIN"))
	{
		userRoutes.POST("/:user_id", c.CreateGoal)                  // [User] Create goal
		userRoutes.PUT("/:user_id/:id", c.UpdateGoal)               // [User] Update goal
		userRoutes.PUT("/complete-article", c.CompleteGoalProgress) // [User] Complete article
		userRoutes.DELETE("/:user_id/:id", c.DeleteGoal)            // [User] Delete goal
	}
}

// @Summary      Create a new goal
// @Description  Add a new learning goal with articles to read
// @Tags         Goals
// @Accept       json
// @Produce      json
// @Param        user_id path      int                  true  "User ID"
// @Param        request body      dto.CreateGoalRequest  true  "Goal data"
// @Success      201     {object}  dto.GoalResponse
// @Failure      400     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Router       /pijar/goals/{user_id} [post]
func (c *dailyGoalsController) CreateGoal(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req dto.CreateGoalRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	log.Printf("ArticlesToRead %v", req.ArticlesToRead)

	createdGoal, err := c.uc.CreateGoal(ctx.Request.Context(), userID, req.Title, req.Task, req.ArticlesToRead)
	if err != nil {
		if strings.Contains(err.Error(), "invalid article IDs") {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create goal: " + err.Error()})
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

	ctx.JSON(http.StatusCreated, response)
}

// @Summary      Complete an article progress
// @Description  Mark an article as completed in a goal
// @Tags         Goals
// @Accept       json
// @Produce      json
// @Param        request body dto.CompleteArticleRequest true "Completion Data"
// @Success      200  {object}  dto.GoalProgressResponse
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /pijar/goals/complete-article [put]
func (c *dailyGoalsController) CompleteGoalProgress(ctx *gin.Context) {
	// Parse request body
	var req dto.CompleteArticleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Validate IDs
	if req.UserID <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if req.GoalID <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid goal ID"})
		return
	}

	if req.ArticleID <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	// Call usecase to complete the article progress
	result, err := c.uc.CompleteArticleProgress(context.Background(), req.GoalID, req.ArticleID, req.UserID)
	if err != nil {
		// Check if the error is due to article not found in goal
		if err.Error() == "artikel tidak termasuk dalam goal ini" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete article progress: " + err.Error()})
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

	ctx.JSON(http.StatusOK, response)
}

// @Summary      Get user goals
// @Description  Retrieve all goals for a specific user
// @Tags         Goals
// @Produce      json
// @Param        user_id path int true "User ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /pijar/goals/{user_id} [get]
func (c *dailyGoalsController) GetUserGoals(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	goals, err := c.uc.GetUserGoals(ctx.Request.Context(), userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   goals,
	})
}

// @Summary      Update a goal
// @Description  Update goal details including articles to read
// @Tags         Goals
// @Accept       json
// @Produce      json
// @Param        user_id         path      int                   true  "User ID"
// @Param        id              path      int                   true  "Goal ID"
// @Param        request         body      dto.UpdateGoalRequest  true  "Update Data"
// @Success      200             {object}  dto.GoalProgressResponse
// @Failure      400             {object}  map[string]string
// @Failure      500             {object}  map[string]string
// @Router       /pijar/goals/{user_id}/{id} [put]
func (c *dailyGoalsController) UpdateGoal(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	// Get goal ID from URL
	goalID := ctx.Param("id")
	gID, err := strconv.Atoi(goalID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid goal ID"})
		return
	}

	// Bind request body
	var req dto.UpdateGoalRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
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
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update goal: " + err.Error()})
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

	ctx.JSON(http.StatusOK, response)
}

// @Summary      Delete a goal
// @Description  Permanently delete a goal and its progress
// @Tags         Goals
// @Param        user_id path int true "User ID"
// @Param        id      path int true "Goal ID"
// @Success      200  {object}  map[string]string  "Contoh: {'message':'Goal successfully deleted'}"
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /pijar/goals/{user_id}/{id} [delete]
func (c *dailyGoalsController) DeleteGoal(ctx *gin.Context) {
	// Get user_id from URL
	userID, err := strconv.Atoi(ctx.Param("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get goal_id from URL
	goalID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid goal ID"})
		return
	}

	// Call usecase
	err = c.uc.DeleteGoal(context.Background(), userID, goalID)
	if err != nil {
		// Handle specific errors
		if strings.Contains(err.Error(), "goal not found") {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Goal is not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete goal: " + err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Goal successfully deleted"})
}
