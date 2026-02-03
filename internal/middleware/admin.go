// Package middleware provides HTTP middleware for gin handlers.
package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/jaochai/ugc/pkg/response"
)

// AdminMiddleware creates a middleware that checks for admin role
func AdminMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := GetRoleFromContext(c)
		if !ok {
			logger.Debug("role not found in context")
			response.Forbidden(c, "admin access required")
			c.Abort()
			return
		}

		if role != "admin" {
			logger.Debug("non-admin user attempted admin access",
				zap.String("role", role),
			)
			response.Forbidden(c, "admin access required")
			c.Abort()
			return
		}

		c.Next()
	}
}
