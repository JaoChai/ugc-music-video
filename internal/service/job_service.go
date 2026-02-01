// Package service provides business logic for the UGC application.
package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"

	apperrors "github.com/jaochai/ugc/pkg/errors"
	"github.com/jaochai/ugc/pkg/response"

	"github.com/jaochai/ugc/internal/models"
	"github.com/jaochai/ugc/internal/repository"
)

// JobService defines the interface for job business logic.
type JobService interface {
	Create(ctx context.Context, userID uuid.UUID, input models.CreateJobInput, defaultModel string) (*models.Job, error)
	GetByID(ctx context.Context, userID uuid.UUID, jobID uuid.UUID) (*models.Job, error)
	List(ctx context.Context, userID uuid.UUID, page, perPage int) ([]*models.Job, *response.Meta, error)
	Cancel(ctx context.Context, userID uuid.UUID, jobID uuid.UUID) error
	UpdateStatus(ctx context.Context, jobID uuid.UUID, status string) error
	UpdateSongPrompt(ctx context.Context, jobID uuid.UUID, prompt *models.SongPrompt) error
	UpdateGeneratedSongs(ctx context.Context, jobID uuid.UUID, taskID string, songs []models.GeneratedSong) error
	UpdateSelectedSong(ctx context.Context, jobID uuid.UUID, songID string, audioURL string) error
	UpdateImagePrompt(ctx context.Context, jobID uuid.UUID, prompt *models.ImagePrompt) error
	UpdateImageURL(ctx context.Context, jobID uuid.UUID, taskID string, imageURL string) error
	UpdateVideoURL(ctx context.Context, jobID uuid.UUID, videoURL string) error
	MarkFailed(ctx context.Context, jobID uuid.UUID, errorMessage string) error
	MarkCompleted(ctx context.Context, jobID uuid.UUID) error
}

// jobService implements JobService.
type jobService struct {
	jobRepo repository.JobRepository
	logger  *zap.Logger
}

// NewJobService creates a new JobService instance.
func NewJobService(jobRepo repository.JobRepository, logger *zap.Logger) JobService {
	return &jobService{
		jobRepo: jobRepo,
		logger:  logger,
	}
}

// Create creates a new job with pending status.
func (s *jobService) Create(ctx context.Context, userID uuid.UUID, input models.CreateJobInput, defaultModel string) (*models.Job, error) {
	// Determine which model to use
	model := defaultModel
	if input.Model != nil && *input.Model != "" {
		model = *input.Model
	}

	job := &models.Job{
		ID:       uuid.New(),
		UserID:   userID,
		Status:   models.StatusPending,
		Concept:  input.Concept,
		LLMModel: model,
	}

	if err := s.jobRepo.Create(ctx, job); err != nil {
		s.logger.Error("failed to create job",
			zap.Error(err),
			zap.String("user_id", userID.String()),
		)
		return nil, apperrors.NewInternalError(err)
	}

	s.logger.Info("job created",
		zap.String("job_id", job.ID.String()),
		zap.String("user_id", userID.String()),
		zap.String("model", model),
	)

	return job, nil
}

// GetByID retrieves a job by ID and verifies ownership.
func (s *jobService) GetByID(ctx context.Context, userID uuid.UUID, jobID uuid.UUID) (*models.Job, error) {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		if errors.Is(err, repository.ErrJobNotFound) {
			return nil, apperrors.NewNotFound("job not found")
		}
		s.logger.Error("failed to get job",
			zap.Error(err),
			zap.String("job_id", jobID.String()),
		)
		return nil, apperrors.NewInternalError(err)
	}

	// Verify ownership
	if job.UserID != userID {
		s.logger.Warn("unauthorized job access attempt",
			zap.String("job_id", jobID.String()),
			zap.String("owner_id", job.UserID.String()),
			zap.String("requester_id", userID.String()),
		)
		return nil, apperrors.NewForbidden("you do not have access to this job")
	}

	return job, nil
}

// List retrieves paginated jobs for a user.
func (s *jobService) List(ctx context.Context, userID uuid.UUID, page, perPage int) ([]*models.Job, *response.Meta, error) {
	// Set defaults
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}

	jobs, total, err := s.jobRepo.GetByUserID(ctx, userID, page, perPage)
	if err != nil {
		s.logger.Error("failed to list jobs",
			zap.Error(err),
			zap.String("user_id", userID.String()),
		)
		return nil, nil, apperrors.NewInternalError(err)
	}

	meta := response.NewMeta(page, perPage, total)

	return jobs, meta, nil
}

// Cancel cancels a job if it's not in a terminal state.
func (s *jobService) Cancel(ctx context.Context, userID uuid.UUID, jobID uuid.UUID) error {
	// First verify ownership
	job, err := s.GetByID(ctx, userID, jobID)
	if err != nil {
		return err
	}

	// Check if job can be cancelled
	if job.IsTerminal() {
		return apperrors.NewBadRequest("cannot cancel a job that is already completed or failed")
	}

	// Update status to failed with cancellation message
	if err := s.jobRepo.UpdateWithError(ctx, jobID, "job cancelled by user"); err != nil {
		if errors.Is(err, repository.ErrJobNotFound) {
			return apperrors.NewNotFound("job not found")
		}
		s.logger.Error("failed to cancel job",
			zap.Error(err),
			zap.String("job_id", jobID.String()),
		)
		return apperrors.NewInternalError(err)
	}

	s.logger.Info("job cancelled",
		zap.String("job_id", jobID.String()),
		zap.String("user_id", userID.String()),
	)

	return nil
}

// UpdateStatus updates the status of a job.
func (s *jobService) UpdateStatus(ctx context.Context, jobID uuid.UUID, status string) error {
	if err := s.jobRepo.UpdateStatus(ctx, jobID, status); err != nil {
		if errors.Is(err, repository.ErrJobNotFound) {
			return apperrors.NewNotFound("job not found")
		}
		s.logger.Error("failed to update job status",
			zap.Error(err),
			zap.String("job_id", jobID.String()),
			zap.String("status", status),
		)
		return apperrors.NewInternalError(err)
	}

	s.logger.Debug("job status updated",
		zap.String("job_id", jobID.String()),
		zap.String("status", status),
	)

	return nil
}

// UpdateSongPrompt updates the song prompt for a job.
func (s *jobService) UpdateSongPrompt(ctx context.Context, jobID uuid.UUID, prompt *models.SongPrompt) error {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		if errors.Is(err, repository.ErrJobNotFound) {
			return apperrors.NewNotFound("job not found")
		}
		s.logger.Error("failed to get job for song prompt update",
			zap.Error(err),
			zap.String("job_id", jobID.String()),
		)
		return apperrors.NewInternalError(err)
	}

	job.SongPrompt = prompt
	job.Status = models.StatusGeneratingMusic

	if err := s.jobRepo.Update(ctx, job); err != nil {
		s.logger.Error("failed to update song prompt",
			zap.Error(err),
			zap.String("job_id", jobID.String()),
		)
		return apperrors.NewInternalError(err)
	}

	s.logger.Debug("song prompt updated",
		zap.String("job_id", jobID.String()),
	)

	return nil
}

// UpdateGeneratedSongs updates the generated songs and task ID for a job.
func (s *jobService) UpdateGeneratedSongs(ctx context.Context, jobID uuid.UUID, taskID string, songs []models.GeneratedSong) error {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		if errors.Is(err, repository.ErrJobNotFound) {
			return apperrors.NewNotFound("job not found")
		}
		s.logger.Error("failed to get job for generated songs update",
			zap.Error(err),
			zap.String("job_id", jobID.String()),
		)
		return apperrors.NewInternalError(err)
	}

	job.SunoTaskID = &taskID
	job.GeneratedSongs = songs
	job.Status = models.StatusSelectingSong

	if err := s.jobRepo.Update(ctx, job); err != nil {
		s.logger.Error("failed to update generated songs",
			zap.Error(err),
			zap.String("job_id", jobID.String()),
		)
		return apperrors.NewInternalError(err)
	}

	s.logger.Debug("generated songs updated",
		zap.String("job_id", jobID.String()),
		zap.String("task_id", taskID),
		zap.Int("song_count", len(songs)),
	)

	return nil
}

// UpdateSelectedSong updates the selected song ID and audio URL.
func (s *jobService) UpdateSelectedSong(ctx context.Context, jobID uuid.UUID, songID string, audioURL string) error {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		if errors.Is(err, repository.ErrJobNotFound) {
			return apperrors.NewNotFound("job not found")
		}
		s.logger.Error("failed to get job for selected song update",
			zap.Error(err),
			zap.String("job_id", jobID.String()),
		)
		return apperrors.NewInternalError(err)
	}

	job.SelectedSongID = &songID
	job.AudioURL = &audioURL
	job.Status = models.StatusGeneratingImage

	if err := s.jobRepo.Update(ctx, job); err != nil {
		s.logger.Error("failed to update selected song",
			zap.Error(err),
			zap.String("job_id", jobID.String()),
		)
		return apperrors.NewInternalError(err)
	}

	s.logger.Debug("selected song updated",
		zap.String("job_id", jobID.String()),
		zap.String("song_id", songID),
	)

	return nil
}

// UpdateImagePrompt updates the image prompt for a job.
func (s *jobService) UpdateImagePrompt(ctx context.Context, jobID uuid.UUID, prompt *models.ImagePrompt) error {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		if errors.Is(err, repository.ErrJobNotFound) {
			return apperrors.NewNotFound("job not found")
		}
		s.logger.Error("failed to get job for image prompt update",
			zap.Error(err),
			zap.String("job_id", jobID.String()),
		)
		return apperrors.NewInternalError(err)
	}

	job.ImagePrompt = prompt

	if err := s.jobRepo.Update(ctx, job); err != nil {
		s.logger.Error("failed to update image prompt",
			zap.Error(err),
			zap.String("job_id", jobID.String()),
		)
		return apperrors.NewInternalError(err)
	}

	s.logger.Debug("image prompt updated",
		zap.String("job_id", jobID.String()),
	)

	return nil
}

// UpdateImageURL updates the image URL and task ID for a job.
func (s *jobService) UpdateImageURL(ctx context.Context, jobID uuid.UUID, taskID string, imageURL string) error {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		if errors.Is(err, repository.ErrJobNotFound) {
			return apperrors.NewNotFound("job not found")
		}
		s.logger.Error("failed to get job for image URL update",
			zap.Error(err),
			zap.String("job_id", jobID.String()),
		)
		return apperrors.NewInternalError(err)
	}

	job.NanoTaskID = &taskID
	job.ImageURL = &imageURL
	job.Status = models.StatusProcessingVideo

	if err := s.jobRepo.Update(ctx, job); err != nil {
		s.logger.Error("failed to update image URL",
			zap.Error(err),
			zap.String("job_id", jobID.String()),
		)
		return apperrors.NewInternalError(err)
	}

	s.logger.Debug("image URL updated",
		zap.String("job_id", jobID.String()),
		zap.String("task_id", taskID),
	)

	return nil
}

// UpdateVideoURL updates the video URL for a job.
func (s *jobService) UpdateVideoURL(ctx context.Context, jobID uuid.UUID, videoURL string) error {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		if errors.Is(err, repository.ErrJobNotFound) {
			return apperrors.NewNotFound("job not found")
		}
		s.logger.Error("failed to get job for video URL update",
			zap.Error(err),
			zap.String("job_id", jobID.String()),
		)
		return apperrors.NewInternalError(err)
	}

	job.VideoURL = &videoURL
	job.Status = models.StatusUploading

	if err := s.jobRepo.Update(ctx, job); err != nil {
		s.logger.Error("failed to update video URL",
			zap.Error(err),
			zap.String("job_id", jobID.String()),
		)
		return apperrors.NewInternalError(err)
	}

	s.logger.Debug("video URL updated",
		zap.String("job_id", jobID.String()),
	)

	return nil
}

// MarkFailed marks a job as failed with an error message.
func (s *jobService) MarkFailed(ctx context.Context, jobID uuid.UUID, errorMessage string) error {
	if err := s.jobRepo.UpdateWithError(ctx, jobID, errorMessage); err != nil {
		if errors.Is(err, repository.ErrJobNotFound) {
			return apperrors.NewNotFound("job not found")
		}
		s.logger.Error("failed to mark job as failed",
			zap.Error(err),
			zap.String("job_id", jobID.String()),
		)
		return apperrors.NewInternalError(err)
	}

	s.logger.Info("job marked as failed",
		zap.String("job_id", jobID.String()),
		zap.String("error_message", errorMessage),
	)

	return nil
}

// MarkCompleted marks a job as completed.
func (s *jobService) MarkCompleted(ctx context.Context, jobID uuid.UUID) error {
	if err := s.jobRepo.UpdateStatus(ctx, jobID, models.StatusCompleted); err != nil {
		if errors.Is(err, repository.ErrJobNotFound) {
			return apperrors.NewNotFound("job not found")
		}
		s.logger.Error("failed to mark job as completed",
			zap.Error(err),
			zap.String("job_id", jobID.String()),
		)
		return apperrors.NewInternalError(err)
	}

	s.logger.Info("job completed",
		zap.String("job_id", jobID.String()),
	)

	return nil
}
