package controller

import (
	"fmt"
	"net/http"
	"pijar/model"
	"pijar/usecase"
	"strconv"

	"github.com/gin-gonic/gin"
)

type JournalController struct {
	usecase     usecase.JournalUsecase
	RouterGroup *gin.RouterGroup
}

func NewJournalController(usecase usecase.JournalUsecase, rg *gin.RouterGroup) *JournalController {
	return &JournalController{
		usecase:     usecase,
		RouterGroup: rg,
	}
}

func (c *JournalController) Route() {
	c.RouterGroup.POST("/journals/", c.CreateJournal)
	c.RouterGroup.GET("/journals/user/:userID", c.GetJournals)
	c.RouterGroup.GET("/journals/:journalID", c.GetJournalByID)
	c.RouterGroup.PUT("/journals/:journalID", c.UpdateJournal)
	c.RouterGroup.DELETE("/journals/:journalID", c.DeleteJournal)
}

func (c *JournalController) CreateJournal(ctx *gin.Context) {
	var journal model.Journal
	if err := ctx.ShouldBindJSON(&journal); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if journal.UserID <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "UserID is required"})
		return
	}

	if journal.Judul == "" || journal.Isi == "" || journal.Perasaan == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Judul, Isi, dan Perasaan wajib diisi"})
		return
	}

	if err := c.usecase.Create(ctx, &journal); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, journal)
}

func (c *JournalController) GetJournals(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("userID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	journals, err := c.usecase.FindAll(ctx, userID)
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

	var journal model.Journal
	if err := ctx.ShouldBindJSON(&journal); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

	ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Journal %d deleted successfully", journalID)})
}
