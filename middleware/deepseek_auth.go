package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// DeepseekAuthMiddleware validates the Deepseek API key from environment variables
func DeepseekAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := os.Getenv("DEEPSEEK_API_KEY")
		if apiKey == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Deepseek API key not configured",
			})
			c.Abort()
			return
		}

		// Store API key in context for later use
		c.Set("deepseek_api_key", apiKey)
		c.Next()
	}
}
