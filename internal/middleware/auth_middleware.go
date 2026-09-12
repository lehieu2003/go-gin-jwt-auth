package middleware

import (
	"net/http"
	"strings"

	"github.com/example/go-gin-jwt-auth/internal/config"
	"github.com/example/go-gin-jwt-auth/internal/utils"
	"github.com/gin-gonic/gin"
)

const (
	ContextUserIDKey    = "userID"
	ContextUserEmailKey = "userEmail"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header format must be Bearer <token>",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := utils.ValidateAccessToken(tokenString, cfg.AccessTokenSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired access token",
			})
			c.Abort()
			return
		}

		// Store user details in Gin context
		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextUserEmailKey, claims.Email)

		c.Next()
	}
}
