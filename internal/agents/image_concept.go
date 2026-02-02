// Package agents provides AI agents for content generation.
package agents

import (
	"context"
	"fmt"
	"strings"

	"github.com/jaochai/ugc/internal/external/openrouter"
	"go.uber.org/zap"
)


// ImageConceptAgent generates image prompts based on song concepts.
type ImageConceptAgent struct {
	*BaseAgent
	customPrompt *string
}

// ImageConceptInput contains the input data for image concept generation.
type ImageConceptInput struct {
	OriginalConcept string // concept from user
	SongTitle       string // title of the song
	SongStyle       string // music style used
	Lyrics          string // optional, if available
}

// ImageConceptOutput contains the generated image prompt data.
// This matches models.ImagePrompt structure.
type ImageConceptOutput struct {
	Prompt      string `json:"prompt"`
	AspectRatio string `json:"aspectRatio"`
	Resolution  string `json:"resolution"`
}

// NewImageConceptAgent creates a new ImageConceptAgent.
func NewImageConceptAgent(llmClient *openrouter.Client, model string, logger *zap.Logger) *ImageConceptAgent {
	return &ImageConceptAgent{
		BaseAgent:    NewBaseAgent(llmClient, model, logger),
		customPrompt: nil,
	}
}

// NewImageConceptAgentWithPrompt creates a new ImageConceptAgent with a custom system prompt.
func NewImageConceptAgentWithPrompt(llmClient *openrouter.Client, model string, logger *zap.Logger, customPrompt *string) *ImageConceptAgent {
	return &ImageConceptAgent{
		BaseAgent:    NewBaseAgent(llmClient, model, logger),
		customPrompt: customPrompt,
	}
}

// getSystemPrompt returns the system prompt for the image concept agent.
func (a *ImageConceptAgent) getSystemPrompt() string {
	if a.customPrompt != nil && *a.customPrompt != "" {
		return *a.customPrompt
	}
	return DefaultImageConceptPrompt
}

// Generate creates an image prompt based on the song concept and info.
func (a *ImageConceptAgent) Generate(ctx context.Context, input ImageConceptInput) (*ImageConceptOutput, error) {
	a.Logger().Info("generating image concept",
		zap.String("song_title", input.SongTitle),
		zap.String("song_style", input.SongStyle),
	)

	userPrompt := a.buildUserPrompt(input)

	var output ImageConceptOutput
	if err := a.ChatJSON(ctx, a.getSystemPrompt(), userPrompt, &output); err != nil {
		a.Logger().Error("failed to generate image concept",
			zap.Error(err),
			zap.String("song_title", input.SongTitle),
		)
		return nil, fmt.Errorf("failed to generate image concept: %w", err)
	}

	// Validate prompt is not empty
	if output.Prompt == "" {
		return nil, fmt.Errorf("empty prompt in response")
	}

	// Set defaults if not provided
	if output.AspectRatio == "" {
		output.AspectRatio = "16:9"
	}
	if output.Resolution == "" {
		output.Resolution = "1K"
	}

	a.Logger().Info("image concept generated successfully",
		zap.String("song_title", input.SongTitle),
		zap.Int("prompt_length", len(output.Prompt)),
	)

	return &output, nil
}

// buildUserPrompt creates the user prompt from the input.
func (a *ImageConceptAgent) buildUserPrompt(input ImageConceptInput) string {
	var sb strings.Builder

	sb.WriteString("Create an image prompt for a music video with the following details:\n\n")

	sb.WriteString(fmt.Sprintf("Original Concept: %s\n", input.OriginalConcept))
	sb.WriteString(fmt.Sprintf("Song Title: %s\n", input.SongTitle))
	sb.WriteString(fmt.Sprintf("Music Style: %s\n", input.SongStyle))

	if input.Lyrics != "" {
		sb.WriteString(fmt.Sprintf("\nLyrics:\n%s\n", input.Lyrics))
	}

	sb.WriteString("\nGenerate a visually compelling image prompt that captures the essence of this song.")

	return sb.String()
}
