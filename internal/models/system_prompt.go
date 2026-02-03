package models

import (
	"time"

	"github.com/google/uuid"
)

// SystemPrompt represents a system-wide default prompt stored in DB
type SystemPrompt struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	PromptType    string     `json:"prompt_type" gorm:"uniqueIndex;not null"`
	PromptContent string     `json:"prompt_content" gorm:"not null"`
	UpdatedBy     *uuid.UUID `json:"updated_by"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (SystemPrompt) TableName() string {
	return "system_prompts"
}

// UpdateSystemPromptInput represents the input for updating a system prompt
type UpdateSystemPromptInput struct {
	PromptType    string `json:"prompt_type" validate:"required,oneof=song_concept song_selector image_concept"`
	PromptContent string `json:"prompt_content" validate:"required,min=100,max=15000"`
}

// SystemPromptsResponse represents all system prompts
type SystemPromptsResponse struct {
	SongConcept  SystemPrompt `json:"song_concept"`
	SongSelector SystemPrompt `json:"song_selector"`
	ImageConcept SystemPrompt `json:"image_concept"`
}
