// Package worker provides background task processing using asynq.
package worker

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// NewAnalyzeConceptTask creates a new analyze concept task.
func NewAnalyzeConceptTask(jobID uuid.UUID) (*asynq.Task, error) {
	payload := TaskPayload{
		JobID: jobID,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeAnalyzeConcept, payloadBytes), nil
}

// NewGenerateMusicTask creates a new generate music task.
func NewGenerateMusicTask(jobID uuid.UUID) (*asynq.Task, error) {
	payload := TaskPayload{
		JobID: jobID,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeGenerateMusic, payloadBytes), nil
}

// NewSelectSongTask creates a new select song task.
// Uses TaskID for deduplication to prevent duplicate processing from webhook retries.
func NewSelectSongTask(jobID uuid.UUID) (*asynq.Task, error) {
	payload := TaskPayload{
		JobID: jobID,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	// TaskID ensures only one select song task can be enqueued per job
	taskID := fmt.Sprintf("select-song-%s", jobID.String())
	return asynq.NewTask(TypeSelectSong, payloadBytes, asynq.TaskID(taskID)), nil
}

// NewGenerateImageTask creates a new generate image task.
func NewGenerateImageTask(jobID uuid.UUID) (*asynq.Task, error) {
	payload := TaskPayload{
		JobID: jobID,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeGenerateImage, payloadBytes), nil
}

// NewProcessVideoTask creates a new process video task.
// Uses TaskID for deduplication to prevent duplicate processing from webhook retries.
func NewProcessVideoTask(jobID uuid.UUID) (*asynq.Task, error) {
	payload := TaskPayload{
		JobID: jobID,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	// TaskID ensures only one process video task can be enqueued per job
	taskID := fmt.Sprintf("process-video-%s", jobID.String())
	return asynq.NewTask(TypeProcessVideo, payloadBytes, asynq.TaskID(taskID)), nil
}

// NewUploadAssetsTask creates a new upload assets task.
func NewUploadAssetsTask(jobID uuid.UUID) (*asynq.Task, error) {
	payload := TaskPayload{
		JobID: jobID,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeUploadAssets, payloadBytes), nil
}
