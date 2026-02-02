// Package middleware provides HTTP middleware for the UGC API.
package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// WebhookAuthConfig holds configuration for webhook authentication middleware.
type WebhookAuthConfig struct {
	Secret string
	Logger *zap.Logger
}

// WebhookAuthMiddleware validates webhook requests using token-based authentication.
// The token can be provided in the URL path parameter (:token) or in the X-Webhook-Token header.
// Since KIE API doesn't support HMAC signatures, we use a shared secret token.
func WebhookAuthMiddleware(cfg WebhookAuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth if no secret is configured (development mode)
		if cfg.Secret == "" {
			cfg.Logger.Warn("webhook authentication disabled - no WEBHOOK_SECRET configured")
			c.Next()
			return
		}

		// Get token from URL path parameter
		token := c.Param("token")
		if token == "" {
			// Also check header for flexibility
			token = c.GetHeader("X-Webhook-Token")
		}

		if token == "" {
			cfg.Logger.Warn("webhook request without token",
				zap.String("ip", c.ClientIP()),
				zap.String("path", c.Request.URL.Path),
			)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
			return
		}

		// Constant-time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(token), []byte(cfg.Secret)) != 1 {
			cfg.Logger.Warn("webhook request with invalid token",
				zap.String("ip", c.ClientIP()),
				zap.String("path", c.Request.URL.Path),
			)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
			return
		}

		c.Next()
	}
}
