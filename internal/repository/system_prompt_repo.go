package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/jaochai/ugc/internal/database"
	"github.com/jaochai/ugc/internal/models"
)

// ErrSystemPromptNotFound is returned when a system prompt is not found.
var ErrSystemPromptNotFound = errors.New("system prompt not found")

// SystemPromptRepository defines the interface for system prompt data access.
type SystemPromptRepository interface {
	GetByType(ctx context.Context, promptType string) (*models.SystemPrompt, error)
	GetAll(ctx context.Context) ([]models.SystemPrompt, error)
	Update(ctx context.Context, promptType string, content string, updatedBy uuid.UUID) error
}

type systemPromptRepository struct {
	db *database.DB
}

// NewSystemPromptRepository creates a new SystemPromptRepository instance.
func NewSystemPromptRepository(db *database.DB) SystemPromptRepository {
	return &systemPromptRepository{db: db}
}

// GetByType retrieves a system prompt by its type.
func (r *systemPromptRepository) GetByType(ctx context.Context, promptType string) (*models.SystemPrompt, error) {
	query := `
		SELECT id, prompt_type, prompt_content, updated_by, created_at, updated_at
		FROM system_prompts
		WHERE prompt_type = $1
	`

	prompt := &models.SystemPrompt{}
	err := r.db.Pool().QueryRow(ctx, query, promptType).Scan(
		&prompt.ID,
		&prompt.PromptType,
		&prompt.PromptContent,
		&prompt.UpdatedBy,
		&prompt.CreatedAt,
		&prompt.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSystemPromptNotFound
		}
		return nil, fmt.Errorf("failed to get system prompt: %w", err)
	}

	return prompt, nil
}

// GetAll retrieves all system prompts.
func (r *systemPromptRepository) GetAll(ctx context.Context) ([]models.SystemPrompt, error) {
	query := `
		SELECT id, prompt_type, prompt_content, updated_by, created_at, updated_at
		FROM system_prompts
		ORDER BY prompt_type
	`

	rows, err := r.db.Pool().Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query system prompts: %w", err)
	}
	defer rows.Close()

	var prompts []models.SystemPrompt
	for rows.Next() {
		var prompt models.SystemPrompt
		if err := rows.Scan(
			&prompt.ID,
			&prompt.PromptType,
			&prompt.PromptContent,
			&prompt.UpdatedBy,
			&prompt.CreatedAt,
			&prompt.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan system prompt: %w", err)
		}
		prompts = append(prompts, prompt)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating system prompts: %w", err)
	}

	return prompts, nil
}

// Update updates a system prompt's content.
func (r *systemPromptRepository) Update(ctx context.Context, promptType string, content string, updatedBy uuid.UUID) error {
	query := `
		UPDATE system_prompts
		SET prompt_content = $2, updated_by = $3, updated_at = NOW()
		WHERE prompt_type = $1
	`

	result, err := r.db.Pool().Exec(ctx, query, promptType, content, updatedBy)
	if err != nil {
		return fmt.Errorf("failed to update system prompt: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrSystemPromptNotFound
	}

	return nil
}
