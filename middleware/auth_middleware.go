package middleware

import (
	"log"
	"net/http"
	"pijar/utils/service"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthMiddlewareInterface interface {
	RequireToken(allowedRoles ...string) gin.HandlerFunc
}

type AuthMiddleware struct {
	jwtService service.JwtService
}

var _ AuthMiddlewareInterface = &AuthMiddleware{}

func NewAuthMiddleware(jwtService service.JwtService) *AuthMiddleware {
	return &AuthMiddleware{jwtService: jwtService}
}

func (a *AuthMiddleware) RequireToken(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := a.jwtService.VerifyToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		log.Println("claims.Role:", claims.Role)
		log.Println("allowedRoles:", allowedRoles)
		// cek role
		allowed := false
		for _, r := range allowedRoles {
			if strings.EqualFold(r, claims.Role) {
				allowed = true
				break
			}
		}
		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: role not allowed"})
			return
		}

		intUserID, err := strconv.Atoi(claims.UserId)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid user ID in token",
			})
			return
		}

		c.Set("userID", intUserID) // note lowercase key and int value
		c.Set("role", claims.Role)
		c.Next()
	}
}
