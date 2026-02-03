// Package tasks provides task handlers for the async worker.
package tasks

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/jaochai/ugc/internal/agents"
	"github.com/jaochai/ugc/internal/external/kie"
	"github.com/jaochai/ugc/internal/external/openrouter"
	"github.com/jaochai/ugc/internal/external/r2"
	"github.com/jaochai/ugc/internal/ffmpeg"
	"github.com/jaochai/ugc/internal/models"
	"github.com/jaochai/ugc/internal/repository"
)

// CryptoService interface for decrypting API keys.
type CryptoService interface {
	Decrypt(ciphertext string) (string, error)
}

// Dependencies holds all external dependencies required by task handlers.
type Dependencies struct {
	JobRepo          repository.JobRepository
	UserRepo         repository.UserRepository
	SystemPromptRepo repository.SystemPromptRepository
	CryptoService    CryptoService
	R2Client         *r2.Client
	FFmpegProcessor  *ffmpeg.Processor
	AsynqClient      *asynq.Client
	Logger           *zap.Logger
	WebhookBaseURL   string // Base URL for webhooks, empty to disable
	WebhookSecret    string // Secret token for webhook authentication
	KIEBaseURL       string // Base URL for KIE API
}

// DefaultLLMModel is the default model to use if user hasn't configured one.
const DefaultLLMModel = "anthropic/claude-3.5-sonnet"

// getEffectivePrompt returns the system default prompt from DB.
func getEffectivePrompt(ctx context.Context, deps *Dependencies, promptType string) *string {
	systemPrompt, err := deps.SystemPromptRepo.GetByType(ctx, promptType)
	if err != nil {
		deps.Logger.Warn("failed to get system prompt from DB, using hardcoded default",
			zap.String("prompt_type", promptType),
			zap.Error(err),
		)
		return nil // Will fallback to hardcoded default in agent
	}

	return &systemPrompt.PromptContent
}

// getUserAPIKeys retrieves and decrypts the user's API keys.
func getUserAPIKeys(ctx context.Context, deps *Dependencies, userID uuid.UUID) (openRouterKey, kieKey string, err error) {
	encOpenRouterKey, encKIEKey, err := deps.UserRepo.GetAPIKeys(ctx, userID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get API keys: %w", err)
	}

	if encOpenRouterKey != nil && *encOpenRouterKey != "" {
		openRouterKey, err = deps.CryptoService.Decrypt(*encOpenRouterKey)
		if err != nil {
			return "", "", fmt.Errorf("failed to decrypt OpenRouter API key: %w", err)
		}
	}

	if encKIEKey != nil && *encKIEKey != "" {
		kieKey, err = deps.CryptoService.Decrypt(*encKIEKey)
		if err != nil {
			return "", "", fmt.Errorf("failed to decrypt KIE API key: %w", err)
		}
	}

	return openRouterKey, kieKey, nil
}

// HandleAnalyzeConcept creates a handler for the analyze concept task.
// This handler:
// 1. Loads the job from database
// 2. Loads the user to get their LLM model preference
// 3. Creates a SongConceptAgent
// 4. Analyzes the concept
// 5. Updates the job with the song_prompt
// 6. Enqueues TypeGenerateMusic
func HandleAnalyzeConcept(deps *Dependencies) asynq.HandlerFunc {
	return func(ctx context.Context, task *asynq.Task) error {
		logger := deps.Logger.With(zap.String("task_type", TypeAnalyzeConcept))

		// Parse payload
		payload, err := UnmarshalTaskPayload(task.Payload())
		if err != nil {
			logger.Error("failed to unmarshal task payload", zap.Error(err))
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		logger = logger.With(zap.String("job_id", payload.JobID.String()))
		logger.Info("starting analyze concept task")

		// Load job from database
		job, err := deps.JobRepo.GetByID(ctx, payload.JobID)
		if err != nil {
			logger.Error("failed to load job", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to load job: %v", err))
		}

		// Update job status to analyzing
		job.Status = models.StatusAnalyzing
		if err := deps.JobRepo.Update(ctx, job); err != nil {
			logger.Error("failed to update job status", zap.Error(err))
			return fmt.Errorf("failed to update job status: %w", err)
		}

		// Load user to get LLM model preference
		user, err := deps.UserRepo.GetByID(ctx, job.UserID)
		if err != nil {
			logger.Error("failed to load user", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to load user: %v", err))
		}

		// Get user's API keys
		openRouterKey, _, err := getUserAPIKeys(ctx, deps, job.UserID)
		if err != nil {
			logger.Error("failed to get user API keys", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to get API keys: %v", err))
		}
		if openRouterKey == "" {
			logger.Error("user has no OpenRouter API key")
			return markJobFailed(ctx, deps, payload.JobID, "user has no OpenRouter API key configured")
		}

		// Determine which LLM model to use
		llmModel := user.OpenRouterModel
		if llmModel == "" {
			llmModel = DefaultLLMModel
		}

		// Get effective prompt from system defaults
		effectivePrompt := getEffectivePrompt(ctx, deps, "song_concept")

		// Create per-user OpenRouter client and SongConceptAgent
		openRouterClient := openrouter.NewClient(openRouterKey)
		agent := agents.NewSongConceptAgentWithPrompt(openRouterClient, llmModel, logger, effectivePrompt)

		// Analyze concept
		input := agents.SongConceptInput{
			Concept:  job.Concept,
			Language: "Thai", // Default to Thai
		}

		output, err := agent.Analyze(ctx, input)
		if err != nil {
			logger.Error("failed to analyze concept", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to analyze concept: %v", err))
		}

		// Update job with song_prompt
		job.SongPrompt = output.ToSongPrompt()
		// Force model to V5 regardless of LLM output
		job.SongPrompt.Model = "V5"
		job.LLMModel = llmModel
		if err := deps.JobRepo.Update(ctx, job); err != nil {
			logger.Error("failed to update job with song prompt", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to update job: %v", err))
		}

		logger.Info("concept analysis complete",
			zap.String("title", output.Title),
			zap.String("style", output.Style),
		)

		// Enqueue next task: generate music
		nextPayload, _ := (&TaskPayload{JobID: payload.JobID}).Marshal()
		nextTask := asynq.NewTask(TypeGenerateMusic, nextPayload)
		if _, err := deps.AsynqClient.Enqueue(nextTask); err != nil {
			logger.Error("failed to enqueue generate music task", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to enqueue next task: %v", err))
		}

		logger.Info("enqueued generate music task")
		return nil
	}
}

// HandleGenerateMusic creates a handler for the generate music task.
// This handler:
// 1. Loads the job
// 2. Calls SunoClient.Generate() with song_prompt
// 3. Updates the job with suno_task_id and status = generating_music
// 4. If webhook is configured, returns nil (webhook will trigger next task)
// 5. Otherwise polls for completion and updates job with generated songs
func HandleGenerateMusic(deps *Dependencies) asynq.HandlerFunc {
	return func(ctx context.Context, task *asynq.Task) error {
		logger := deps.Logger.With(zap.String("task_type", TypeGenerateMusic))

		// Parse payload
		payload, err := UnmarshalTaskPayload(task.Payload())
		if err != nil {
			logger.Error("failed to unmarshal task payload", zap.Error(err))
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		logger = logger.With(zap.String("job_id", payload.JobID.String()))
		logger.Info("starting generate music task")

		// Load job
		job, err := deps.JobRepo.GetByID(ctx, payload.JobID)
		if err != nil {
			logger.Error("failed to load job", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to load job: %v", err))
		}

		// Verify song_prompt exists
		if job.SongPrompt == nil {
			logger.Error("job missing song_prompt")
			return markJobFailed(ctx, deps, payload.JobID, "job missing song_prompt")
		}

		// Get user's KIE API key
		_, kieKey, err := getUserAPIKeys(ctx, deps, job.UserID)
		if err != nil {
			logger.Error("failed to get user API keys", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to get API keys: %v", err))
		}
		if kieKey == "" {
			logger.Error("user has no KIE API key")
			return markJobFailed(ctx, deps, payload.JobID, "user has no KIE API key configured")
		}

		// Create per-user Suno client
		sunoClient := kie.NewSunoClient(kieKey, deps.KIEBaseURL)

		// Build Suno generate request
		req := kie.GenerateRequest{
			Prompt:       job.SongPrompt.Prompt,
			CustomMode:   true,
			Instrumental: job.SongPrompt.Instrumental,
			Model:        job.SongPrompt.Model,
			Style:        job.SongPrompt.Style,
			Title:        job.SongPrompt.Title,
		}

		// Add webhook URL if configured
		// Route: /api/v1/webhooks/:token/suno/:job_id (matches RegisterRoutes in webhook_handler.go)
		if deps.WebhookBaseURL != "" && deps.WebhookSecret != "" {
			req.CallBackUrl = fmt.Sprintf("%s/api/v1/webhooks/%s/suno/%s", deps.WebhookBaseURL, deps.WebhookSecret, payload.JobID.String())
		}

		// Call Suno API to start generation
		taskID, err := sunoClient.Generate(ctx, req)
		if err != nil {
			logger.Error("failed to generate music", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to generate music: %v", err))
		}

		logger.Info("music generation started", zap.String("suno_task_id", taskID))

		// Update job with suno_task_id and status
		job.SunoTaskID = &taskID
		job.Status = models.StatusGeneratingMusic
		if err := deps.JobRepo.Update(ctx, job); err != nil {
			logger.Error("failed to update job with suno task id", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to update job: %v", err))
		}

		// If webhook is configured, return and let webhook handle completion
		if deps.WebhookBaseURL != "" {
			logger.Info("webhook configured, waiting for callback")
			return nil
		}

		// Otherwise, poll for completion
		logger.Info("polling for music generation completion")
		taskResp, err := sunoClient.WaitForCompletion(ctx, taskID, 10*time.Minute)
		if err != nil {
			logger.Error("music generation failed or timed out", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("music generation failed: %v", err))
		}

		// Convert songs to models.GeneratedSong (using new response structure)
		generatedSongs := make([]models.GeneratedSong, len(taskResp.Data.Response.SunoData))
		for i, song := range taskResp.Data.Response.SunoData {
			generatedSongs[i] = models.GeneratedSong{
				ID:       song.Id,
				AudioURL: song.AudioUrl,
				Title:    song.Title,
				Duration: song.Duration,
			}
		}

		// Update job with generated songs
		job.GeneratedSongs = generatedSongs
		if err := deps.JobRepo.Update(ctx, job); err != nil {
			logger.Error("failed to update job with generated songs", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to update job: %v", err))
		}

		logger.Info("music generation complete", zap.Int("song_count", len(generatedSongs)))

		// Enqueue next task: select song
		nextPayload, _ := (&TaskPayload{JobID: payload.JobID}).Marshal()
		nextTask := asynq.NewTask(TypeSelectSong, nextPayload)
		if _, err := deps.AsynqClient.Enqueue(nextTask); err != nil {
			logger.Error("failed to enqueue select song task", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to enqueue next task: %v", err))
		}

		logger.Info("enqueued select song task")
		return nil
	}
}

// HandleSelectSong creates a handler for the select song task.
// This handler:
// 1. Loads the job (must have generated_songs)
// 2. Creates a SongSelectorAgent
// 3. Selects the best song
// 4. Updates the job with selected_song_id and audio_url
// 5. Enqueues TypeGenerateImage
func HandleSelectSong(deps *Dependencies) asynq.HandlerFunc {
	return func(ctx context.Context, task *asynq.Task) error {
		logger := deps.Logger.With(zap.String("task_type", TypeSelectSong))

		// Parse payload
		payload, err := UnmarshalTaskPayload(task.Payload())
		if err != nil {
			logger.Error("failed to unmarshal task payload", zap.Error(err))
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		logger = logger.With(zap.String("job_id", payload.JobID.String()))
		logger.Info("starting select song task")

		// Load job
		job, err := deps.JobRepo.GetByID(ctx, payload.JobID)
		if err != nil {
			logger.Error("failed to load job", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to load job: %v", err))
		}

		// Verify generated_songs exists
		if len(job.GeneratedSongs) == 0 {
			logger.Error("job has no generated songs")
			return markJobFailed(ctx, deps, payload.JobID, "job has no generated songs")
		}

		// Update status
		job.Status = models.StatusSelectingSong
		if err := deps.JobRepo.Update(ctx, job); err != nil {
			logger.Error("failed to update job status", zap.Error(err))
		}

		// Get user's OpenRouter API key
		openRouterKey, _, err := getUserAPIKeys(ctx, deps, job.UserID)
		if err != nil {
			logger.Error("failed to get user API keys", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to get API keys: %v", err))
		}
		if openRouterKey == "" {
			logger.Error("user has no OpenRouter API key")
			return markJobFailed(ctx, deps, payload.JobID, "user has no OpenRouter API key configured")
		}

		// Determine LLM model
		llmModel := job.LLMModel
		if llmModel == "" {
			llmModel = DefaultLLMModel
		}

		// Get effective prompt from system defaults
		effectivePrompt := getEffectivePrompt(ctx, deps, "song_selector")

		// Create per-user OpenRouter client and SongSelectorAgent
		openRouterClient := openrouter.NewClient(openRouterKey)
		agent := agents.NewSongSelectorAgentWithPrompt(openRouterClient, llmModel, logger, effectivePrompt)

		// Build song candidates
		candidates := make([]agents.SongCandidate, len(job.GeneratedSongs))
		for i, song := range job.GeneratedSongs {
			candidates[i] = agents.SongCandidate{
				ID:       song.ID,
				Title:    song.Title,
				Duration: song.Duration,
				AudioURL: song.AudioURL,
			}
		}

		// Select best song
		input := agents.SongSelectorInput{
			OriginalConcept: job.Concept,
			Songs:           candidates,
		}

		output, err := agent.Select(ctx, input)
		if err != nil {
			logger.Error("failed to select song", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to select song: %v", err))
		}

		// Find selected song's audio URL
		var selectedAudioURL string
		for _, song := range job.GeneratedSongs {
			if song.ID == output.SelectedSongID {
				selectedAudioURL = song.AudioURL
				break
			}
		}

		if selectedAudioURL == "" {
			logger.Error("selected song not found in generated songs",
				zap.String("selected_id", output.SelectedSongID))
			return markJobFailed(ctx, deps, payload.JobID, "selected song not found")
		}

		// Update job with selected song
		job.SelectedSongID = &output.SelectedSongID
		job.AudioURL = &selectedAudioURL
		if err := deps.JobRepo.Update(ctx, job); err != nil {
			logger.Error("failed to update job with selected song", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to update job: %v", err))
		}

		logger.Info("song selected",
			zap.String("selected_song_id", output.SelectedSongID),
			zap.String("reasoning", output.Reasoning),
		)

		// Enqueue next task: generate image
		nextPayload, _ := (&TaskPayload{JobID: payload.JobID}).Marshal()
		nextTask := asynq.NewTask(TypeGenerateImage, nextPayload)
		if _, err := deps.AsynqClient.Enqueue(nextTask); err != nil {
			logger.Error("failed to enqueue generate image task", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to enqueue next task: %v", err))
		}

		logger.Info("enqueued generate image task")
		return nil
	}
}

// HandleGenerateImage creates a handler for the generate image task.
// This handler:
// 1. Loads the job
// 2. Creates an ImageConceptAgent
// 3. Generates the image prompt
// 4. Updates the job with image_prompt
// 5. Calls NanoBananaClient.CreateTask()
// 6. Updates the job with nano_task_id
// 7. If webhook is configured, returns nil; otherwise polls for completion
func HandleGenerateImage(deps *Dependencies) asynq.HandlerFunc {
	return func(ctx context.Context, task *asynq.Task) error {
		logger := deps.Logger.With(zap.String("task_type", TypeGenerateImage))

		// Parse payload
		payload, err := UnmarshalTaskPayload(task.Payload())
		if err != nil {
			logger.Error("failed to unmarshal task payload", zap.Error(err))
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		logger = logger.With(zap.String("job_id", payload.JobID.String()))
		logger.Info("starting generate image task")

		// Load job
		job, err := deps.JobRepo.GetByID(ctx, payload.JobID)
		if err != nil {
			logger.Error("failed to load job", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to load job: %v", err))
		}

		// Update status
		job.Status = models.StatusGeneratingImage
		if err := deps.JobRepo.Update(ctx, job); err != nil {
			logger.Error("failed to update job status", zap.Error(err))
		}

		// Get user's API keys
		openRouterKey, kieKey, err := getUserAPIKeys(ctx, deps, job.UserID)
		if err != nil {
			logger.Error("failed to get user API keys", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to get API keys: %v", err))
		}
		if openRouterKey == "" {
			logger.Error("user has no OpenRouter API key")
			return markJobFailed(ctx, deps, payload.JobID, "user has no OpenRouter API key configured")
		}
		if kieKey == "" {
			logger.Error("user has no KIE API key")
			return markJobFailed(ctx, deps, payload.JobID, "user has no KIE API key configured")
		}

		// Determine LLM model
		llmModel := job.LLMModel
		if llmModel == "" {
			llmModel = DefaultLLMModel
		}

		// Get effective prompt from system defaults
		effectivePrompt := getEffectivePrompt(ctx, deps, "image_concept")

		// Create per-user OpenRouter client and ImageConceptAgent
		openRouterClient := openrouter.NewClient(openRouterKey)
		agent := agents.NewImageConceptAgentWithPrompt(openRouterClient, llmModel, logger, effectivePrompt)

		// Build input
		var songTitle, songStyle, lyrics string
		if job.SongPrompt != nil {
			songTitle = job.SongPrompt.Title
			songStyle = job.SongPrompt.Style
			lyrics = job.SongPrompt.Prompt // Lyrics are stored in the prompt
		}

		input := agents.ImageConceptInput{
			OriginalConcept: job.Concept,
			SongTitle:       songTitle,
			SongStyle:       songStyle,
			Lyrics:          lyrics,
		}

		// Generate image prompt
		output, err := agent.Generate(ctx, input)
		if err != nil {
			logger.Error("failed to generate image prompt", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to generate image prompt: %v", err))
		}

		// Update job with image_prompt
		job.ImagePrompt = &models.ImagePrompt{
			Prompt:      output.Prompt,
			AspectRatio: output.AspectRatio,
			Resolution:  output.Resolution,
		}
		if err := deps.JobRepo.Update(ctx, job); err != nil {
			logger.Error("failed to update job with image prompt", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to update job: %v", err))
		}

		logger.Info("image prompt generated", zap.Int("prompt_length", len(output.Prompt)))

		// Create per-user NanoBanana client
		nanoBananaClient := kie.NewNanoBananaClient(kieKey, deps.KIEBaseURL)

		// Build NanoBanana request
		req := kie.CreateTaskRequest{
			Model: kie.ModelNanoBananaPro,
			Input: kie.NanoInput{
				Prompt:       output.Prompt,
				AspectRatio:  output.AspectRatio,
				Resolution:   output.Resolution,
				OutputFormat: kie.FormatPNG,
			},
		}

		// Add webhook URL if configured
		// Route: /api/v1/webhooks/:token/nano/:job_id (matches RegisterRoutes in webhook_handler.go)
		if deps.WebhookBaseURL != "" && deps.WebhookSecret != "" {
			req.CallBackUrl = fmt.Sprintf("%s/api/v1/webhooks/%s/nano/%s", deps.WebhookBaseURL, deps.WebhookSecret, payload.JobID.String())
		}

		// Create image generation task
		nanoTaskID, err := nanoBananaClient.CreateTask(ctx, req)
		if err != nil {
			logger.Error("failed to create image generation task", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to create image task: %v", err))
		}

		logger.Info("image generation started", zap.String("nano_task_id", nanoTaskID))

		// Update job with nano_task_id
		job.NanoTaskID = &nanoTaskID
		if err := deps.JobRepo.Update(ctx, job); err != nil {
			logger.Error("failed to update job with nano task id", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to update job: %v", err))
		}

		// If webhook is configured, return and let webhook handle completion
		if deps.WebhookBaseURL != "" {
			logger.Info("webhook configured, waiting for callback")
			return nil
		}

		// Otherwise, poll for completion
		logger.Info("polling for image generation completion")
		statusResp, err := nanoBananaClient.WaitForCompletion(ctx, nanoTaskID, 5*time.Minute)
		if err != nil {
			logger.Error("image generation failed or timed out", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("image generation failed: %v", err))
		}

		// Update job with image URL (parse from ResultJson)
		imageURL, err := nanoBananaClient.GetImageUrl(statusResp)
		if err != nil {
			logger.Error("failed to extract image URL from response", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to get image URL: %v", err))
		}
		job.ImageURL = &imageURL
		if err := deps.JobRepo.Update(ctx, job); err != nil {
			logger.Error("failed to update job with image url", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to update job: %v", err))
		}

		logger.Info("image generation complete", zap.String("image_url", imageURL))

		// Enqueue next task: process video
		nextPayload, _ := (&TaskPayload{JobID: payload.JobID}).Marshal()
		nextTask := asynq.NewTask(TypeProcessVideo, nextPayload)
		if _, err := deps.AsynqClient.Enqueue(nextTask); err != nil {
			logger.Error("failed to enqueue process video task", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to enqueue next task: %v", err))
		}

		logger.Info("enqueued process video task")
		return nil
	}
}

// HandleProcessVideo creates a handler for the process video task.
// This handler:
// 1. Loads the job (must have audio_url and image_url)
// 2. Uses FFmpegProcessor.CreateMusicVideo()
// 3. Saves video to temp file
// 4. Enqueues TypeUploadAssets
func HandleProcessVideo(deps *Dependencies) asynq.HandlerFunc {
	return func(ctx context.Context, task *asynq.Task) error {
		logger := deps.Logger.With(zap.String("task_type", TypeProcessVideo))

		// Parse payload
		payload, err := UnmarshalTaskPayload(task.Payload())
		if err != nil {
			logger.Error("failed to unmarshal task payload", zap.Error(err))
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		logger = logger.With(zap.String("job_id", payload.JobID.String()))
		logger.Info("starting process video task")

		// Load job
		job, err := deps.JobRepo.GetByID(ctx, payload.JobID)
		if err != nil {
			logger.Error("failed to load job", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to load job: %v", err))
		}

		// Verify required URLs exist
		if job.AudioURL == nil || *job.AudioURL == "" {
			logger.Error("job missing audio_url")
			return markJobFailed(ctx, deps, payload.JobID, "job missing audio_url")
		}
		if job.ImageURL == nil || *job.ImageURL == "" {
			logger.Error("job missing image_url")
			return markJobFailed(ctx, deps, payload.JobID, "job missing image_url")
		}

		// Update status
		job.Status = models.StatusProcessingVideo
		if err := deps.JobRepo.Update(ctx, job); err != nil {
			logger.Error("failed to update job status", zap.Error(err))
		}

		// Create temp output path for video
		tempDir, err := os.MkdirTemp("", "ugc-output-*")
		if err != nil {
			logger.Error("failed to create temp directory", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to create temp directory: %v", err))
		}
		// Note: Don't defer cleanup here - we need the file for upload task

		outputPath := filepath.Join(tempDir, fmt.Sprintf("%s.mp4", payload.JobID.String()))

		// Create music video
		input := ffmpeg.CreateMusicVideoInput{
			AudioURL:   *job.AudioURL,
			ImageURL:   *job.ImageURL,
			OutputPath: outputPath,
		}

		videoOutput, err := deps.FFmpegProcessor.CreateMusicVideo(ctx, input)
		if err != nil {
			logger.Error("failed to create music video", zap.Error(err))
			// Clean up temp directory on error
			os.RemoveAll(tempDir)
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to create video: %v", err))
		}

		logger.Info("video created successfully",
			zap.String("output_path", videoOutput.OutputPath),
			zap.Int64("file_size", videoOutput.FileSize),
			zap.Duration("duration", videoOutput.Duration),
		)

		// Enqueue next task: upload assets
		// Include the video path in metadata for the upload task
		nextPayload, _ := (&TaskPayload{JobID: payload.JobID}).Marshal()
		nextTask := asynq.NewTask(TypeUploadAssets, nextPayload)
		if _, err := deps.AsynqClient.Enqueue(nextTask, asynq.TaskID(fmt.Sprintf("upload-%s", payload.JobID.String()))); err != nil {
			logger.Error("failed to enqueue upload assets task", zap.Error(err))
			os.RemoveAll(tempDir)
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to enqueue next task: %v", err))
		}

		logger.Info("enqueued upload assets task")
		return nil
	}
}

// HandleUploadAssets creates a handler for the upload assets task.
// This handler:
// 1. Loads the job
// 2. Finds the generated video file
// 3. Uploads video to R2
// 4. Updates the job with video_url
// 5. Marks the job as completed
func HandleUploadAssets(deps *Dependencies) asynq.HandlerFunc {
	return func(ctx context.Context, task *asynq.Task) error {
		logger := deps.Logger.With(zap.String("task_type", TypeUploadAssets))

		// Parse payload
		payload, err := UnmarshalTaskPayload(task.Payload())
		if err != nil {
			logger.Error("failed to unmarshal task payload", zap.Error(err))
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		logger = logger.With(zap.String("job_id", payload.JobID.String()))
		logger.Info("starting upload assets task")

		// Load job
		job, err := deps.JobRepo.GetByID(ctx, payload.JobID)
		if err != nil {
			logger.Error("failed to load job", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to load job: %v", err))
		}

		// Update status
		job.Status = models.StatusUploading
		if err := deps.JobRepo.Update(ctx, job); err != nil {
			logger.Error("failed to update job status", zap.Error(err))
		}

		// Find the video file - it should be in a temp directory
		// Look for the file based on the job ID pattern
		pattern := fmt.Sprintf("/tmp/ugc-output-*/%s.mp4", payload.JobID.String())
		matches, err := filepath.Glob(pattern)
		if err != nil || len(matches) == 0 {
			logger.Error("video file not found", zap.String("pattern", pattern))
			return markJobFailed(ctx, deps, payload.JobID, "video file not found")
		}

		videoPath := matches[0]
		logger.Info("found video file", zap.String("path", videoPath))

		// Get parent directory for cleanup later
		tempDir := filepath.Dir(videoPath)
		defer os.RemoveAll(tempDir)

		// Open video file
		videoFile, err := os.Open(videoPath)
		if err != nil {
			logger.Error("failed to open video file", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to open video file: %v", err))
		}
		defer videoFile.Close()

		// Upload to R2
		// Key format: videos/{job_id}.mp4
		r2Key := fmt.Sprintf("videos/%s.mp4", payload.JobID.String())

		if err := deps.R2Client.Upload(ctx, r2Key, videoFile, "video/mp4"); err != nil {
			logger.Error("failed to upload video to R2", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to upload video: %v", err))
		}

		logger.Info("video uploaded to R2", zap.String("key", r2Key))

		// Get public URL
		videoURL := deps.R2Client.GetPublicURL(r2Key)
		if videoURL == "" {
			// If no public URL configured, use presigned URL
			presignedURL, err := deps.R2Client.GetPresignedURL(ctx, r2Key, 24*time.Hour)
			if err != nil {
				logger.Error("failed to generate presigned URL", zap.Error(err))
				return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to generate presigned URL: %v", err))
			}
			videoURL = presignedURL
		}

		// Update job with video URL and mark as completed
		job.VideoURL = &videoURL
		job.Status = models.StatusCompleted
		if err := deps.JobRepo.Update(ctx, job); err != nil {
			logger.Error("failed to update job with video url", zap.Error(err))
			return markJobFailed(ctx, deps, payload.JobID, fmt.Sprintf("failed to update job: %v", err))
		}

		logger.Info("job completed successfully",
			zap.String("video_url", videoURL),
		)

		return nil
	}
}

// markJobFailed updates the job status to failed with the given error message.
// It returns the original error for proper task failure handling.
func markJobFailed(ctx context.Context, deps *Dependencies, jobID uuid.UUID, errorMessage string) error {
	if err := deps.JobRepo.UpdateWithError(ctx, jobID, errorMessage); err != nil {
		deps.Logger.Error("failed to mark job as failed",
			zap.String("job_id", jobID.String()),
			zap.Error(err),
		)
	}
	return fmt.Errorf("%s", errorMessage)
}
