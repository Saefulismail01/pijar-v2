package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"pijar/usecase"

	"github.com/gin-gonic/gin"
)

type SessionHandler struct {
	usecase usecase.SessionUsecase
	rg      gin.RouterGroup
}

type CoachRequest struct {
	UserInput string `json:"user_input"`
}

type CoachResponse struct {
	AIResponse string `json:"ai_response"`
}

func (h *SessionHandler) HandleChat(c *gin.Context) {
	var req CoachRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := strconv.Atoi(c.GetHeader("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ser id tidak valid"})
		return
	}

	response, err := h.usecase.StartSession(c.Request.Context(), userID, req.UserInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, CoachResponse{AIResponse: response})
}

func (h *SessionHandler) GetSessionByUserID(c *gin.Context) {
	userID, err := strconv.Atoi(c.GetHeader("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user id tidak valid"})
		return
	}

	response, err := h.usecase.GetSessionByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *SessionHandler) DeleteSessionByUserID(c *gin.Context) {
	userID, err := strconv.Atoi(c.GetHeader("user_id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": "Nomor id tidak valid"})
		return
	}

	err = h.usecase.DeleteSessionByUserID(c.Request.Context(), userID)
	if err != nil {
		if strings.Contains(err.Error(), "tidak ditemukan") {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"err": err.Error()})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Session untuk user ID %d berhasil dihapus", userID),
	})
	
}

func (h *SessionHandler) Route() {
	h.rg.POST("/coach", h.HandleChat)
	h.rg.GET("/coach", h.GetSessionByUserID)
	h.rg.DELETE("/coach", h.DeleteSessionByUserID)
}

func NewSessionHandler(uc usecase.SessionUsecase, rg *gin.RouterGroup) *SessionHandler {
	return &SessionHandler{
		usecase: uc,
		rg:      *rg}
}
