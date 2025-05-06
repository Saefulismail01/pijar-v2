package controller

import (
	"context"
	"net/http"
	// "pijar/model"
	"pijar/model/dto"
	"pijar/usecase"
	"strconv"

	"github.com/gin-gonic/gin"
)

type dailyGoalsController struct {
	uc usecase.DailyGoalUseCase
	rg *gin.RouterGroup
}

func NewGoalController(uc usecase.DailyGoalUseCase, rg *gin.RouterGroup) *dailyGoalsController {
	return &dailyGoalsController{uc: uc, rg: rg}
}

func (c *dailyGoalsController) Route() {
    c.rg.POST("/goals/:user_id", c.CreateGoal)
}

func (c *dailyGoalsController) CreateGoal(ctx *gin.Context) {
    userID, err := strconv.Atoi(ctx.Param("user_id"))
    if err != nil {
        return 
    }

	var req dto.CreateGoalRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	createdGoal, err := c.uc.CreateGoal(
		context.Background(),
		userID,
		req.Title,
		req.Task,
		req.ArticleIDs,
	)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create goal" + err.Error()})
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
