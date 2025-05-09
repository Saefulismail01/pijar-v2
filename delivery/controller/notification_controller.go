package controller

import (
	"net/http"
	"pijar/repository"
	"pijar/usecase"

	"github.com/gin-gonic/gin"
)

type NotificationController struct {
	UserRepo repository.UserRepo
	UC       usecase.NotificationUseCase
	RG       *gin.RouterGroup
}

func NewNotificationController(
	userRepo repository.UserRepo,
	uc usecase.NotificationUseCase,
	rg *gin.RouterGroup,
) *NotificationController {
	return &NotificationController{
		UserRepo: userRepo,
		UC:       uc,
		RG:       rg,
	}
}

func (c *NotificationController) Route() {
	c.RG.POST("/device-token", c.saveDeviceToken)
}

func (c *NotificationController) saveDeviceToken(ctx *gin.Context) {
	userID, _ := ctx.Get("userID")
	var req struct {
		Token    string `json:"token" binding:"required"`
		Platform string `json:"platform" binding:"required,oneof=android ios web"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.UserRepo.SaveDeviceToken(
		userID.(int),
		req.Token,
		req.Platform,
	); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Device token saved"})
}
