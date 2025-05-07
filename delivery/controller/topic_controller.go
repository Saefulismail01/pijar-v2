package controller

import (
	"fmt"
	"net/http"
	"pijar/model"
	"pijar/usecase"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TopicController interface {
	CreateTopic(c *gin.Context)
	GetAllTopics(c *gin.Context)
	GetTopicByID(c *gin.Context)
	UpdateTopic(c *gin.Context)
	DeleteTopic(c *gin.Context)
	RegisterRoutes(rg *gin.RouterGroup, protected *gin.RouterGroup)
}

type topicControllerImpl struct {
	topicUsecase usecase.TopicUserUsecase
}

func NewTopicController(tu usecase.TopicUserUsecase) TopicController {
	return &topicControllerImpl{topicUsecase: tu}
}

func (tc *topicControllerImpl) RegisterRoutes(rg *gin.RouterGroup, protected *gin.RouterGroup) {
	// Public routes
	rg.GET("/topics", tc.GetAllTopics)
	rg.GET("/topics/:id", tc.GetTopicByID)

	// Protected routes
	protected.POST("/topics", tc.CreateTopic)
	protected.PUT("/topics/:id", tc.UpdateTopic)
	protected.DELETE("/topics/:id", tc.DeleteTopic)
}

func (tc *topicControllerImpl) CreateTopic(c *gin.Context) {
	var topicUser model.TopicUser
	if err := c.ShouldBindJSON(&topicUser); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Format request tidak valid",
		})
		return
	}

	id, err := tc.topicUsecase.CreateTopicUser(c.Request.Context(), topicUser.UserID, topicUser.Preference)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal membuat topic user: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Topic user berhasil dibuat",
		"data": gin.H{
			"id": id,
		},
	})
}

func (tc *topicControllerImpl) GetAllTopics(c *gin.Context) {
	topics, err := tc.topicUsecase.GetAllTopicUsers(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil daftar topic user: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil semua topic user",
		"data":    topics,
	})
}

func (c *topicControllerImpl) GetTopicByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "ID topic tidak valid",
		})
		return
	}

	topics, err := c.topicUsecase.GetTopicByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if len(topics) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Topic dengan ID %d tidak ditemukan", id),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil topic by ID",
		"data":    topics,
	})
}

func (tc *topicControllerImpl) UpdateTopic(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "ID topic tidak valid",
		})
		return
	}

	var topicUser model.TopicUser
	if err := c.ShouldBindJSON(&topicUser); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Format request tidak valid",
		})
		return
	}

	err = tc.topicUsecase.UpdateTopicUser(c.Request.Context(), id, topicUser.Preference)
	if err != nil {
		if err.Error() == fmt.Sprintf("topic user dengan ID %d tidak ditemukan", id) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal memperbarui topic user: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Topic user dengan ID %d berhasil diperbarui", id),
	})
}

func (tc *topicControllerImpl) DeleteTopic(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "ID topic tidak valid",
		})
		return
	}

	err = tc.topicUsecase.DeleteTopicUser(c.Request.Context(), id)
	if err != nil {
		if err.Error() == fmt.Sprintf("topic user dengan ID %d tidak ditemukan", id) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal menghapus topic user: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Topic user dengan ID %d berhasil dihapus", id),
	})
}
