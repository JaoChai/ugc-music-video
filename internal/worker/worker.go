// Package worker provides background job processing using Asynq.
package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/jaochai/ugc/internal/external/kie"
	"github.com/jaochai/ugc/internal/external/openrouter"
	"github.com/jaochai/ugc/internal/external/r2"
	"github.com/jaochai/ugc/internal/ffmpeg"
	"github.com/jaochai/ugc/internal/repository"
	"github.com/jaochai/ugc/internal/service"
)

// Task type constants define the types of background jobs.
const (
	TypeAnalyzeConcept  = "job:analyze_concept"
	TypeGenerateMusic   = "job:generate_music"
	TypePollMusicStatus = "job:poll_music_status"
	TypeSelectSong      = "job:select_song"
	TypeGenerateImage   = "job:generate_image"
	TypePollImageStatus = "job:poll_image_status"
	TypeProcessVideo    = "job:process_video"
	TypeUploadAssets    = "job:upload_assets"
)

// TaskPayload is a generic payload for all task types.
type TaskPayload struct {
	JobID uuid.UUID `json:"job_id"`
}

// AnalyzeConceptPayload is the payload for the analyze concept task.
type AnalyzeConceptPayload = TaskPayload

// GenerateMusicPayload is the payload for the generate music task.
type GenerateMusicPayload = TaskPayload

// PollMusicStatusPayload is the payload for polling music generation status.
type PollMusicStatusPayload struct {
	JobID     uuid.UUID `json:"job_id"`
	TaskID    string    `json:"task_id"`
	PollCount int       `json:"poll_count"`
}

// SelectSongPayload is the payload for the select song task.
type SelectSongPayload = TaskPayload

// GenerateImagePayload is the payload for the generate image task.
type GenerateImagePayload = TaskPayload

// PollImageStatusPayload is the payload for polling image generation status.
type PollImageStatusPayload struct {
	JobID     uuid.UUID `json:"job_id"`
	TaskID    string    `json:"task_id"`
	PollCount int       `json:"poll_count"`
}

// ProcessVideoPayload is the payload for the process video task.
type ProcessVideoPayload = TaskPayload

// UploadAssetsPayload is the payload for the upload assets task.
type UploadAssetsPayload = TaskPayload

// Dependencies holds all dependencies needed by task handlers.
type Dependencies struct {
	JobService       service.JobService
	UserRepo         repository.UserRepository
	OpenRouterClient *openrouter.Client
	SunoClient       *kie.SunoClient
	NanoBananaClient *kie.NanoBananaClient
	R2Client         *r2.Client
	FFmpegProcessor  *ffmpeg.Processor
	Logger           *zap.Logger
}

// Worker represents the Asynq worker server.
type Worker struct {
	server *asynq.Server
	mux    *asynq.ServeMux
	logger *zap.Logger
}

// NewWorker creates a new Worker instance.
func NewWorker(redisURL string, deps Dependencies, logger *zap.Logger) (*Worker, error) {
	// Parse Redis URL to get connection options
	redisOpt, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	// Create Asynq server with configuration
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			// Maximum number of concurrent workers
			Concurrency: 10,
			// Queue priorities (higher number = higher priority)
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			// Retry configuration
			RetryDelayFunc: func(n int, e error, t *asynq.Task) time.Duration {
				return time.Duration(n) * time.Minute
			},
			// Error handler for logging
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				logger.Error("task failed",
					zap.String("type", task.Type()),
					zap.ByteString("payload", task.Payload()),
					zap.Error(err),
				)
			}),
			// Logger adapter
			Logger: newAsynqLogger(logger),
		},
	)

	// Create ServeMux and register handlers
	mux := asynq.NewServeMux()

	// Register task handlers (using constants from tasks.go)
	mux.HandleFunc(TypeAnalyzeConcept, newAnalyzeConceptHandler(deps))
	mux.HandleFunc(TypeGenerateMusic, newGenerateMusicHandler(deps))
	mux.HandleFunc(TypePollMusicStatus, newPollMusicStatusHandler(deps))
	mux.HandleFunc(TypeSelectSong, newSelectSongHandler(deps))
	mux.HandleFunc(TypeGenerateImage, newGenerateImageHandler(deps))
	mux.HandleFunc(TypePollImageStatus, newPollImageStatusHandler(deps))
	mux.HandleFunc(TypeProcessVideo, newProcessVideoHandler(deps))
	mux.HandleFunc(TypeUploadAssets, newUploadAssetsHandler(deps))

	return &Worker{
		server: server,
		mux:    mux,
		logger: logger,
	}, nil
}

// Start starts the worker server.
func (w *Worker) Start() error {
	w.logger.Info("starting worker server")
	return w.server.Start(w.mux)
}

// Shutdown gracefully shuts down the worker server.
func (w *Worker) Shutdown() {
	w.logger.Info("shutting down worker server")
	w.server.Shutdown()
}

// EnqueueTask is a helper function to enqueue a task to the queue.
func EnqueueTask(ctx context.Context, client *asynq.Client, taskType string, jobID uuid.UUID, opts ...asynq.Option) error {
	payload := TaskPayload{
		JobID: jobID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := asynq.NewTask(taskType, payloadBytes, opts...)

	info, err := client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	_ = info // info contains task ID, queue, etc.
	return nil
}

// Task handlers

func newAnalyzeConceptHandler(deps Dependencies) asynq.HandlerFunc {
	return func(ctx context.Context, task *asynq.Task) error {
		var payload AnalyzeConceptPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		deps.Logger.Info("processing analyze concept task",
			zap.String("job_id", payload.JobID.String()),
		)

		// TODO: Implement concept analysis logic
		// 1. Get job from database
		// 2. Use OpenRouter to analyze the concept
		// 3. Generate song prompt
		// 4. Update job with song prompt
		// 5. Enqueue next task (generate music)

		return nil
	}
}

func newGenerateMusicHandler(deps Dependencies) asynq.HandlerFunc {
	return func(ctx context.Context, task *asynq.Task) error {
		var payload GenerateMusicPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		deps.Logger.Info("processing generate music task",
			zap.String("job_id", payload.JobID.String()),
		)

		// TODO: Implement music generation logic
		// 1. Get job from database
		// 2. Use Suno client to generate music
		// 3. Enqueue poll music status task

		return nil
	}
}

func newPollMusicStatusHandler(deps Dependencies) asynq.HandlerFunc {
	return func(ctx context.Context, task *asynq.Task) error {
		var payload PollMusicStatusPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		deps.Logger.Info("processing poll music status task",
			zap.String("job_id", payload.JobID.String()),
			zap.String("task_id", payload.TaskID),
			zap.Int("poll_count", payload.PollCount),
		)

		// TODO: Implement music status polling logic
		// 1. Check Suno task status
		// 2. If completed, update job with generated songs and enqueue select song task
		// 3. If still processing, re-enqueue poll task with incremented count
		// 4. If failed or max polls exceeded, mark job as failed

		return nil
	}
}

func newSelectSongHandler(deps Dependencies) asynq.HandlerFunc {
	return func(ctx context.Context, task *asynq.Task) error {
		var payload SelectSongPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		deps.Logger.Info("processing select song task",
			zap.String("job_id", payload.JobID.String()),
		)

		// TODO: Implement song selection logic
		// 1. Get job from database
		// 2. Use OpenRouter to analyze songs and select best one
		// 3. Update job with selected song
		// 4. Enqueue next task (generate image)

		return nil
	}
}

func newGenerateImageHandler(deps Dependencies) asynq.HandlerFunc {
	return func(ctx context.Context, task *asynq.Task) error {
		var payload GenerateImagePayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		deps.Logger.Info("processing generate image task",
			zap.String("job_id", payload.JobID.String()),
		)

		// TODO: Implement image generation logic
		// 1. Get job from database
		// 2. Use OpenRouter to generate image prompt
		// 3. Use NanoBanana client to generate image
		// 4. Enqueue poll image status task

		return nil
	}
}

func newPollImageStatusHandler(deps Dependencies) asynq.HandlerFunc {
	return func(ctx context.Context, task *asynq.Task) error {
		var payload PollImageStatusPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		deps.Logger.Info("processing poll image status task",
			zap.String("job_id", payload.JobID.String()),
			zap.String("task_id", payload.TaskID),
			zap.Int("poll_count", payload.PollCount),
		)

		// TODO: Implement image status polling logic
		// 1. Check NanoBanana task status
		// 2. If completed, update job with image URL and enqueue process video task
		// 3. If still processing, re-enqueue poll task with incremented count
		// 4. If failed or max polls exceeded, mark job as failed

		return nil
	}
}

func newProcessVideoHandler(deps Dependencies) asynq.HandlerFunc {
	return func(ctx context.Context, task *asynq.Task) error {
		var payload ProcessVideoPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		deps.Logger.Info("processing video task",
			zap.String("job_id", payload.JobID.String()),
		)

		// TODO: Implement video processing logic
		// 1. Get job from database
		// 2. Use FFmpeg to combine audio and image into video
		// 3. Update job with video URL
		// 4. Enqueue next task (upload assets)

		return nil
	}
}

func newUploadAssetsHandler(deps Dependencies) asynq.HandlerFunc {
	return func(ctx context.Context, task *asynq.Task) error {
		var payload UploadAssetsPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		deps.Logger.Info("processing upload assets task",
			zap.String("job_id", payload.JobID.String()),
		)

		// TODO: Implement asset upload logic
		// 1. Get job from database
		// 2. Upload audio, image, and video to R2
		// 3. Update job with final URLs
		// 4. Mark job as completed

		return nil
	}
}

// asynqLogger adapts zap.Logger to asynq.Logger interface.
type asynqLogger struct {
	logger *zap.Logger
}

func newAsynqLogger(logger *zap.Logger) *asynqLogger {
	return &asynqLogger{logger: logger.Named("asynq")}
}

func (l *asynqLogger) Debug(args ...interface{}) {
	l.logger.Debug(fmt.Sprint(args...))
}

func (l *asynqLogger) Info(args ...interface{}) {
	l.logger.Info(fmt.Sprint(args...))
}

func (l *asynqLogger) Warn(args ...interface{}) {
	l.logger.Warn(fmt.Sprint(args...))
}

func (l *asynqLogger) Error(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
}

func (l *asynqLogger) Fatal(args ...interface{}) {
	l.logger.Fatal(fmt.Sprint(args...))
}
