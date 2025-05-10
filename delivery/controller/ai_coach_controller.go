package controller

import (
	"net/http"
	"strconv"

	"pijar/middleware"
	"pijar/model/dto"
	"pijar/usecase"

	"github.com/gin-gonic/gin"
)

type SessionHandler struct {
	usecase usecase.SessionUsecase
	rg      gin.RouterGroup
	aM      middleware.AuthMiddleware
}

func NewSessionHandler(uc usecase.SessionUsecase, rg *gin.RouterGroup, aM middleware.AuthMiddleware) *SessionHandler {
	return &SessionHandler{
		usecase: uc,
		rg:      *rg,
		aM:      aM,
	}
}

// Route defines API routes
func (h *SessionHandler) Route() {
	sessionGroup := h.rg.Group("/sessions")
	userRoutes := sessionGroup.Use(h.aM.RequireToken("USER", "ADMIN"))
	{
		userRoutes.POST("/start/:user_id", h.HandleStartSession)
		userRoutes.POST("/continue/:sessionId/:user_id", h.HandleContinueSession)
		userRoutes.GET("/history/:sessionId/:user_id", h.HandleGetSessionHistory)
	}

	adminRoutes := sessionGroup.Use(h.aM.RequireToken("ADMIN"))
	{
		adminRoutes.GET("/user/:user_id", h.HandleGetUserSessions)
	}
}

// HandleStartSession handles requests to start a new session
func (h *SessionHandler) HandleStartSession(c *gin.Context) {
	var req dto.CoachRequest
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

	c.JSON(http.StatusOK, dto.StartSessionResponse{
		SessionID: sessionID,
		Response:  response,
	})
}

// HandleContinueSession handles requests to continue an existing session
func (h *SessionHandler) HandleContinueSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	var req dto.ContinueSessionRequest
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

	c.JSON(http.StatusOK, dto.StartSessionResponse{
		SessionID: sessionID,
		Response:  response,
	})
}

// HandleGetSessionHistory retrieves conversation history
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

	// Default limit 20 messages
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

	c.JSON(http.StatusOK, dto.SessionHistoryResponse{
		SessionID: sessionID,
		Messages:  history,
	})
}

// HandleGetUserSessions retrieves user session list
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


