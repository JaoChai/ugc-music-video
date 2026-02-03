// Package handler provides HTTP handlers for the UGC API.
package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/jaochai/ugc/internal/models"
	"github.com/jaochai/ugc/internal/repository"
	"github.com/jaochai/ugc/internal/security"
	"github.com/jaochai/ugc/internal/service"
	"github.com/jaochai/ugc/internal/worker"
)

// SunoWebhookPayload represents the callback payload from KIE Suno API.
// https://docs.kie.ai/suno-api/quickstart#callback-format
type SunoWebhookPayload struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		CallbackType string `json:"callbackType"` // "text", "first", "complete"
		TaskID       string `json:"task_id"`      // Note: snake_case from KIE API
		Data         []struct {
			ID             string  `json:"id"`
			AudioURL       string  `json:"audio_url"` // Note: snake_case from KIE API
			StreamAudioURL string  `json:"stream_audio_url,omitempty"`
			ImageURL       string  `json:"image_url,omitempty"`
			Title          string  `json:"title"`
			Prompt         string  `json:"prompt,omitempty"`
			Tags           string  `json:"tags,omitempty"`
			Duration       float64 `json:"duration"`
			CreateTime     int64   `json:"createTime,omitempty"`
		} `json:"data"`
		ErrorMessage string `json:"errorMessage,omitempty"`
	} `json:"data"`
}

// NanoWebhookPayload represents the callback payload from KIE NanoBanana API.
// Uses the same format as TaskStatusResponse but delivered via webhook.
// https://docs.kie.ai/market/common/get-task-detail
type NanoWebhookPayload struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		TaskID     string `json:"taskId"`
		Model      string `json:"model,omitempty"`
		State      string `json:"state"` // "waiting", "queuing", "generating", "success", "fail"
		ResultJson string `json:"resultJson"`
		FailCode   string `json:"failCode,omitempty"`
		FailMsg    string `json:"failMsg,omitempty"`
	} `json:"data"`
}

// WebhookHandler handles webhook callbacks from external services.
type WebhookHandler struct {
	jobRepo      repository.JobRepository
	jobService   service.JobService
	asynqClient  *asynq.Client
	urlValidator *security.URLValidator
	logger       *zap.Logger
}

// NewWebhookHandler creates a new WebhookHandler instance.
func NewWebhookHandler(
	jobRepo repository.JobRepository,
	jobService service.JobService,
	asynqClient *asynq.Client,
	urlValidator *security.URLValidator,
	logger *zap.Logger,
) *WebhookHandler {
	// Use default validator if none provided
	if urlValidator == nil {
		urlValidator = security.NewURLValidator(nil)
	}
	return &WebhookHandler{
		jobRepo:      jobRepo,
		jobService:   jobService,
		asynqClient:  asynqClient,
		urlValidator: urlValidator,
		logger:       logger,
	}
}

// RegisterRoutes registers webhook routes to the given router group.
// rateLimitMiddleware is applied to all webhook routes.
// authMiddleware is applied to the authenticated webhook routes.
func (h *WebhookHandler) RegisterRoutes(rg *gin.RouterGroup, rateLimitMiddleware, authMiddleware gin.HandlerFunc) {
	webhooks := rg.Group("/webhooks")

	// Apply rate limiting to all webhook routes
	if rateLimitMiddleware != nil {
		webhooks.Use(rateLimitMiddleware)
	}

	{
		// Legacy KIE-style routes (deprecated, for backward compatibility only)
		// TODO: Remove these after migration period
		// WARNING: These routes have rate limiting but NO authentication
		kie := webhooks.Group("/kie")
		{
			kie.POST("/suno", h.SunoCallback)
			kie.POST("/nano", h.NanoCallback)
		}

		// Authenticated webhook routes (new, with token in path)
		// Format: /webhooks/:token/suno/:job_id
		authenticated := webhooks.Group("/:token")
		if authMiddleware != nil {
			authenticated.Use(authMiddleware)
		}
		{
			authenticated.POST("/suno/:job_id", h.SunoCallbackWithJobID)
			authenticated.POST("/nano/:job_id", h.NanoCallbackWithJobID)
		}
	}
}

// SunoCallback handles the callback from KIE Suno API when music generation is complete.
// @Summary Handle Suno webhook callback
// @Description Receives callback from KIE Suno API when music generation is complete or failed
// @Tags webhooks
// @Accept json
// @Produce json
// @Param payload body SunoWebhookPayload true "Suno webhook payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /webhooks/kie/suno [post]
func (h *WebhookHandler) SunoCallback(c *gin.Context) {
	var payload SunoWebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		h.logger.Error("failed to parse suno webhook payload",
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid payload"})
		return
	}

	h.logger.Info("received suno webhook callback",
		zap.String("task_id", payload.Data.TaskID),
		zap.String("callback_type", payload.Data.CallbackType),
		zap.Int("code", payload.Code),
	)

	// Validate task_id length to prevent memory/DB issues
	const maxTaskIDLength = 256
	if len(payload.Data.TaskID) == 0 || len(payload.Data.TaskID) > maxTaskIDLength {
		h.logger.Warn("invalid task_id length",
			zap.Int("length", len(payload.Data.TaskID)),
		)
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid task_id"})
		return
	}

	// Find job by suno_task_id
	job, err := h.jobRepo.GetBySunoTaskID(c.Request.Context(), payload.Data.TaskID)
	if err != nil {
		if errors.Is(err, repository.ErrJobNotFound) {
			// Log warning but return 200 for idempotency
			h.logger.Warn("job not found for suno task",
				zap.String("task_id", payload.Data.TaskID),
			)
			c.JSON(http.StatusOK, gin.H{"message": "acknowledged"})
			return
		}
		h.logger.Error("failed to find job by suno task ID",
			zap.Error(err),
			zap.String("task_id", payload.Data.TaskID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal error"})
		return
	}

	// Idempotency check: only process if job is in expected status
	if job.Status != models.StatusGeneratingMusic {
		h.logger.Warn("suno callback received for job not in expected status",
			zap.String("job_id", job.ID.String()),
			zap.String("current_status", job.Status),
			zap.String("expected_status", models.StatusGeneratingMusic),
		)
		c.JSON(http.StatusOK, gin.H{"message": "acknowledged"})
		return
	}

	// Handle failed status (code != 200 or callbackType indicates failure)
	if payload.Code != 200 {
		errorMsg := payload.Data.ErrorMessage
		if errorMsg == "" {
			errorMsg = payload.Msg
		}
		if errorMsg == "" {
			errorMsg = "music generation failed"
		}
		if err := h.jobService.MarkFailed(c.Request.Context(), job.ID, errorMsg); err != nil {
			h.logger.Error("failed to mark job as failed",
				zap.Error(err),
				zap.String("job_id", job.ID.String()),
			)
		}
		c.JSON(http.StatusOK, gin.H{"message": "acknowledged"})
		return
	}

	// Handle completed status (callbackType == "complete" or "first")
	// "first" means first track is ready, "complete" means all tracks are ready
	if payload.Data.CallbackType == "complete" || payload.Data.CallbackType == "first" {
		// Validate songs array is not empty
		if len(payload.Data.Data) == 0 {
			h.logger.Error("suno callback has empty songs array",
				zap.String("job_id", job.ID.String()),
				zap.String("task_id", payload.Data.TaskID),
			)
			_ = h.jobService.MarkFailed(c.Request.Context(), job.ID, "music generation returned no songs")
			c.JSON(http.StatusOK, gin.H{"message": "acknowledged"})
			return
		}

		// Filter songs with valid AudioURL and validate URLs
		songs := make([]models.GeneratedSong, 0, len(payload.Data.Data))
		for _, s := range payload.Data.Data {
			// Skip songs with empty AudioURL
			if s.AudioURL == "" {
				h.logger.Warn("skipping song with empty audio_url",
					zap.String("job_id", job.ID.String()),
					zap.String("song_id", s.ID),
				)
				continue
			}

			// Validate AudioURL to prevent SSRF
			if err := h.urlValidator.ValidateURL(s.AudioURL); err != nil {
				h.logger.Warn("skipping song with invalid audio_url",
					zap.String("job_id", job.ID.String()),
					zap.String("song_id", s.ID),
					zap.Error(err),
				)
				continue
			}

			songs = append(songs, models.GeneratedSong{
				ID:       s.ID,
				AudioURL: s.AudioURL,
				Title:    s.Title,
				Duration: s.Duration,
			})
		}

		// Check if any valid songs remain
		if len(songs) == 0 {
			h.logger.Error("all songs have invalid audio URLs",
				zap.String("job_id", job.ID.String()),
				zap.Int("total_songs", len(payload.Data.Data)),
			)
			_ = h.jobService.MarkFailed(c.Request.Context(), job.ID, "all songs have invalid audio URLs")
			c.JSON(http.StatusOK, gin.H{"message": "acknowledged"})
			return
		}

		// Update job with generated songs
		if err := h.jobService.UpdateGeneratedSongs(c.Request.Context(), job.ID, payload.Data.TaskID, songs); err != nil {
			h.logger.Error("failed to update job with generated songs",
				zap.Error(err),
				zap.String("job_id", job.ID.String()),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "internal error"})
			return
		}

		// Enqueue select song task with deduplication
		task, err := worker.NewSelectSongTask(job.ID)
		if err != nil {
			h.logger.Error("failed to create select song task",
				zap.Error(err),
				zap.String("job_id", job.ID.String()),
			)
			_ = h.jobService.MarkFailed(c.Request.Context(), job.ID, "failed to enqueue select song task")
			c.JSON(http.StatusInternalServerError, gin.H{"message": "internal error"})
			return
		}

		if _, err := h.asynqClient.Enqueue(task); err != nil {
			// Check if it's a duplicate task error (already enqueued)
			if errors.Is(err, asynq.ErrTaskIDConflict) {
				h.logger.Warn("select song task already enqueued (duplicate callback)",
					zap.String("job_id", job.ID.String()),
				)
				c.JSON(http.StatusOK, gin.H{"message": "acknowledged"})
				return
			}
			h.logger.Error("failed to enqueue select song task",
				zap.Error(err),
				zap.String("job_id", job.ID.String()),
			)
			_ = h.jobService.MarkFailed(c.Request.Context(), job.ID, "failed to enqueue select song task")
			c.JSON(http.StatusInternalServerError, gin.H{"message": "internal error"})
			return
		}

		h.logger.Info("suno callback processed, select song task enqueued",
			zap.String("job_id", job.ID.String()),
			zap.Int("valid_song_count", len(songs)),
			zap.Int("total_song_count", len(payload.Data.Data)),
		)
	}

	// For "text" callbackType, just acknowledge - lyrics generated but audio not ready
	c.JSON(http.StatusOK, gin.H{"message": "acknowledged"})
}

// SunoCallbackWithJobID handles the callback with job_id in the URL path.
// This is used when the callback URL format is /webhooks/:token/suno/:job_id
func (h *WebhookHandler) SunoCallbackWithJobID(c *gin.Context) {
	jobID := c.Param("job_id")
	h.logger.Debug("suno callback with job_id in path", zap.String("job_id", jobID))

	// Delegate to main handler - the job_id in path is for reference,
	// but we use task_id from payload for lookup (more reliable)
	h.SunoCallback(c)
}

// NanoCallback handles the callback from KIE NanoBanana API when image generation is complete.
// @Summary Handle Nano webhook callback
// @Description Receives callback from KIE NanoBanana API when image generation is complete or failed
// @Tags webhooks
// @Accept json
// @Produce json
// @Param payload body NanoWebhookPayload true "Nano webhook payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /webhooks/kie/nano [post]
func (h *WebhookHandler) NanoCallback(c *gin.Context) {
	var payload NanoWebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		h.logger.Error("failed to parse nano webhook payload",
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid payload"})
		return
	}

	h.logger.Info("received nano webhook callback",
		zap.String("task_id", payload.Data.TaskID),
		zap.String("state", payload.Data.State),
		zap.Int("code", payload.Code),
	)

	// Validate task_id length to prevent memory/DB issues
	const maxTaskIDLength = 256
	if len(payload.Data.TaskID) == 0 || len(payload.Data.TaskID) > maxTaskIDLength {
		h.logger.Warn("invalid task_id length",
			zap.Int("length", len(payload.Data.TaskID)),
		)
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid task_id"})
		return
	}

	// Find job by nano_task_id
	job, err := h.jobRepo.GetByNanoTaskID(c.Request.Context(), payload.Data.TaskID)
	if err != nil {
		if errors.Is(err, repository.ErrJobNotFound) {
			// Log warning but return 200 for idempotency
			h.logger.Warn("job not found for nano task",
				zap.String("task_id", payload.Data.TaskID),
			)
			c.JSON(http.StatusOK, gin.H{"message": "acknowledged"})
			return
		}
		h.logger.Error("failed to find job by nano task ID",
			zap.Error(err),
			zap.String("task_id", payload.Data.TaskID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal error"})
		return
	}

	// Idempotency check: only process if job is in expected status
	if job.Status != models.StatusGeneratingImage {
		h.logger.Warn("nano callback received for job not in expected status",
			zap.String("job_id", job.ID.String()),
			zap.String("current_status", job.Status),
			zap.String("expected_status", models.StatusGeneratingImage),
		)
		c.JSON(http.StatusOK, gin.H{"message": "acknowledged"})
		return
	}

	// Handle failed status
	if payload.Code != 200 || payload.Data.State == "fail" {
		errorMsg := payload.Data.FailMsg
		if errorMsg == "" {
			errorMsg = payload.Message
		}
		if errorMsg == "" {
			errorMsg = "image generation failed"
		}
		if err := h.jobService.MarkFailed(c.Request.Context(), job.ID, errorMsg); err != nil {
			h.logger.Error("failed to mark job as failed",
				zap.Error(err),
				zap.String("job_id", job.ID.String()),
			)
		}
		c.JSON(http.StatusOK, gin.H{"message": "acknowledged"})
		return
	}

	// Handle completed status
	if payload.Data.State == "success" {
		// Extract image URL from resultJson
		imageURL, err := extractImageURL(payload.Data.ResultJson)
		if err != nil {
			h.logger.Error("failed to extract image URL from callback",
				zap.Error(err),
				zap.String("job_id", job.ID.String()),
				zap.Int("result_json_length", len(payload.Data.ResultJson)), // Sanitized log
			)
			_ = h.jobService.MarkFailed(c.Request.Context(), job.ID, "failed to extract image URL from callback")
			c.JSON(http.StatusInternalServerError, gin.H{"message": "internal error"})
			return
		}

		// Validate image URL to prevent SSRF
		if err := h.urlValidator.ValidateURL(imageURL); err != nil {
			h.logger.Error("image URL validation failed",
				zap.Error(err),
				zap.String("job_id", job.ID.String()),
			)
			_ = h.jobService.MarkFailed(c.Request.Context(), job.ID, "image URL validation failed")
			c.JSON(http.StatusOK, gin.H{"message": "acknowledged"})
			return
		}

		// Update job with image URL
		if err := h.jobService.UpdateImageURL(c.Request.Context(), job.ID, payload.Data.TaskID, imageURL); err != nil {
			h.logger.Error("failed to update job with image URL",
				zap.Error(err),
				zap.String("job_id", job.ID.String()),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "internal error"})
			return
		}

		// Enqueue process video task with deduplication
		task, err := worker.NewProcessVideoTask(job.ID)
		if err != nil {
			h.logger.Error("failed to create process video task",
				zap.Error(err),
				zap.String("job_id", job.ID.String()),
			)
			_ = h.jobService.MarkFailed(c.Request.Context(), job.ID, "failed to enqueue process video task")
			c.JSON(http.StatusInternalServerError, gin.H{"message": "internal error"})
			return
		}

		if _, err := h.asynqClient.Enqueue(task); err != nil {
			// Check if it's a duplicate task error (already enqueued)
			if errors.Is(err, asynq.ErrTaskIDConflict) {
				h.logger.Warn("process video task already enqueued (duplicate callback)",
					zap.String("job_id", job.ID.String()),
				)
				c.JSON(http.StatusOK, gin.H{"message": "acknowledged"})
				return
			}
			h.logger.Error("failed to enqueue process video task",
				zap.Error(err),
				zap.String("job_id", job.ID.String()),
			)
			_ = h.jobService.MarkFailed(c.Request.Context(), job.ID, "failed to enqueue process video task")
			c.JSON(http.StatusInternalServerError, gin.H{"message": "internal error"})
			return
		}

		h.logger.Info("nano callback processed, process video task enqueued",
			zap.String("job_id", job.ID.String()),
			zap.Bool("has_image_url", true), // Sanitized log - don't log the actual URL
		)
	}

	c.JSON(http.StatusOK, gin.H{"message": "acknowledged"})
}

// NanoCallbackWithJobID handles the callback with job_id in the URL path.
// This is used when the callback URL format is /webhooks/:token/nano/:job_id
func (h *WebhookHandler) NanoCallbackWithJobID(c *gin.Context) {
	jobID := c.Param("job_id")
	h.logger.Debug("nano callback with job_id in path", zap.String("job_id", jobID))

	// Delegate to main handler - the job_id in path is for reference,
	// but we use task_id from payload for lookup (more reliable)
	h.NanoCallback(c)
}

// extractImageURL parses the resultJson and extracts the first image URL.
// The resultJson format is: {"resultUrls":["https://..."]}
func extractImageURL(resultJson string) (string, error) {
	if resultJson == "" {
		return "", fmt.Errorf("empty resultJson")
	}

	var result struct {
		ResultUrls []string `json:"resultUrls"`
	}
	if err := json.Unmarshal([]byte(resultJson), &result); err != nil {
		return "", fmt.Errorf("failed to parse resultJson: %w", err)
	}

	if len(result.ResultUrls) == 0 {
		return "", fmt.Errorf("no image URLs in resultJson")
	}

	return result.ResultUrls[0], nil
}
