// Package middleware provides HTTP middleware for gin handlers.
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/jaochai/ugc/internal/service"
	"github.com/jaochai/ugc/pkg/response"
)

// Context keys for user data
const (
	ContextKeyUserID = "user_id"
	ContextKeyEmail  = "email"
)

// AuthMiddleware creates a middleware for JWT authentication
func AuthMiddleware(authService service.AuthService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "authorization header required")
			c.Abort()
			return
		}

		// Check Bearer token format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Unauthorized(c, "invalid authorization header format")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			logger.Debug("token validation failed", zap.Error(err))
			response.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		// Set user info in context
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyEmail, claims.Email)

		c.Next()
	}
}

// GetUserIDFromContext extracts the user ID from gin context
func GetUserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := c.Get(ContextKeyUserID)
	if !exists {
		return uuid.UUID{}, false
	}

	id, ok := userID.(uuid.UUID)
	return id, ok
}

// GetEmailFromContext extracts the email from gin context
func GetEmailFromContext(c *gin.Context) (string, bool) {
	email, exists := c.Get(ContextKeyEmail)
	if !exists {
		return "", false
	}

	emailStr, ok := email.(string)
	return emailStr, ok
}
