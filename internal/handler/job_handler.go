// Package handler provides HTTP handlers for the UGC API.
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/jaochai/ugc/internal/middleware"
	"github.com/jaochai/ugc/internal/models"
	"github.com/jaochai/ugc/internal/repository"
	"github.com/jaochai/ugc/internal/service"
	"github.com/jaochai/ugc/internal/worker"
	"github.com/jaochai/ugc/pkg/response"
)

// JobHandler handles job-related HTTP requests.
type JobHandler struct {
	jobService    service.JobService
	userRepo      repository.UserRepository
	cryptoService service.CryptoService
	asynqClient   *asynq.Client
	logger        *zap.Logger
}

// NewJobHandler creates a new JobHandler instance.
func NewJobHandler(
	jobService service.JobService,
	userRepo repository.UserRepository,
	cryptoService service.CryptoService,
	asynqClient *asynq.Client,
	logger *zap.Logger,
) *JobHandler {
	return &JobHandler{
		jobService:    jobService,
		userRepo:      userRepo,
		cryptoService: cryptoService,
		asynqClient:   asynqClient,
		logger:        logger,
	}
}

// RegisterRoutes registers job-related routes to the given router group.
func (h *JobHandler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	jobs := rg.Group("/jobs")
	jobs.Use(authMiddleware)
	{
		jobs.POST("", h.Create)
		jobs.GET("", h.List)
		jobs.GET("/:id", h.GetByID)
		jobs.DELETE("/:id", h.Cancel)
	}
}

// Create handles job creation requests.
// @Summary Create a new job
// @Description Creates a new UGC generation job with the given concept
// @Tags jobs
// @Accept json
// @Produce json
// @Param input body models.CreateJobInput true "Job creation input"
// @Success 201 {object} response.Response{data=models.JobResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /jobs [post]
func (h *JobHandler) Create(c *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	// Bind JSON input
	var input models.CreateJobInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	// Validate input
	if input.Concept == "" {
		response.ValidationError(c, map[string]string{
			"concept": "concept is required",
		})
		return
	}
	if len(input.Concept) < 5 {
		response.ValidationError(c, map[string]string{
			"concept": "concept must be at least 5 characters",
		})
		return
	}

	// Get user to retrieve default model and check API keys
	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get user for job creation",
			zap.Error(err),
			zap.String("user_id", userID.String()),
		)
		response.Error(c, err)
		return
	}

	// Validate user has required API keys (user already has encrypted keys from GetByID)
	hasOpenRouterKey := false
	if user.OpenRouterAPIKey != nil && *user.OpenRouterAPIKey != "" {
		decrypted, err := h.cryptoService.Decrypt(*user.OpenRouterAPIKey)
		if err != nil {
			h.logger.Warn("failed to decrypt OpenRouter API key", zap.Error(err))
		} else if decrypted != "" {
			hasOpenRouterKey = true
		}
	}
	if !hasOpenRouterKey {
		response.BadRequest(c, "OpenRouter API key is required. Please configure in Settings.")
		return
	}

	hasKIEKey := false
	if user.KIEAPIKey != nil && *user.KIEAPIKey != "" {
		decrypted, err := h.cryptoService.Decrypt(*user.KIEAPIKey)
		if err != nil {
			h.logger.Warn("failed to decrypt KIE API key", zap.Error(err))
		} else if decrypted != "" {
			hasKIEKey = true
		}
	}
	if !hasKIEKey {
		response.BadRequest(c, "KIE API key is required. Please configure in Settings.")
		return
	}

	// Create job
	job, err := h.jobService.Create(c.Request.Context(), userID, input, user.OpenRouterModel)
	if err != nil {
		h.logger.Error("failed to create job",
			zap.Error(err),
			zap.String("user_id", userID.String()),
		)
		response.Error(c, err)
		return
	}

	// Enqueue analyze concept task
	task, err := worker.NewAnalyzeConceptTask(job.ID)
	if err != nil {
		h.logger.Error("failed to create analyze concept task",
			zap.Error(err),
			zap.String("job_id", job.ID.String()),
		)
		// Job is created but task enqueue failed - mark job as failed
		_ = h.jobService.MarkFailed(c.Request.Context(), job.ID, "failed to enqueue analyze task")
		response.Error(c, err)
		return
	}

	if _, err := h.asynqClient.Enqueue(task); err != nil {
		h.logger.Error("failed to enqueue analyze concept task",
			zap.Error(err),
			zap.String("job_id", job.ID.String()),
		)
		// Job is created but task enqueue failed - mark job as failed
		_ = h.jobService.MarkFailed(c.Request.Context(), job.ID, "failed to enqueue analyze task")
		response.Error(c, err)
		return
	}

	h.logger.Info("job created and task enqueued",
		zap.String("job_id", job.ID.String()),
		zap.String("user_id", userID.String()),
	)

	response.Created(c, job.ToResponse())
}

// List handles listing jobs for the authenticated user.
// @Summary List jobs
// @Description Lists all jobs for the authenticated user with pagination
// @Tags jobs
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(10) maximum(100)
// @Success 200 {object} response.Response{data=[]models.JobResponse,meta=response.Meta}
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /jobs [get]
func (h *JobHandler) List(c *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	// Parse pagination params
	page := 1
	perPage := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if perPageStr := c.Query("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 {
			perPage = pp
			if perPage > 100 {
				perPage = 100
			}
		}
	}

	// Get jobs
	jobs, meta, err := h.jobService.List(c.Request.Context(), userID, page, perPage)
	if err != nil {
		h.logger.Error("failed to list jobs",
			zap.Error(err),
			zap.String("user_id", userID.String()),
		)
		response.Error(c, err)
		return
	}

	// Convert to responses
	jobResponses := make([]*models.JobResponse, len(jobs))
	for i, job := range jobs {
		jobResponses[i] = job.ToResponse()
	}

	response.SuccessWithMeta(c, jobResponses, meta)
}

// GetByID handles getting a job by ID.
// @Summary Get job by ID
// @Description Gets a job by its ID for the authenticated user
// @Tags jobs
// @Produce json
// @Param id path string true "Job ID" format(uuid)
// @Success 200 {object} response.Response{data=models.JobResponse}
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /jobs/{id} [get]
func (h *JobHandler) GetByID(c *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	// Parse job ID from URL
	jobIDStr := c.Param("id")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		response.BadRequest(c, "invalid job ID format")
		return
	}

	// Get job
	job, err := h.jobService.GetByID(c.Request.Context(), userID, jobID)
	if err != nil {
		h.logger.Debug("failed to get job",
			zap.Error(err),
			zap.String("job_id", jobIDStr),
			zap.String("user_id", userID.String()),
		)
		response.Error(c, err)
		return
	}

	response.Success(c, job.ToResponse())
}

// Cancel handles job cancellation requests.
// @Summary Cancel a job
// @Description Cancels a job if it's not in a terminal state
// @Tags jobs
// @Produce json
// @Param id path string true "Job ID" format(uuid)
// @Success 204 "No Content"
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /jobs/{id} [delete]
func (h *JobHandler) Cancel(c *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	// Parse job ID from URL
	jobIDStr := c.Param("id")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		response.BadRequest(c, "invalid job ID format")
		return
	}

	// Cancel job
	if err := h.jobService.Cancel(c.Request.Context(), userID, jobID); err != nil {
		h.logger.Debug("failed to cancel job",
			zap.Error(err),
			zap.String("job_id", jobIDStr),
			zap.String("user_id", userID.String()),
		)
		response.Error(c, err)
		return
	}

	h.logger.Info("job cancelled",
		zap.String("job_id", jobIDStr),
		zap.String("user_id", userID.String()),
	)

	response.NoContent(c)
}
