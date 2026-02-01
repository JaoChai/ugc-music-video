// Package handler provides HTTP handlers for the UGC API.
package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/jaochai/ugc/internal/models"
	"github.com/jaochai/ugc/internal/repository"
	"github.com/jaochai/ugc/internal/service"
	"github.com/jaochai/ugc/internal/worker"
)

// SunoWebhookPayload represents the callback payload from KIE Suno API.
type SunoWebhookPayload struct {
	TaskID string `json:"taskId"`
	Status string `json:"status"` // completed, failed
	Data   struct {
		Songs []struct {
			ID       string  `json:"id"`
			AudioURL string  `json:"audioUrl"`
			Title    string  `json:"title"`
			Duration float64 `json:"duration"`
		} `json:"songs"`
	} `json:"data"`
	Error string `json:"error,omitempty"`
}

// NanoWebhookPayload represents the callback payload from KIE NanoBanana API.
type NanoWebhookPayload struct {
	TaskID string `json:"taskId"`
	Status string `json:"status"` // completed, failed
	Data   struct {
		Output struct {
			ImageURL string `json:"imageUrl"`
		} `json:"output"`
	} `json:"data"`
	Error string `json:"error,omitempty"`
}

// WebhookHandler handles webhook callbacks from external services.
type WebhookHandler struct {
	jobRepo     repository.JobRepository
	jobService  service.JobService
	asynqClient *asynq.Client
	logger      *zap.Logger
}

// NewWebhookHandler creates a new WebhookHandler instance.
func NewWebhookHandler(
	jobRepo repository.JobRepository,
	jobService service.JobService,
	asynqClient *asynq.Client,
	logger *zap.Logger,
) *WebhookHandler {
	return &WebhookHandler{
		jobRepo:     jobRepo,
		jobService:  jobService,
		asynqClient: asynqClient,
		logger:      logger,
	}
}

// RegisterRoutes registers webhook routes to the given router group.
func (h *WebhookHandler) RegisterRoutes(rg *gin.RouterGroup) {
	webhooks := rg.Group("/webhooks")
	{
		kie := webhooks.Group("/kie")
		{
			kie.POST("/suno", h.SunoCallback)
			kie.POST("/nano", h.NanoCallback)
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
		zap.String("task_id", payload.TaskID),
		zap.String("status", payload.Status),
	)

	// Find job by suno_task_id
	job, err := h.jobRepo.GetBySunoTaskID(c.Request.Context(), payload.TaskID)
	if err != nil {
		if errors.Is(err, repository.ErrJobNotFound) {
			// Log warning but return 200 for idempotency
			h.logger.Warn("job not found for suno task",
				zap.String("task_id", payload.TaskID),
			)
			c.JSON(http.StatusOK, gin.H{"message": "acknowledged"})
			return
		}
		h.logger.Error("failed to find job by suno task ID",
			zap.Error(err),
			zap.String("task_id", payload.TaskID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal error"})
		return
	}

	// Handle failed status
	if payload.Status == "failed" {
		errorMsg := payload.Error
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

	// Handle completed status
	if payload.Status == "completed" {
		// Convert payload songs to model format
		songs := make([]models.GeneratedSong, len(payload.Data.Songs))
		for i, s := range payload.Data.Songs {
			songs[i] = models.GeneratedSong{
				ID:       s.ID,
				AudioURL: s.AudioURL,
				Title:    s.Title,
				Duration: s.Duration,
			}
		}

		// Update job with generated songs
		if err := h.jobService.UpdateGeneratedSongs(c.Request.Context(), job.ID, payload.TaskID, songs); err != nil {
			h.logger.Error("failed to update job with generated songs",
				zap.Error(err),
				zap.String("job_id", job.ID.String()),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "internal error"})
			return
		}

		// Enqueue select song task
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
			zap.Int("song_count", len(songs)),
		)
	}

	c.JSON(http.StatusOK, gin.H{"message": "acknowledged"})
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
		zap.String("task_id", payload.TaskID),
		zap.String("status", payload.Status),
	)

	// Find job by nano_task_id
	job, err := h.jobRepo.GetByNanoTaskID(c.Request.Context(), payload.TaskID)
	if err != nil {
		if errors.Is(err, repository.ErrJobNotFound) {
			// Log warning but return 200 for idempotency
			h.logger.Warn("job not found for nano task",
				zap.String("task_id", payload.TaskID),
			)
			c.JSON(http.StatusOK, gin.H{"message": "acknowledged"})
			return
		}
		h.logger.Error("failed to find job by nano task ID",
			zap.Error(err),
			zap.String("task_id", payload.TaskID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal error"})
		return
	}

	// Handle failed status
	if payload.Status == "failed" {
		errorMsg := payload.Error
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
	if payload.Status == "completed" {
		// Update job with image URL
		if err := h.jobService.UpdateImageURL(c.Request.Context(), job.ID, payload.TaskID, payload.Data.Output.ImageURL); err != nil {
			h.logger.Error("failed to update job with image URL",
				zap.Error(err),
				zap.String("job_id", job.ID.String()),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "internal error"})
			return
		}

		// Enqueue process video task
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
			zap.String("image_url", payload.Data.Output.ImageURL),
		)
	}

	c.JSON(http.StatusOK, gin.H{"message": "acknowledged"})
}
