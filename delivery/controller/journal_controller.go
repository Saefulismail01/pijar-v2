package controller

import (
	dbsql "database/sql"
	"net/http"
	"pijar/middleware"
	"pijar/model"
	"pijar/model/dto"
	"pijar/usecase"
	"pijar/utils/service"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type JournalController struct {
	usecase usecase.JournalUsecase
	rg      *gin.RouterGroup
	aM      middleware.AuthMiddleware
}

func NewJournalController(usecase usecase.JournalUsecase, rg *gin.RouterGroup, aM middleware.AuthMiddleware) *JournalController {
	return &JournalController{
		usecase: usecase,
		rg:      rg,
		aM:      aM,
	}
}

func (c *JournalController) Route() {

	journalGroup := c.rg.Group("/journals")
	userRoutes := journalGroup.Use(c.aM.RequireToken("USER", "ADMIN"))
	{
		userRoutes.POST("/", c.CreateJournal)
		userRoutes.GET("/user", c.GetJournalsByUserID)
		userRoutes.PUT("/:journalID", c.UpdateJournal)
		userRoutes.DELETE("/:journalID", c.DeleteJournal)
		userRoutes.GET("/export", c.ExportJournalsToPDF)
	}

	adminRoutes := journalGroup.Use(c.aM.RequireToken("ADMIN"))
	{
		adminRoutes.GET("", c.GetAllJournals)
		adminRoutes.GET("/:journalID", c.GetJournalByID) // Admin Only

	}
}

func (c *JournalController) CreateJournal(ctx *gin.Context) {
	// get user ID from jwt body
	val, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.Response{
			Message: "Authentication required",
		})
		return
	}
	userID, ok := val.(int)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Message: "Invalid user identity in context",
		})
		return
	}

	var journal model.Journal
	if err := ctx.ShouldBindJSON(&journal); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Pastikan UserID ada dan valid
	journal.UserID = userID

	// Pastikan Judul, Isi, dan Perasaan ada
	if journal.Judul == "" || journal.Isi == "" || journal.Perasaan == "" {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Missing required fields",
			Error:   "Judul, Isi, dan Perasaan wajib diisi",
		})
		return
	}

	// Set waktu pembuatan
	journal.CreatedAt = time.Now()

	if err := c.usecase.Create(ctx, &journal); err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Failed to create journal",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, journal)
}

func (c *JournalController) GetAllJournals(ctx *gin.Context) {
	journals, err := c.usecase.FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Failed to fetch journals",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, journals)
}

func (c *JournalController) GetJournalsByUserID(ctx *gin.Context) {
	// get user ID from jwt body
	val, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.Response{
			Message: "Authentication required",
		})
		return
	}
	userID, ok := val.(int)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Message: "Invalid user identity in context",
		})
		return
	}

	journals, err := c.usecase.FindByUserID(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Failed to fetch journals",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, journals)
}

func (c *JournalController) GetJournalByID(ctx *gin.Context) {
	journalID, err := strconv.Atoi(ctx.Param("journalID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Invalid journal ID",
			Error:   "invalid journal ID",
		})
		return
	}

	journal, err := c.usecase.FindByID(ctx, journalID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Failed to fetch journals",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, journal)
}

func (c *JournalController) UpdateJournal(ctx *gin.Context) {
	journalID, err := strconv.Atoi(ctx.Param("journalID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Invalid journal ID",
			Error:   "invalid journal ID",
		})
		return
	}

	// Get user ID from JWT token
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Message: "Authentication required",
			Error:   "Invalid token",
		})
		return
	}
	userIDInt, ok := userID.(int)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Invalid user identity",
			Error:   "Failed to parse user ID from token",
		})
		return
	}

	// Get existing journal to check ownership
	existingJournal, err := c.usecase.FindByID(ctx, journalID)
	if err != nil {
		if err == dbsql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{
				Message: "Journal not found",
				Error:   "journal not found",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Failed to fetch journal",
			Error:   err.Error(),
		})
		return
	}

	// Verify journal ownership
	if existingJournal.UserID != userIDInt {
		ctx.JSON(http.StatusForbidden, dto.ErrorResponse{
			Message: "Forbidden",
			Error:   "Cannot update journal that doesn't belong to you",
		})
		return
	}

	var journal model.Journal
	if err := ctx.ShouldBindJSON(&journal); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Set user_id from existing journal and journal ID
	journal.UserID = existingJournal.UserID
	journal.ID = journalID

	if err := c.usecase.Update(ctx, &journal); err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Failed to update journal",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, journal)
}

func (c *JournalController) DeleteJournal(ctx *gin.Context) {
	// Get journal ID from URL parameter
	journalID, err := strconv.Atoi(ctx.Param("journalID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Invalid journal ID format",
			Error:   "invalid_journal_id",
		})
		return
	}

	// Get user ID from context (assuming it's set by auth middleware)
	userID, userIDExists := ctx.Get("userID")
	if !userIDExists {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Invalid user ID format",
			Error:   "invalid_user_id",
		})
		return
	}
	// First, get the journal to check if it exists and belongs to the user
	journal, err := c.usecase.FindByID(ctx, journalID)
	if err != nil {
		if err == dbsql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{
				Message: "Journal not found",
				Error:   "not_found",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Failed to fetch journal",
			Error:   "internal_server_error",
		})
		return
	}

	// Verify ownership (assuming journal has a UserID field)
	if journal.UserID != userID {
		ctx.JSON(http.StatusForbidden, dto.ErrorResponse{
			Message: "You don't have permission to delete this journal",
			Error:   "forbidden",
		})
		return
	}

	// Proceed with deletion
	if err := c.usecase.Delete(ctx, journalID); err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Failed to delete journal",
			Error:   "internal_server_error",
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Message: "Journal deleted successfully",
	})
}

func (c *JournalController) ExportJournalsToPDF(ctx *gin.Context) {
	// get user ID from jwt body
	val, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.Response{
			Message: "Authentication required",
		})
		return
	}
	userID, ok := val.(int)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Message: "Invalid user identity in context",
		})
		return
	}

	journals, err := c.usecase.FindByUserID(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Failed to fetch journals",
			Error:   err.Error(),
		})
		return
	}

	if len(journals) == 0 {
		ctx.JSON(http.StatusNotFound, dto.ErrorResponse{
			Message: "No journals found for this user",
			Error:   "no journals found for this user",
		})
		return
	}

	// Convert journals to PDF format and generate PDF
	pdfJournals := service.ConvertToPDFFormat(journals)
	pdf, err := service.GenerateJournalsPDF(pdfJournals)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Failed to generate PDF",
			Error:   "failed to generate PDF",
		})
		return
	}

	// Set response headers
	ctx.Header("Content-Type", "application/pdf")
	ctx.Header("Content-Disposition", "attachment; filename=journal_export_"+time.Now().Format("20060102_150405")+".pdf")

	// Output PDF to response writer
	if err := pdf.Output(ctx.Writer); err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Failed to generate PDF",
			Error:   "failed to generate PDF",
		})
		return
	}
}
