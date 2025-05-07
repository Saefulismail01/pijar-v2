package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"pijar/model"
	"pijar/usecase"
)

type SessionHandler struct {
	usecase usecase.SessionUsecase
	rg      gin.RouterGroup
}

type CoachRequest struct {
	UserInput string `json:"user_input"`
}

type StartSessionResponse struct {
	SessionID string `json:"session_id"`
	Response  string `json:"response"`
}

type ContinueSessionRequest struct {
	UserInput string `json:"user_input"`
}

type SessionHistoryResponse struct {
	SessionID string          `json:"session_id"`
	Messages  []model.Message `json:"messages"`
}

// HandleStartSession menangani permintaan untuk memulai sesi baru
func (h *SessionHandler) HandleStartSession(c *gin.Context) {
	var req CoachRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	sessionID, response, err := h.usecase.StartSession(userID, req.UserInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, StartSessionResponse{
		SessionID: sessionID,
		Response:  response,
	})
}

// HandleContinueSession menangani permintaan untuk melanjutkan sesi yang ada
func (h *SessionHandler) HandleContinueSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	var req ContinueSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	response, err := h.usecase.ContinueSession(userID, sessionID, req.UserInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, StartSessionResponse{
		SessionID: sessionID,
		Response:  response,
	})
}

// HandleGetSessionHistory mengambil riwayat percakapan
func (h *SessionHandler) HandleGetSessionHistory(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	// Default limit 20 pesan
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
			return
		}
	}

	history, err := h.usecase.GetSessionHistory(userID, sessionID, limit)
	if err != nil {
		if err.Error() == "sesi tidak ditemukan atau tidak dapat diakses" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, SessionHistoryResponse{
		SessionID: sessionID,
		Messages:  history,
	})
}

// HandleGetUserSessions mengambil daftar sesi pengguna
func (h *SessionHandler) HandleGetUserSessions(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	sessions, err := h.usecase.GetUserSessions(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
	})
}

// Route mendefinisikan rute-rute API
func (h *SessionHandler) Route() {
	h.rg.POST("/sessions/start/:user_id", h.HandleStartSession)
	h.rg.POST("/sessions/continue/:sessionId/:user_id", h.HandleContinueSession)
	h.rg.GET("/sessions/history/:sessionId/:user_id", h.HandleGetSessionHistory)
	h.rg.GET("/sessions/user/:user_id", h.HandleGetUserSessions)
}

func NewSessionHandler(uc usecase.SessionUsecase, rg *gin.RouterGroup) *SessionHandler {
	return &SessionHandler{
		usecase: uc,
		rg:      *rg,
	}
}

