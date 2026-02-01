// Package middleware provides HTTP middleware for gin handlers.
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/cors"
)

// CORSConfig holds configuration for CORS middleware.
type CORSConfig struct {
	// AllowOrigins is a list of origins that may access the resource.
	AllowOrigins []string

	// AllowMethods is a list of methods the client is allowed to use.
	AllowMethods []string

	// AllowHeaders is a list of headers the client is allowed to use.
	AllowHeaders []string

	// ExposeHeaders is a list of headers that are safe to expose to the API of a CORS response.
	ExposeHeaders []string

	// AllowCredentials indicates whether the request can include user credentials.
	AllowCredentials bool

	// MaxAge indicates how long the results of a preflight request can be cached (in seconds).
	MaxAge int
}

// DefaultCORSConfig returns a CORSConfig with default values suitable for development.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://localhost:5173",
		},
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
		},
		ExposeHeaders: []string{
			"Content-Length",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}
}

// ProductionCORSConfig returns a strict CORSConfig suitable for production.
// Only the specified allowedOrigins will be permitted.
func ProductionCORSConfig(allowedOrigins []string) CORSConfig {
	return CORSConfig{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
		},
		ExposeHeaders: []string{
			"Content-Length",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}
}

// CORSMiddleware creates a gin middleware handler for CORS using the rs/cors library.
func CORSMiddleware(cfg CORSConfig) gin.HandlerFunc {
	c := cors.New(cors.Options{
		AllowedOrigins:   cfg.AllowOrigins,
		AllowedMethods:   cfg.AllowMethods,
		AllowedHeaders:   cfg.AllowHeaders,
		ExposedHeaders:   cfg.ExposeHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           cfg.MaxAge,
	})

	return func(ctx *gin.Context) {
		c.HandlerFunc(ctx.Writer, ctx.Request)

		// Handle preflight OPTIONS requests
		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx.Next()
	}
}
