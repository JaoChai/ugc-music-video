// Package handler provides HTTP handlers for the API.
package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/jaochai/ugc/internal/middleware"
	"github.com/jaochai/ugc/internal/models"
	"github.com/jaochai/ugc/internal/repository"
	"github.com/jaochai/ugc/pkg/response"
)

const maxSystemPromptLength = 15000

// AdminHandler handles admin-related HTTP requests
type AdminHandler struct {
	systemPromptRepo repository.SystemPromptRepository
	logger           *zap.Logger
}

// NewAdminHandler creates a new AdminHandler instance
func NewAdminHandler(
	systemPromptRepo repository.SystemPromptRepository,
	logger *zap.Logger,
) *AdminHandler {
	return &AdminHandler{
		systemPromptRepo: systemPromptRepo,
		logger:           logger,
	}
}

// RegisterRoutes registers all admin routes to the given router group
func (h *AdminHandler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware, adminMiddleware gin.HandlerFunc) {
	admin := rg.Group("/admin")
	admin.Use(authMiddleware)
	admin.Use(adminMiddleware)
	{
		admin.GET("/system-prompts", h.GetSystemPrompts)
		admin.PUT("/system-prompts", h.UpdateSystemPrompt)
	}
}

// GetSystemPrompts returns all system prompts
// @Summary Get all system prompts
// @Description Returns all system-wide default prompts (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=models.SystemPromptsResponse}
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/system-prompts [get]
func (h *AdminHandler) GetSystemPrompts(c *gin.Context) {
	prompts, err := h.systemPromptRepo.GetAll(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to get system prompts", zap.Error(err))
		response.Error(c, err)
		return
	}

	// Build response object
	resp := models.SystemPromptsResponse{}
	for _, p := range prompts {
		switch p.PromptType {
		case "song_concept":
			resp.SongConcept = p
		case "song_selector":
			resp.SongSelector = p
		case "image_concept":
			resp.ImageConcept = p
		}
	}

	response.Success(c, resp)
}

// UpdateSystemPrompt updates a specific system prompt
// @Summary Update a system prompt
// @Description Updates a system-wide default prompt (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Param input body models.UpdateSystemPromptInput true "Prompt data to update"
// @Security BearerAuth
// @Success 200 {object} response.Response{data=models.SystemPrompt}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/system-prompts [put]
func (h *AdminHandler) UpdateSystemPrompt(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	var input models.UpdateSystemPromptInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	// Validate prompt type
	validTypes := map[string]bool{
		"song_concept":  true,
		"song_selector": true,
		"image_concept": true,
	}
	if !validTypes[input.PromptType] {
		response.BadRequest(c, "invalid prompt type. Must be: song_concept, song_selector, or image_concept")
		return
	}

	// Validate prompt length
	if len(input.PromptContent) < 100 {
		response.BadRequest(c, "prompt must be at least 100 characters")
		return
	}
	if len(input.PromptContent) > maxSystemPromptLength {
		response.BadRequest(c, fmt.Sprintf("prompt must be %d characters or less", maxSystemPromptLength))
		return
	}

	// Update prompt
	if err := h.systemPromptRepo.Update(
		c.Request.Context(),
		input.PromptType,
		input.PromptContent,
		userID,
	); err != nil {
		h.logger.Error("failed to update system prompt",
			zap.Error(err),
			zap.String("prompt_type", input.PromptType),
		)
		response.Error(c, err)
		return
	}

	h.logger.Info("system prompt updated",
		zap.String("prompt_type", input.PromptType),
		zap.String("updated_by", userID.String()),
	)

	// Return updated prompt
	prompt, err := h.systemPromptRepo.GetByType(c.Request.Context(), input.PromptType)
	if err != nil {
		h.logger.Error("failed to get updated prompt", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, prompt)
}
