// Package middleware provides HTTP middleware for the UGC API.
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RateLimitConfig holds configuration for rate limiting middleware.
type RateLimitConfig struct {
	RedisClient *redis.Client
	RPS         int    // Requests per second
	Burst       int    // Burst size (max requests in window)
	KeyPrefix   string // Redis key prefix
	Logger      *zap.Logger
}

// RateLimitMiddleware implements sliding window rate limiting using Redis.
// It limits requests per IP address.
func RateLimitMiddleware(cfg RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip if Redis client is not configured
		if cfg.RedisClient == nil {
			cfg.Logger.Warn("rate limiting disabled - no Redis client configured")
			c.Next()
			return
		}

		// Use IP address as rate limit key
		key := fmt.Sprintf("%s:webhook:ratelimit:%s", cfg.KeyPrefix, c.ClientIP())

		ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Millisecond)
		defer cancel()

		// Check rate limit
		allowed, err := checkRateLimit(ctx, cfg.RedisClient, key, cfg.Burst)
		if err != nil {
			// Fail open for availability - log error but allow request
			cfg.Logger.Error("rate limit check failed",
				zap.Error(err),
				zap.String("ip", c.ClientIP()),
			)
			c.Next()
			return
		}

		if !allowed {
			cfg.Logger.Warn("rate limit exceeded",
				zap.String("ip", c.ClientIP()),
				zap.String("path", c.Request.URL.Path),
			)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"message": "rate limit exceeded",
			})
			return
		}

		c.Next()
	}
}

// checkRateLimit uses Redis sorted set for sliding window rate limiting.
// Returns true if request is allowed, false if rate limit exceeded.
func checkRateLimit(ctx context.Context, client *redis.Client, key string, burst int) (bool, error) {
	now := time.Now().UnixMilli()
	windowMs := int64(1000) // 1 second window

	pipe := client.Pipeline()

	// Remove old entries outside the window
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", now-windowMs))

	// Count current entries in the window
	countCmd := pipe.ZCard(ctx, key)

	// Add current request with timestamp as score and member
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: fmt.Sprintf("%d-%d", now, time.Now().UnixNano())})

	// Set expiry on the key (2x window to ensure cleanup)
	pipe.Expire(ctx, key, time.Duration(windowMs)*time.Millisecond*2)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("redis pipeline exec failed: %w", err)
	}

	count := countCmd.Val()
	return count < int64(burst), nil
}
