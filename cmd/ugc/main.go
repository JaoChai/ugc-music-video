package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/jaochai/ugc/internal/config"
	"github.com/jaochai/ugc/internal/database"
	"github.com/jaochai/ugc/internal/external/kie"
	"github.com/jaochai/ugc/internal/external/openrouter"
	"github.com/jaochai/ugc/internal/external/r2"
	"github.com/jaochai/ugc/internal/ffmpeg"
	"github.com/jaochai/ugc/internal/handler"
	"github.com/jaochai/ugc/internal/middleware"
	"github.com/jaochai/ugc/internal/repository"
	"github.com/jaochai/ugc/internal/service"
	"github.com/jaochai/ugc/internal/worker"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Setup zap logger
	logger, err := setupLogger(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to setup logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("starting UGC service",
		zap.String("env", cfg.Server.Env),
		zap.String("port", cfg.Server.Port),
	)

	// Create context for setup
	ctx := context.Background()

	// Connect to database
	db, err := database.New(ctx, cfg.Database.URL)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()
	logger.Info("connected to database")

	// Run migrations
	if err := database.RunMigrations(ctx, db); err != nil {
		logger.Fatal("failed to run migrations", zap.Error(err))
	}
	logger.Info("database migrations completed")

	// Create repositories
	userRepo := repository.NewUserRepository(db)
	jobRepo := repository.NewJobRepository(db)

	// Create external clients
	openRouterClient := openrouter.NewClient(cfg.OpenRouter.APIKey)
	sunoClient := kie.NewSunoClient(cfg.KIE.APIKey, cfg.KIE.BaseURL)
	nanoBananaClient := kie.NewNanoBananaClient(cfg.KIE.APIKey, cfg.KIE.BaseURL)

	// Create R2 client (optional - skip if not configured)
	var r2Client *r2.Client
	if cfg.R2.AccountID != "" {
		r2Client, err = r2.NewClient(ctx, r2.Config{
			AccountID:       cfg.R2.AccountID,
			AccessKeyID:     cfg.R2.AccessKeyID,
			SecretAccessKey: cfg.R2.SecretAccessKey,
			BucketName:      cfg.R2.BucketName,
			PublicURL:       cfg.R2.PublicURL,
		})
		if err != nil {
			logger.Warn("failed to create R2 client - video uploads will be disabled", zap.Error(err))
		} else {
			logger.Info("R2 client initialized")
		}
	} else {
		logger.Warn("R2 not configured - video uploads will be disabled")
	}

	// Create services
	authService := service.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.Expiry, logger)
	jobService := service.NewJobService(jobRepo, logger)

	// Create FFmpeg processor
	ffmpegProcessor := ffmpeg.NewProcessor(logger)

	// Create Asynq client
	redisOpt, err := asynq.ParseRedisURI(cfg.Redis.URL)
	if err != nil {
		logger.Fatal("failed to parse redis URL", zap.Error(err))
	}
	asynqClient := asynq.NewClient(redisOpt)
	defer asynqClient.Close()
	logger.Info("asynq client initialized")

	// Create worker dependencies
	workerDeps := worker.Dependencies{
		JobService:       jobService,
		UserRepo:         userRepo,
		OpenRouterClient: openRouterClient,
		SunoClient:       sunoClient,
		NanoBananaClient: nanoBananaClient,
		R2Client:         r2Client,
		FFmpegProcessor:  ffmpegProcessor,
		Logger:           logger,
	}

	// Create worker
	asynqWorker, err := worker.NewWorker(cfg.Redis.URL, workerDeps, logger)
	if err != nil {
		logger.Fatal("failed to create worker", zap.Error(err))
	}

	// Setup Gin router
	router := setupRouter(cfg, authService, jobService, userRepo, asynqClient, logger)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start worker in goroutine
	go func() {
		logger.Info("starting asynq worker")
		if err := asynqWorker.Start(); err != nil {
			logger.Error("worker error", zap.Error(err))
		}
	}()

	// Start HTTP server in goroutine
	go func() {
		logger.Info("starting HTTP server", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to start HTTP server", zap.Error(err))
		}
	}()

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown error", zap.Error(err))
	}
	logger.Info("HTTP server stopped")

	// Shutdown worker
	asynqWorker.Shutdown()
	logger.Info("worker stopped")

	// Close database connection
	db.Close()
	logger.Info("database connection closed")

	logger.Info("server shutdown complete")
}

// setupLogger creates a zap logger configured based on environment.
func setupLogger(cfg *config.Config) (*zap.Logger, error) {
	var zapConfig zap.Config

	if cfg.IsProduction() {
		// Production: JSON format, info level
		zapConfig = zap.NewProductionConfig()
		zapConfig.EncoderConfig.TimeKey = "timestamp"
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		// Development: console format, debug level
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	return zapConfig.Build()
}

// setupRouter creates and configures the Gin router with all routes and middleware.
func setupRouter(
	cfg *config.Config,
	authService service.AuthService,
	jobService service.JobService,
	userRepo repository.UserRepository,
	asynqClient *asynq.Client,
	logger *zap.Logger,
) *gin.Engine {
	// Set Gin mode based on environment
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(ginLogger(logger))

	// CORS middleware
	var corsConfig middleware.CORSConfig
	if cfg.IsProduction() {
		// Production: restrict to specific origins (configure via env)
		corsConfig = middleware.ProductionCORSConfig([]string{})
	} else {
		// Development: allow localhost origins
		corsConfig = middleware.DefaultCORSConfig()
	}
	router.Use(middleware.CORSMiddleware(corsConfig))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "ugc",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes
		authHandler := handler.NewAuthHandler(authService, logger)
		authHandler.RegisterRoutes(v1)

		// Job routes (protected)
		authMiddleware := middleware.AuthMiddleware(authService, logger)
		jobHandler := handler.NewJobHandler(jobService, userRepo, asynqClient, logger)
		jobHandler.RegisterRoutes(v1, authMiddleware)

		// Webhook routes
		webhooks := v1.Group("/webhooks")
		{
			// Suno webhook for music generation callbacks
			webhooks.POST("/suno", func(c *gin.Context) {
				// TODO: Implement Suno webhook handler
				c.JSON(http.StatusOK, gin.H{"status": "received"})
			})

			// KIE webhook for image generation callbacks
			webhooks.POST("/kie", func(c *gin.Context) {
				// TODO: Implement KIE webhook handler
				c.JSON(http.StatusOK, gin.H{"status": "received"})
			})
		}
	}

	return router
}

// ginLogger creates a gin middleware that logs requests using zap.
func ginLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		// Log request
		fields := []zap.Field{
			zap.Int("status", status),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Duration("latency", latency),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
		}

		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		switch {
		case status >= 500:
			logger.Error("server error", fields...)
		case status >= 400:
			logger.Warn("client error", fields...)
		default:
			logger.Info("request", fields...)
		}
	}
}
