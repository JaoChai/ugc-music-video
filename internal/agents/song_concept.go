// Package agents provides AI agents for content generation tasks.
package agents

import (
	"context"
	"fmt"

	"github.com/jaochai/ugc/internal/external/openrouter"
	"github.com/jaochai/ugc/internal/models"
	"go.uber.org/zap"
)

// SongConceptInput represents the input for song concept analysis.
type SongConceptInput struct {
	Concept  string // User's song idea/concept
	Language string // Language for lyrics (default: "Thai")
}

// SongConceptOutput represents the output from song concept analysis.
// Note: Model is NOT included - LLM doesn't have knowledge about Suno API versions.
// Model is hardcoded to "V5" in ToSongPrompt().
type SongConceptOutput struct {
	Prompt       string `json:"prompt"`       // Lyrics/description for Suno
	Style        string `json:"style"`        // Music style (e.g., "pop ballad", "rock", "EDM")
	Title        string `json:"title"`        // Song title (Thai)
	TitleEn      string `json:"title_en"`     // Song title (English translation)
	Instrumental bool   `json:"instrumental"` // Whether the song should be instrumental
}

// ToSongPrompt converts SongConceptOutput to models.SongPrompt.
// Model is hardcoded to "V5" - LLM doesn't have knowledge about Suno API versions.
func (o *SongConceptOutput) ToSongPrompt() *models.SongPrompt {
	return &models.SongPrompt{
		Prompt:       o.Prompt,
		Style:        o.Style,
		Title:        o.Title,
		TitleEn:      o.TitleEn,
		Model:        "V5", // Hardcoded - LLM cannot reliably select Suno model versions
		Instrumental: o.Instrumental,
	}
}

// SongConceptAgent analyzes song concepts and generates prompts for Suno AI.
type SongConceptAgent struct {
	*BaseAgent
	customPrompt *string
}

// NewSongConceptAgent creates a new SongConceptAgent instance.
func NewSongConceptAgent(llmClient *openrouter.Client, model string, logger *zap.Logger) *SongConceptAgent {
	return &SongConceptAgent{
		BaseAgent:    NewBaseAgent(llmClient, model, logger),
		customPrompt: nil,
	}
}

// NewSongConceptAgentWithPrompt creates a new SongConceptAgent with a custom system prompt.
func NewSongConceptAgentWithPrompt(llmClient *openrouter.Client, model string, logger *zap.Logger, customPrompt *string) *SongConceptAgent {
	return &SongConceptAgent{
		BaseAgent:    NewBaseAgent(llmClient, model, logger),
		customPrompt: customPrompt,
	}
}

// systemPrompt returns the system prompt for the song concept agent.
// If a custom prompt is set, it will be used instead of the default.
func (a *SongConceptAgent) systemPrompt(language string) string {
	if a.customPrompt != nil && *a.customPrompt != "" {
		// Use custom prompt - replace {{LANGUAGE}} placeholder if present
		return fmt.Sprintf(*a.customPrompt, language, language, language)
	}
	// Use default prompt template
	return fmt.Sprintf(DefaultSongConceptPromptTemplate, language, language, language)
}

// Analyze processes a song concept and generates an optimized Suno prompt.
func (a *SongConceptAgent) Analyze(ctx context.Context, input SongConceptInput) (*SongConceptOutput, error) {
	// Set default language if not provided
	language := input.Language
	if language == "" {
		language = "Thai"
	}

	a.Logger().Info("analyzing song concept",
		zap.String("concept", truncateString(input.Concept, 100)),
		zap.String("language", language),
	)

	// Build user prompt
	userPrompt := fmt.Sprintf("Song concept: %s\n\nGenerate the Suno AI prompt for this concept.", input.Concept)

	// Use ChatJSON to get structured output
	var output SongConceptOutput
	if err := a.ChatJSON(ctx, a.systemPrompt(language), userPrompt, &output); err != nil {
		a.Logger().Error("failed to analyze song concept",
			zap.Error(err),
			zap.String("concept", truncateString(input.Concept, 100)),
		)
		return nil, fmt.Errorf("song concept analysis failed: %w", err)
	}

	// Validate the output
	if err := a.validateOutput(&output); err != nil {
		a.Logger().Error("invalid output from LLM",
			zap.Error(err),
		)
		return nil, fmt.Errorf("invalid output: %w", err)
	}

	a.Logger().Info("song concept analysis complete",
		zap.String("title", output.Title),
		zap.String("style", output.Style),
		zap.Bool("instrumental", output.Instrumental),
	)

	return &output, nil
}

// validateOutput validates the SongConceptOutput.
func (a *SongConceptAgent) validateOutput(output *SongConceptOutput) error {
	if output.Prompt == "" {
		return fmt.Errorf("prompt is required")
	}
	// Suno V5 supports up to 5000 characters for prompt
	if len(output.Prompt) > 5000 {
		return fmt.Errorf("prompt exceeds 5000 character limit")
	}
	if output.Style == "" {
		return fmt.Errorf("style is required")
	}
	if output.Title == "" {
		return fmt.Errorf("title is required")
	}
	return nil
}
