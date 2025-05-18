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
		userRoutes.POST("/start", h.HandleStartSession)
		userRoutes.POST("/continue/:sessionId", h.HandleContinueSession)
		userRoutes.GET("/history/:sessionId", h.HandleGetSessionHistory)
		userRoutes.DELETE("/:sessionId", h.HandleDeleteSession)
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
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Bad Request",
			Error:   err.Error(),
		})
		return
	}

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

	sessionID, response, err := h.usecase.StartSession(c, userID, req.UserInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Internal Server Error",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Message: "Session started successfully",
		Data: dto.StartSessionResponse{
			SessionID: sessionID,
			Response:  response,
		},
	})
}

// HandleContinueSession handles requests to continue an existing session
func (h *SessionHandler) HandleContinueSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Bad Request",
			Error:   "session_id is required",
		})
		return
	}

	var req dto.ContinueSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Bad Request",
			Error:   err.Error(),
		})
		return
	}

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

	response, err := h.usecase.ContinueSession(c, userID, sessionID, req.UserInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Internal Server Error",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Message: "Session continued successfully",
		Data: dto.StartSessionResponse{
			SessionID: sessionID,
			Response:  response,
		},
	})
}

// HandleGetSessionHistory retrieves conversation history
func (h *SessionHandler) HandleGetSessionHistory(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Bad Request",
			Error:   "session_id is required",
		})
		return
	}

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

	// Default limit 20 messages
	limits := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Message: "Bad Request",
				Error:   "invalid limit parameter",
			})
			return
		}
		limits = limit
	}

	history, err := h.usecase.GetSessionHistory(c, userID, sessionID, limits)
	if err != nil {
		if err.Error() == "sesi tidak ditemukan atau tidak dapat diakses" {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Message: "Forbidden",
				Error:   err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Message: "Internal Server Error",
				Error:   err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Message: "Session history retrieved successfully",
		Data: dto.SessionHistoryResponse{
			SessionID: sessionID,
			Messages:  history,
		},
	})
}

// HandleGetUserSessions retrieves user session list
func (h *SessionHandler) HandleGetUserSessions(c *gin.Context) {
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

	sessions, err := h.usecase.GetUserSessions(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Internal Server Error",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Message: "User sessions retrieved successfully",
		Data:    sessions,
	})
}

func (h *SessionHandler) HandleDeleteSession(c *gin.Context) {
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

	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Bad Request",
			Error:   "session_id is required",
		})
		return
	}

	err := h.usecase.DeleteSession(c, userID, sessionID)
	if err != nil {
		if err.Error() == "session not found or not owned by user" {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Message: "Not Found",
				Error:   "session not found or not owned by user",
			})
		} else {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Message: "Internal Server Error",
				Error:   err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Message: "session deleted successfully",
		Data:    nil,
	})
}
