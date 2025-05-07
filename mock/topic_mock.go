package mock

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	DummyUserID = 1 // Static user ID for testing
)

// Model
type TopicUser struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id" binding:"required"`
	Preference string    `json:"preference" binding:"required"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Mock Storage for Topics
type topicStorage struct {
	mu     sync.RWMutex
	topics map[int]TopicUser
	lastID int
}

var topicStore = &topicStorage{
	topics: make(map[int]TopicUser),
	lastID: 0,
}

// Mock Handler
type TopicMockHandler struct{}

func NewTopicMockHandler() *TopicMockHandler {
	return &TopicMockHandler{}
}

func (h *TopicMockHandler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.POST("/topics", h.CreateTopic)
		api.GET("/topics", h.GetAllTopics)
		api.GET("/topics/user/:userID", h.GetTopicByUserID)
		api.PUT("/topics/:id", h.UpdateTopic)
		api.DELETE("/topics/:id", h.DeleteTopic)
	}
}

func (h *TopicMockHandler) CreateTopic(c *gin.Context) {
	var topic TopicUser
	if err := c.ShouldBindJSON(&topic); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Format request tidak valid",
		})
		return
	}

	if topic.UserID != DummyUserID {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": fmt.Sprintf("Unauthorized: hanya bisa menggunakan user_id %d untuk testing", DummyUserID),
		})
		return
	}

	topicStore.mu.Lock()
	defer topicStore.mu.Unlock()

	topicStore.lastID++
	now := time.Now().UTC()

	topic.ID = topicStore.lastID
	topic.CreatedAt = now
	topic.UpdatedAt = now

	topicStore.topics[topic.ID] = topic

	c.JSON(http.StatusCreated, gin.H{
		"message": "Topic user berhasil dibuat",
		"data": gin.H{
			"id": topic.ID,
		},
	})
}

func (h *TopicMockHandler) GetAllTopics(c *gin.Context) {
	topicStore.mu.RLock()
	defer topicStore.mu.RUnlock()

	topics := make([]TopicUser, 0, len(topicStore.topics))
	for _, topic := range topicStore.topics {
		topics = append(topics, topic)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil semua topic user",
		"data":    topics,
	})
}

func (h *TopicMockHandler) GetTopicByUserID(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID user tidak valid",
		})
		return
	}

	topicStore.mu.RLock()
	defer topicStore.mu.RUnlock()

	var userTopics []TopicUser
	for _, topic := range topicStore.topics {
		if topic.UserID == userID {
			userTopics = append(userTopics, topic)
		}
	}

	if len(userTopics) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("topic user dengan ID user %d tidak ditemukan", userID),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil topic user",
		"data":    userTopics,
	})
}

func (h *TopicMockHandler) UpdateTopic(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID topic tidak valid",
		})
		return
	}

	var updateData TopicUser
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Format request tidak valid",
		})
		return
	}

	topicStore.mu.Lock()
	defer topicStore.mu.Unlock()

	topic, exists := topicStore.topics[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("topic user dengan ID %d tidak ditemukan", id),
		})
		return
	}

	topic.Preference = updateData.Preference
	topic.UpdatedAt = time.Now().UTC()
	topicStore.topics[id] = topic

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Topic user dengan ID %d berhasil diperbarui", id),
	})
}

func (h *TopicMockHandler) DeleteTopic(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID topic tidak valid",
		})
		return
	}

	topicStore.mu.Lock()
	defer topicStore.mu.Unlock()

	if _, exists := topicStore.topics[id]; !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("topic user dengan ID %d tidak ditemukan", id),
		})
		return
	}

	delete(topicStore.topics, id)

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Topic user dengan ID %d berhasil dihapus", id),
	})
}
