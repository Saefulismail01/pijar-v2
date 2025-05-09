package controller

import (
	"net/http"
	"pijar/middleware"
	"pijar/model"
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
<<<<<<< HEAD

	journalGroup := c.rg.Group("/journals")
	userRoutes := journalGroup.Use(c.aM.RequireToken("USER", "ADMIN"))
	{
		userRoutes.POST("/", c.CreateJournal)
		userRoutes.GET("/user/:userID", c.GetJournalsByUserID)
		userRoutes.PUT("/:journalID", c.UpdateJournal)
		userRoutes.DELETE("/:journalID", c.DeleteJournal)
		userRoutes.GET("/user/:userID/export", c.ExportJournalsToPDF)
	}

	adminRoutes := journalGroup.Use(c.aM.RequireToken("ADMIN"))
	{
		adminRoutes.GET("", c.GetAllJournals)
		adminRoutes.GET("/:journalID", c.GetJournalByID) // Khusus Admin

	}
=======
	c.RouterGroup.POST("/journals", c.CreateJournal)
	c.RouterGroup.GET("/journals", c.GetAllJournals)
	c.RouterGroup.GET("/journals/user/:userID", c.GetJournalsByUserID)
	c.RouterGroup.GET("/journals/:journalID", c.GetJournalByID)
	c.RouterGroup.PUT("/journals/:journalID", c.UpdateJournal)
	c.RouterGroup.DELETE("/journals/:journalID", c.DeleteJournal)
	c.RouterGroup.GET("/journals/user/:userID/export", c.ExportJournalsToPDF)
>>>>>>> a527605 (feat(journal): update fitur journal)
}

func (c *JournalController) CreateJournal(ctx *gin.Context) {
	var journal model.Journal
	if err := ctx.ShouldBindJSON(&journal); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Pastikan UserID ada dan valid
	if journal.UserID <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "UserID is required"})
		return
	}

	// Pastikan Judul, Isi, dan Perasaan ada
	if journal.Judul == "" || journal.Isi == "" || journal.Perasaan == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Judul, Isi, dan Perasaan wajib diisi"})
		return
	}

	// Set waktu pembuatan
	journal.CreatedAt = time.Now()

	if err := c.usecase.Create(ctx, &journal); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, journal)
}

func (c *JournalController) GetAllJournals(ctx *gin.Context) {
	journals, err := c.usecase.FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, journals)
}

func (c *JournalController) GetJournalsByUserID(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("userID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	journals, err := c.usecase.FindByUserID(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, journals)
}

func (c *JournalController) GetJournalByID(ctx *gin.Context) {
	journalID, err := strconv.Atoi(ctx.Param("journalID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid journal ID"})
		return
	}

	journal, err := c.usecase.FindByID(ctx, journalID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, journal)
}

func (c *JournalController) UpdateJournal(ctx *gin.Context) {
    journalID, err := strconv.Atoi(ctx.Param("journalID"))
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid journal ID"})
        return
    }

<<<<<<< HEAD
	// Get existing journal to get the correct user_id
	existingJournal, err := c.usecase.FindByID(ctx, journalID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var journal model.Journal
	if err := ctx.ShouldBindJSON(&journal); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set user_id from existing journal
	journal.UserID = existingJournal.UserID
	journal.ID = journalID
=======
    // Get existing journal to get the correct user_id
    existingJournal, err := c.usecase.FindByID(ctx, journalID)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    var journal model.Journal
    if err := ctx.ShouldBindJSON(&journal); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
>>>>>>> a527605 (feat(journal): update fitur journal)

    // Set user_id from existing journal
    journal.UserID = existingJournal.UserID
    journal.ID = journalID

    if err := c.usecase.Update(ctx, &journal); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    ctx.JSON(http.StatusOK, journal)
}

func (c *JournalController) DeleteJournal(ctx *gin.Context) {
	journalID, err := strconv.Atoi(ctx.Param("journalID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid journal ID"})
		return
	}

	if err := c.usecase.Delete(ctx, journalID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (c *JournalController) ExportJournalsToPDF(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("userID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	journals, err := c.usecase.FindByUserID(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(journals) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "no journals found for this user"})
		return
	}

	// Convert journals to PDF format and generate PDF
	pdfJournals := service.ConvertToPDFFormat(journals)
	pdf, err := service.GenerateJournalsPDF(pdfJournals)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate PDF"})
		return
	}

	// Set response headers
	ctx.Header("Content-Type", "application/pdf")
	ctx.Header("Content-Disposition", "attachment; filename=journal_export_"+time.Now().Format("20060102_150405")+".pdf")

	// Output PDF to response writer
	if err := pdf.Output(ctx.Writer); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate PDF"})
		return
	}
}
