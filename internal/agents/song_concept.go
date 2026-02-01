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
// This matches models.SongPrompt for consistency.
type SongConceptOutput struct {
	Prompt       string `json:"prompt"`       // Lyrics/description for Suno
	Style        string `json:"style"`        // Music style (e.g., "pop ballad", "rock", "EDM")
	Title        string `json:"title"`        // Song title
	Model        string `json:"model"`        // Suno model: V4, V4_5, V5
	Instrumental bool   `json:"instrumental"` // Whether the song should be instrumental
}

// ToSongPrompt converts SongConceptOutput to models.SongPrompt.
func (o *SongConceptOutput) ToSongPrompt() *models.SongPrompt {
	return &models.SongPrompt{
		Prompt:       o.Prompt,
		Style:        o.Style,
		Title:        o.Title,
		Model:        o.Model,
		Instrumental: o.Instrumental,
	}
}

// SongConceptAgent analyzes song concepts and generates prompts for Suno AI.
type SongConceptAgent struct {
	*BaseAgent
}

// NewSongConceptAgent creates a new SongConceptAgent instance.
func NewSongConceptAgent(llmClient *openrouter.Client, model string, logger *zap.Logger) *SongConceptAgent {
	return &SongConceptAgent{
		BaseAgent: NewBaseAgent(llmClient, model, logger),
	}
}

// systemPrompt returns the system prompt for the song concept agent.
func (a *SongConceptAgent) systemPrompt(language string) string {
	return fmt.Sprintf(`You are a professional music producer AI specializing in creating optimized prompts for Suno AI music generation.

Your task is to analyze the user's song concept and generate a complete prompt that will produce high-quality music.

Output ONLY valid JSON in this exact format (no markdown, no code blocks, just raw JSON):
{
  "prompt": "lyrics or description for Suno (max 3000 chars)",
  "style": "music genre and style",
  "title": "catchy song title",
  "model": "V4 or V4_5 or V5",
  "instrumental": false
}

Guidelines for each field:

**prompt** (max 3000 characters):
- Write complete song lyrics with verse, chorus, and bridge structure
- Use [Verse], [Chorus], [Bridge], [Outro] tags to structure the song
- Make lyrics emotional, memorable, and catchy
- For %s concepts, write lyrics in %s
- Include vivid imagery and relatable themes
- If the concept is abstract or instrumental, write a descriptive prompt instead

**style** (be specific and detailed):
- Combine genre with mood and instrumentation
- Examples: "Thai pop ballad with piano and strings", "Lo-fi hip hop with jazzy chords", "Epic orchestral cinematic", "Indie folk with acoustic guitar"
- Match the style to the emotional tone of the concept

**title**:
- Create a memorable, catchy title that captures the essence of the song
- Keep it concise (2-5 words typically)
- Can be in %s or English depending on the concept

**model**:
- Use "V4" for standard songs with clear vocals and common genres
- Use "V4_5" for complex arrangements, unique styles, or when higher quality is needed
- Use "V5" for experimental or cutting-edge generation (latest model)

**instrumental**:
- Set to true ONLY if the concept explicitly asks for no vocals or an instrumental piece
- Default to false for most concepts

Remember: Output ONLY the JSON object, no explanations or additional text.`, language, language, language)
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
		zap.String("model", output.Model),
		zap.Bool("instrumental", output.Instrumental),
	)

	return &output, nil
}

// validateOutput validates the SongConceptOutput.
func (a *SongConceptAgent) validateOutput(output *SongConceptOutput) error {
	if output.Prompt == "" {
		return fmt.Errorf("prompt is required")
	}
	if len(output.Prompt) > 3000 {
		return fmt.Errorf("prompt exceeds 3000 character limit")
	}
	if output.Style == "" {
		return fmt.Errorf("style is required")
	}
	if output.Title == "" {
		return fmt.Errorf("title is required")
	}
	if output.Model == "" {
		return fmt.Errorf("model is required")
	}

	// Validate model value
	validModels := map[string]bool{
		"V4":   true,
		"V4_5": true,
		"V5":   true,
	}
	if !validModels[output.Model] {
		return fmt.Errorf("invalid model: %s (must be V4, V4_5, or V5)", output.Model)
	}

	return nil
}
