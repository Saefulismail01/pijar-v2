package controller

import (
	"net/http"
	"strconv"

	"pijar/middleware"
	"pijar/usecase"

	"github.com/gin-gonic/gin"
)

type NotificationController struct {
	// NotifRepo repository.NotifRepoInterface
	UC usecase.NotificationUseCase
	RG *gin.RouterGroup
	aM middleware.AuthMiddleware
}

func NewNotificationController(
	// notifRepo repository.NotifRepoInterface,
	uc usecase.NotificationUseCase,
	rg *gin.RouterGroup,
	aM middleware.AuthMiddleware,
) *NotificationController {
	return &NotificationController{
		// NotifRepo: notifRepo,
		UC: uc,
		RG: rg,
		aM: aM,
	}
}

func (c *NotificationController) Route() {
	authGroup := c.RG.Group("").Use(c.aM.RequireToken("USER"))
	authGroup.POST("/device-token/:userID", c.saveDeviceToken)
}

func (c *NotificationController) saveDeviceToken(ctx *gin.Context) {
	var req struct {
		Token    string `json:"token" binding:"required"`
		Platform string `json:"platform" binding:"required,oneof=android ios web"`
	}
	// Ambil userID dengan error handling
	userID := ctx.Param("userID")

	// if !d {
	// 	ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
	// 	return
	// }

	// Konversi dengan type assertion yang aman
	userIDInt, error := strconv.Atoi(userID)
	if error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID type"})
		return
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.UC.NotifRepo.SaveDeviceToken(
		userIDInt,
		req.Token,
		req.Platform,
	); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Device token saved"})
}
