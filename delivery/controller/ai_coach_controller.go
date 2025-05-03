package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"pijar/usecase"
)

type SessionHandler struct {
	usecase usecase.SessionUsecase
	rg gin.RouterGroup
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	response, err := h.usecase.StartSession(userID, req.UserInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, CoachResponse{AIResponse: response})
}

func (h *SessionHandler) GetSessionByUserID(c *gin.Context) {
	userID, err := strconv.Atoi(c.GetHeader("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	response, err := h.usecase.GetSessionByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *SessionHandler) Route(){
	h.rg.POST("/coach", h.HandleChat)
	h.rg.GET("/coach", h.GetSessionByUserID)
}

func NewSessionHandler(uc usecase.SessionUsecase, rg *gin.RouterGroup) *SessionHandler {
	return &SessionHandler{
		usecase: uc, 
		rg: *rg}
}
