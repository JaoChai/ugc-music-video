// Package repository provides data access layer for the UGC application.
package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/jaochai/ugc/internal/database"
	"github.com/jaochai/ugc/internal/models"
)

// ErrJobNotFound is returned when a job is not found.
var ErrJobNotFound = errors.New("job not found")

// JobRepository defines the interface for job data access.
type JobRepository interface {
	Create(ctx context.Context, job *models.Job) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Job, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, page, perPage int) ([]*models.Job, int64, error)
	GetBySunoTaskID(ctx context.Context, taskID string) (*models.Job, error)
	GetByNanoTaskID(ctx context.Context, taskID string) (*models.Job, error)
	Update(ctx context.Context, job *models.Job) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateWithError(ctx context.Context, id uuid.UUID, errorMessage string) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// jobRepository implements JobRepository using PostgreSQL.
type jobRepository struct {
	db *database.DB
}

// NewJobRepository creates a new JobRepository instance.
func NewJobRepository(db *database.DB) JobRepository {
	return &jobRepository{db: db}
}

// Create inserts a new job into the database.
func (r *jobRepository) Create(ctx context.Context, job *models.Job) error {
	songPromptJSON, err := marshalJSONB(job.SongPrompt)
	if err != nil {
		return fmt.Errorf("failed to marshal song_prompt: %w", err)
	}

	generatedSongsJSON, err := marshalJSONB(job.GeneratedSongs)
	if err != nil {
		return fmt.Errorf("failed to marshal generated_songs: %w", err)
	}

	imagePromptJSON, err := marshalJSONB(job.ImagePrompt)
	if err != nil {
		return fmt.Errorf("failed to marshal image_prompt: %w", err)
	}

	query := `
		INSERT INTO jobs (
			id, user_id, status, concept, llm_model,
			song_prompt, suno_task_id, generated_songs, selected_song_id,
			image_prompt, nano_task_id, audio_url, image_url, video_url,
			error_message, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12, $13, $14,
			$15, $16, $17
		)
	`

	now := time.Now().UTC()
	if job.ID == uuid.Nil {
		job.ID = uuid.New()
	}
	job.CreatedAt = now
	job.UpdatedAt = now

	_, err = r.db.Pool().Exec(ctx, query,
		job.ID,
		job.UserID,
		job.Status,
		job.Concept,
		job.LLMModel,
		songPromptJSON,
		job.SunoTaskID,
		generatedSongsJSON,
		job.SelectedSongID,
		imagePromptJSON,
		job.NanoTaskID,
		job.AudioURL,
		job.ImageURL,
		job.VideoURL,
		job.ErrorMessage,
		job.CreatedAt,
		job.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	return nil
}

// GetByID retrieves a job by its ID.
func (r *jobRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Job, error) {
	query := `
		SELECT
			id, user_id, status, concept, llm_model,
			song_prompt, suno_task_id, generated_songs, selected_song_id,
			image_prompt, nano_task_id, audio_url, image_url, video_url,
			error_message, created_at, updated_at
		FROM jobs
		WHERE id = $1
	`

	row := r.db.Pool().QueryRow(ctx, query, id)
	job, err := scanJob(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrJobNotFound
		}
		return nil, fmt.Errorf("failed to get job by id: %w", err)
	}

	return job, nil
}

// GetBySunoTaskID retrieves a job by its Suno task ID.
func (r *jobRepository) GetBySunoTaskID(ctx context.Context, taskID string) (*models.Job, error) {
	query := `
		SELECT
			id, user_id, status, concept, llm_model,
			song_prompt, suno_task_id, generated_songs, selected_song_id,
			image_prompt, nano_task_id, audio_url, image_url, video_url,
			error_message, created_at, updated_at
		FROM jobs
		WHERE suno_task_id = $1
	`

	row := r.db.Pool().QueryRow(ctx, query, taskID)
	job, err := scanJob(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrJobNotFound
		}
		return nil, fmt.Errorf("failed to get job by suno_task_id: %w", err)
	}

	return job, nil
}

// GetByNanoTaskID retrieves a job by its Nano task ID.
func (r *jobRepository) GetByNanoTaskID(ctx context.Context, taskID string) (*models.Job, error) {
	query := `
		SELECT
			id, user_id, status, concept, llm_model,
			song_prompt, suno_task_id, generated_songs, selected_song_id,
			image_prompt, nano_task_id, audio_url, image_url, video_url,
			error_message, created_at, updated_at
		FROM jobs
		WHERE nano_task_id = $1
	`

	row := r.db.Pool().QueryRow(ctx, query, taskID)
	job, err := scanJob(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrJobNotFound
		}
		return nil, fmt.Errorf("failed to get job by nano_task_id: %w", err)
	}

	return job, nil
}

// GetByUserID retrieves jobs for a user with pagination.
func (r *jobRepository) GetByUserID(ctx context.Context, userID uuid.UUID, page, perPage int) ([]*models.Job, int64, error) {
	// Calculate offset
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	offset := (page - 1) * perPage

	// Get total count
	countQuery := `SELECT COUNT(*) FROM jobs WHERE user_id = $1`
	var total int64
	err := r.db.Pool().QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count jobs: %w", err)
	}

	// Get jobs with pagination
	query := `
		SELECT
			id, user_id, status, concept, llm_model,
			song_prompt, suno_task_id, generated_songs, selected_song_id,
			image_prompt, nano_task_id, audio_url, image_url, video_url,
			error_message, created_at, updated_at
		FROM jobs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool().Query(ctx, query, userID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query jobs: %w", err)
	}
	defer rows.Close()

	jobs := make([]*models.Job, 0)
	for rows.Next() {
		job, err := scanJobFromRows(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan job: %w", err)
		}
		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating jobs: %w", err)
	}

	return jobs, total, nil
}

// Update updates all fields of a job.
func (r *jobRepository) Update(ctx context.Context, job *models.Job) error {
	songPromptJSON, err := marshalJSONB(job.SongPrompt)
	if err != nil {
		return fmt.Errorf("failed to marshal song_prompt: %w", err)
	}

	generatedSongsJSON, err := marshalJSONB(job.GeneratedSongs)
	if err != nil {
		return fmt.Errorf("failed to marshal generated_songs: %w", err)
	}

	imagePromptJSON, err := marshalJSONB(job.ImagePrompt)
	if err != nil {
		return fmt.Errorf("failed to marshal image_prompt: %w", err)
	}

	query := `
		UPDATE jobs SET
			status = $2,
			concept = $3,
			llm_model = $4,
			song_prompt = $5,
			suno_task_id = $6,
			generated_songs = $7,
			selected_song_id = $8,
			image_prompt = $9,
			nano_task_id = $10,
			audio_url = $11,
			image_url = $12,
			video_url = $13,
			error_message = $14,
			updated_at = $15
		WHERE id = $1
	`

	job.UpdatedAt = time.Now().UTC()

	result, err := r.db.Pool().Exec(ctx, query,
		job.ID,
		job.Status,
		job.Concept,
		job.LLMModel,
		songPromptJSON,
		job.SunoTaskID,
		generatedSongsJSON,
		job.SelectedSongID,
		imagePromptJSON,
		job.NanoTaskID,
		job.AudioURL,
		job.ImageURL,
		job.VideoURL,
		job.ErrorMessage,
		job.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrJobNotFound
	}

	return nil
}

// UpdateStatus updates only the status of a job.
func (r *jobRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `
		UPDATE jobs SET
			status = $2,
			updated_at = $3
		WHERE id = $1
	`

	result, err := r.db.Pool().Exec(ctx, query, id, status, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrJobNotFound
	}

	return nil
}

// UpdateWithError updates the job status to failed and sets the error message.
func (r *jobRepository) UpdateWithError(ctx context.Context, id uuid.UUID, errorMessage string) error {
	query := `
		UPDATE jobs SET
			status = $2,
			error_message = $3,
			updated_at = $4
		WHERE id = $1
	`

	result, err := r.db.Pool().Exec(ctx, query, id, models.StatusFailed, errorMessage, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to update job with error: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrJobNotFound
	}

	return nil
}

// Delete removes a job from the database.
func (r *jobRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM jobs WHERE id = $1`

	result, err := r.db.Pool().Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrJobNotFound
	}

	return nil
}

// Helper functions for JSONB handling

// marshalJSONB marshals a value to JSON bytes for JSONB storage.
// Returns nil if the value is nil.
func marshalJSONB(v interface{}) ([]byte, error) {
	if v == nil {
		return nil, nil
	}

	// Check for empty slice
	switch val := v.(type) {
	case []models.GeneratedSong:
		if len(val) == 0 {
			return nil, nil
		}
	}

	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// unmarshalJSONB unmarshals JSON bytes to a value.
// If data is nil or empty, the target is left unchanged.
func unmarshalJSONB(data []byte, v interface{}) error {
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, v)
}

// scanJob scans a single row into a Job struct.
func scanJob(row pgx.Row) (*models.Job, error) {
	var job models.Job
	var songPromptJSON, generatedSongsJSON, imagePromptJSON []byte

	err := row.Scan(
		&job.ID,
		&job.UserID,
		&job.Status,
		&job.Concept,
		&job.LLMModel,
		&songPromptJSON,
		&job.SunoTaskID,
		&generatedSongsJSON,
		&job.SelectedSongID,
		&imagePromptJSON,
		&job.NanoTaskID,
		&job.AudioURL,
		&job.ImageURL,
		&job.VideoURL,
		&job.ErrorMessage,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Unmarshal JSONB fields
	if len(songPromptJSON) > 0 {
		var sp models.SongPrompt
		if err := unmarshalJSONB(songPromptJSON, &sp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal song_prompt: %w", err)
		}
		job.SongPrompt = &sp
	}

	if len(generatedSongsJSON) > 0 {
		var gs []models.GeneratedSong
		if err := unmarshalJSONB(generatedSongsJSON, &gs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal generated_songs: %w", err)
		}
		job.GeneratedSongs = gs
	}

	if len(imagePromptJSON) > 0 {
		var ip models.ImagePrompt
		if err := unmarshalJSONB(imagePromptJSON, &ip); err != nil {
			return nil, fmt.Errorf("failed to unmarshal image_prompt: %w", err)
		}
		job.ImagePrompt = &ip
	}

	return &job, nil
}

// scanJobFromRows scans a row from pgx.Rows into a Job struct.
func scanJobFromRows(rows pgx.Rows) (*models.Job, error) {
	var job models.Job
	var songPromptJSON, generatedSongsJSON, imagePromptJSON []byte

	err := rows.Scan(
		&job.ID,
		&job.UserID,
		&job.Status,
		&job.Concept,
		&job.LLMModel,
		&songPromptJSON,
		&job.SunoTaskID,
		&generatedSongsJSON,
		&job.SelectedSongID,
		&imagePromptJSON,
		&job.NanoTaskID,
		&job.AudioURL,
		&job.ImageURL,
		&job.VideoURL,
		&job.ErrorMessage,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Unmarshal JSONB fields
	if len(songPromptJSON) > 0 {
		var sp models.SongPrompt
		if err := unmarshalJSONB(songPromptJSON, &sp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal song_prompt: %w", err)
		}
		job.SongPrompt = &sp
	}

	if len(generatedSongsJSON) > 0 {
		var gs []models.GeneratedSong
		if err := unmarshalJSONB(generatedSongsJSON, &gs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal generated_songs: %w", err)
		}
		job.GeneratedSongs = gs
	}

	if len(imagePromptJSON) > 0 {
		var ip models.ImagePrompt
		if err := unmarshalJSONB(imagePromptJSON, &ip); err != nil {
			return nil, fmt.Errorf("failed to unmarshal image_prompt: %w", err)
		}
		job.ImagePrompt = &ip
	}

	return &job, nil
}
