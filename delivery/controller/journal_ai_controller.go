package controller

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"pijar/middleware"
	"pijar/model"
	"pijar/usecase"

	"github.com/gin-gonic/gin"
)

type JournalAIController struct {
	aiUsecase usecase.JournalAIUsecase
	rg        *gin.RouterGroup
	authMdw   *middleware.AuthMiddleware
}

func NewJournalAIController(
	aiUsecase usecase.JournalAIUsecase,
	rg *gin.RouterGroup,
	authMdw middleware.AuthMiddleware,
) *JournalAIController {
	controller := &JournalAIController{
		aiUsecase: aiUsecase,
		rg:        rg,
		authMdw:   &authMdw,
	}

	return controller
}

func (c *JournalAIController) Route() {
	journalAPI := c.rg.Group("/journals-ai")

	userRoutes := journalAPI.Use(c.authMdw.RequireToken("USER", "ADMIN"))
	{
		// Single analysis
		userRoutes.POST("/analyze", c.analyzeJournal)
		userRoutes.GET("/:id/analysis", c.getJournalAnalysis)
		userRoutes.PUT("/:id/reanalyze", c.reanalyzeJournal)

		// Multiple analyses
		userRoutes.GET("/analyses", c.getUserAnalyses)
		userRoutes.GET("/analyses-with-entries", c.getAnalysisWithJournal)

		// Trend analysis
		userRoutes.POST("/trend-analysis", c.generateTrendAnalysis)
		userRoutes.GET("/trends", c.getTrendHistory)

		// Charts & visualization
		userRoutes.GET("/sentiment-chart", c.getSentimentChart)
	}
}

// analyzeJournal handles journal analysis request
func (c *JournalAIController) analyzeJournal(ctx *gin.Context) {
	var req model.AnalysisRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Get user ID from context
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized - userID not found in context"})
		return
	}
	req.UserID = userID.(int)

	response, err := c.aiUsecase.AnalyzeJournal(ctx.Request.Context(), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// getJournalAnalysis retrieves analysis for a specific journal
func (c *JournalAIController) getJournalAnalysis(ctx *gin.Context) {
	journalID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid journal ID: must be a number"})
		return
	}

	// Get user ID from context
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized - userID not found in context"})
		return
	}

	// Convert userID to int (handling float64 from JWT)
	var userIDInt int
	switch v := userID.(type) {
	case float64:
		userIDInt = int(v)
	case int:
		userIDInt = v
	default:
		log.Printf("user_id has unexpected type: %T", userID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id format"})
		return
	}

	analysis, err := c.aiUsecase.GetJournalAnalysis(ctx.Request.Context(), journalID, userIDInt)
	if err != nil {
		log.Printf("Error getting journal analysis: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get journal analysis: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, analysis)
}

// reanalyzeJournal triggers reanalysis of a journal
func (c *JournalAIController) reanalyzeJournal(ctx *gin.Context) {
	journalID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid journal ID: must be a number"})
		return
	}

	// Get user ID from context
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized - userID not found in context"})
		return
	}

	// Convert userID to int (handling float64 from JWT)
	var userIDInt int
	switch v := userID.(type) {
	case float64:
		userIDInt = int(v)
	case int:
		userIDInt = v
	default:
		log.Printf("user_id has unexpected type: %T", userID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id format"})
		return
	}

	response, err := c.aiUsecase.ReanalyzeJournal(ctx.Request.Context(), journalID, userIDInt)
	if err != nil {
		log.Printf("Error reanalyzing journal: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reanalyze journal: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": response})
}

// getUserAnalyses retrieves all analyses for the current user
func (c *JournalAIController) getUserAnalyses(ctx *gin.Context) {
	// Debug log all context keys and values
	log.Printf("=== Debug: Start of getUserAnalyses ===")
	for k, v := range ctx.Keys {
		log.Printf("Context key: %s, value: %v, type: %T", k, v, v)
	}

	// Get user ID from context
	userID, exists := ctx.Get("userID")
	if !exists {
		// Check if it's using a different case
		for _, key := range []string{"userid", "UserID", "USERID", "UserId"} {
			if val, ok := ctx.Get(key); ok {
				log.Printf("Found user ID with key '%s': %v (type: %T)", key, val, val)
				userID = val
				exists = true
				break
			}
		}

		if !exists {
			log.Printf("Error: userID not found in context. Available keys: %v", ctx.Keys)
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized - userID not found in context",
				"debug": fmt.Sprintf("available keys: %v", ctx.Keys),
			})
			return
		}
	}

	log.Printf("Found userID: %v (type: %T)", userID, userID)

	// Convert userID to int (handling float64 from JWT)
	var userIDInt int
	switch v := userID.(type) {
	case float64:
		userIDInt = int(v)
	case int:
		userIDInt = v
	default:
		log.Printf("user_id has unexpected type: %T", userID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id format"})
		return
	}

	log.Printf("Fetching analyses for user ID: %d", userIDInt)

	limitStr := ctx.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10 // Default limit
	}

	analyses, err := c.aiUsecase.GetUserAnalyses(ctx.Request.Context(), userIDInt, limit)
	if err != nil {
		log.Printf("Error fetching analyses: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch analyses: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": analyses})
}

// getAnalysisWithJournal retrieves analyses with journal entries
func (c *JournalAIController) getAnalysisWithJournal(ctx *gin.Context) {
	// Get user ID from context
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized - userID not found in context"})
		return
	}

	// Convert userID to int (handling float64 from JWT)
	var userIDInt int
	switch v := userID.(type) {
	case float64:
		userIDInt = int(v)
	case int:
		userIDInt = v
	default:
		log.Printf("user_id has unexpected type: %T", userID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id format"})
		return
	}

	// Get limit from query parameter
	limit := 10 // default limit
	if limitStr := ctx.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	analyses, err := c.aiUsecase.GetAnalysisWithJournal(ctx.Request.Context(), userIDInt, limit)
	if err != nil {
		log.Printf("Error getting analyses with journal: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch analyses: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": analyses})
}

// generateTrendAnalysis generates trend analysis for the user
func (c *JournalAIController) generateTrendAnalysis(ctx *gin.Context) {
	// Get user ID from context
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized - userID not found in context"})
		return
	}

	// Convert userID to int (handling float64 from JWT)
	var userIDInt int
	switch v := userID.(type) {
	case float64:
		userIDInt = int(v)
	case int:
		userIDInt = v
	default:
		log.Printf("user_id has unexpected type: %T", userID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id format"})
		return
	}

	var req struct {
		PeriodType string `json:"period_type"`
		Days       int    `json:"days"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.PeriodType == "" {
		req.PeriodType = "weekly"
	}

	if req.Days <= 0 {
		req.Days = 30 // Default to 30 days
	}

	response, err := c.aiUsecase.GenerateTrendAnalysis(ctx.Request.Context(), userIDInt, req.PeriodType, req.Days)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// getTrendHistory retrieves trend analysis history
func (c *JournalAIController) getTrendHistory(ctx *gin.Context) {
	// Get user ID from context
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized - userID not found in context"})
		return
	}

	// Convert userID to int (handling float64 from JWT)
	var userIDInt int
	switch v := userID.(type) {
	case float64:
		userIDInt = int(v)
	case int:
		userIDInt = v
	default:
		log.Printf("user_id has unexpected type: %T", userID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id format"})
		return
	}

	periodType := ctx.DefaultQuery("period_type", "weekly")

	trends, err := c.aiUsecase.GetTrendHistory(ctx.Request.Context(), userIDInt, periodType)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": trends})
}

// getSentimentChart retrieves sentiment data for charting
func (c *JournalAIController) getSentimentChart(ctx *gin.Context) {
	// Get user ID from context
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized - userID not found in context"})
		return
	}

	// Convert userID to int (handling float64 from JWT)
	var userIDInt int
	switch v := userID.(type) {
	case float64:
		userIDInt = int(v)
	case int:
		userIDInt = v
	default:
		log.Printf("user_id has unexpected type: %T", userID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id format"})
		return
	}

	days := 30 // Default to 30 days
	if daysStr := ctx.Query("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	chartData, err := c.aiUsecase.GetSentimentChart(ctx.Request.Context(), userIDInt, days)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": chartData})
}
