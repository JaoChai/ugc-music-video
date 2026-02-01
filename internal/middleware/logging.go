// Package middleware provides HTTP middleware for gin handlers.
package middleware

import (
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/jaochai/ugc/pkg/response"
)

// Context key for request ID
const (
	ContextKeyRequestID = "request_id"
)

// GetRequestID extracts the request ID from gin context
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(ContextKeyRequestID); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// RequestIDMiddleware generates a UUID for each request and sets it in context and response header
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate a new UUID for the request
		requestID := uuid.New().String()

		// Set in context
		c.Set(ContextKeyRequestID, requestID)

		// Add to response header
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// responseWriter wraps gin.ResponseWriter to capture response size
type responseWriter struct {
	gin.ResponseWriter
	size int
}

func (w *responseWriter) Write(data []byte) (int, error) {
	n, err := w.ResponseWriter.Write(data)
	w.size += n
	return n, err
}

func (w *responseWriter) WriteString(s string) (int, error) {
	n, err := w.ResponseWriter.WriteString(s)
	w.size += n
	return n, err
}

// LoggingMiddleware logs request and response information using zap structured logging
func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip logging for health endpoint
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		// Start timer
		start := time.Now()

		// Get request info
		method := c.Request.Method
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()
		requestID := GetRequestID(c)

		// Wrap response writer to capture size
		rw := &responseWriter{ResponseWriter: c.Writer, size: 0}
		c.Writer = rw

		// Log request start
		logger.Info("request started",
			zap.String("request_id", requestID),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("client_ip", clientIP),
			zap.String("user_agent", userAgent),
		)

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get response info
		statusCode := c.Writer.Status()
		bodySize := rw.size

		// Determine log level based on status code
		logFields := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("client_ip", clientIP),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.Int("body_size", bodySize),
		}

		// Log with appropriate level based on status code
		switch {
		case statusCode >= 500:
			logger.Error("request completed with server error", logFields...)
		case statusCode >= 400:
			logger.Warn("request completed with client error", logFields...)
		default:
			logger.Info("request completed", logFields...)
		}
	}
}

// RecoveryMiddleware recovers from panics and logs the error with stack trace
func RecoveryMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get request ID for correlation
				requestID := GetRequestID(c)

				// Get stack trace
				stack := string(debug.Stack())

				// Log panic with stack trace
				logger.Error("panic recovered",
					zap.String("request_id", requestID),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("client_ip", c.ClientIP()),
					zap.Any("error", err),
					zap.String("stack_trace", stack),
				)

				// Abort with internal server error
				c.Abort()
				response.InternalServerError(c, "internal server error")
			}
		}()

		c.Next()
	}
}
