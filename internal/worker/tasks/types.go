// Package tasks defines task types and handlers for the async worker.
package tasks

import (
	"encoding/json"

	"github.com/google/uuid"
)

// Task type constants for asynq.
const (
	TypeAnalyzeConcept = "job:analyze_concept"
	TypeGenerateMusic  = "job:generate_music"
	TypeSelectSong     = "job:select_song"
	TypeGenerateImage  = "job:generate_image"
	TypeProcessVideo    = "job:process_video"
	TypeUploadAssets    = "job:upload_assets"
	TypeUploadYouTube   = "job:upload_youtube"
)

// TaskPayload represents the common payload for all job-related tasks.
type TaskPayload struct {
	JobID uuid.UUID `json:"job_id"`
}

// Marshal serializes the payload to JSON bytes.
func (p *TaskPayload) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// UnmarshalTaskPayload deserializes JSON bytes into a TaskPayload.
func UnmarshalTaskPayload(data []byte) (*TaskPayload, error) {
	var payload TaskPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}
