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

	"github.com/jaochai/ugc/internal/external/r2"
	"github.com/jaochai/ugc/internal/ffmpeg"
	"github.com/jaochai/ugc/internal/repository"
	"github.com/jaochai/ugc/internal/service"
	"github.com/jaochai/ugc/internal/worker/tasks"
)

// Re-export task type constants for convenience.
const (
	TypeAnalyzeConcept  = tasks.TypeAnalyzeConcept
	TypeGenerateMusic   = tasks.TypeGenerateMusic
	TypeSelectSong      = tasks.TypeSelectSong
	TypeGenerateImage   = tasks.TypeGenerateImage
	TypeProcessVideo    = tasks.TypeProcessVideo
	TypeUploadAssets    = tasks.TypeUploadAssets
)

// TaskPayload is a generic payload for all task types.
type TaskPayload struct {
	JobID uuid.UUID `json:"job_id"`
}

// Dependencies holds all dependencies needed by task handlers.
type Dependencies struct {
	JobRepo         repository.JobRepository
	UserRepo        repository.UserRepository
	CryptoService   service.CryptoService
	R2Client        *r2.Client
	FFmpegProcessor *ffmpeg.Processor
	AsynqClient     *asynq.Client
	Logger          *zap.Logger
	WebhookBaseURL  string // Base URL for webhooks, empty to use polling
	WebhookSecret   string // Secret token for webhook authentication
	KIEBaseURL      string // Base URL for KIE API
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

	// Convert worker.Dependencies to tasks.Dependencies
	taskDeps := &tasks.Dependencies{
		JobRepo:         deps.JobRepo,
		UserRepo:        deps.UserRepo,
		CryptoService:   deps.CryptoService,
		R2Client:        deps.R2Client,
		FFmpegProcessor: deps.FFmpegProcessor,
		AsynqClient:     deps.AsynqClient,
		Logger:          deps.Logger,
		WebhookBaseURL:  deps.WebhookBaseURL,
		WebhookSecret:   deps.WebhookSecret,
		KIEBaseURL:      deps.KIEBaseURL,
	}

	// Register task handlers using real implementations from tasks package
	mux.HandleFunc(tasks.TypeAnalyzeConcept, tasks.HandleAnalyzeConcept(taskDeps))
	mux.HandleFunc(tasks.TypeGenerateMusic, tasks.HandleGenerateMusic(taskDeps))
	mux.HandleFunc(tasks.TypeSelectSong, tasks.HandleSelectSong(taskDeps))
	mux.HandleFunc(tasks.TypeGenerateImage, tasks.HandleGenerateImage(taskDeps))
	mux.HandleFunc(tasks.TypeProcessVideo, tasks.HandleProcessVideo(taskDeps))
	mux.HandleFunc(tasks.TypeUploadAssets, tasks.HandleUploadAssets(taskDeps))

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
