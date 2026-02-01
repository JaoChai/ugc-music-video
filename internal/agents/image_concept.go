// Package agents provides AI agents for content generation.
package agents

import (
	"context"
	"fmt"
	"strings"

	"github.com/jaochai/ugc/internal/external/openrouter"
	"go.uber.org/zap"
)

const imageConceptSystemPrompt = `You are a visual artist AI. Create an image prompt for a music video cover/background image.

The image should:
1. Capture the mood and emotion of the song
2. Be visually striking and suitable as a static background for a music video
3. Match the music genre aesthetic
4. Be appropriate for all audiences

Output JSON:
{
  "prompt": "detailed image description for AI generation (be specific about style, colors, composition, mood)",
  "aspectRatio": "16:9",
  "resolution": "1K"
}

Prompt guidelines:
- Start with main subject/scene
- Include art style (photorealistic, anime, abstract, etc.)
- Describe lighting and color palette
- Mention composition and mood
- Keep it under 500 characters
- Always use aspectRatio "16:9" for music videos
- Use "1K" resolution for faster generation, "2K" for higher quality`

// ImageConceptAgent generates image prompts based on song concepts.
type ImageConceptAgent struct {
	*BaseAgent
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
		BaseAgent: NewBaseAgent(llmClient, model, logger),
	}
}

// Generate creates an image prompt based on the song concept and info.
func (a *ImageConceptAgent) Generate(ctx context.Context, input ImageConceptInput) (*ImageConceptOutput, error) {
	a.Logger().Info("generating image concept",
		zap.String("song_title", input.SongTitle),
		zap.String("song_style", input.SongStyle),
	)

	userPrompt := a.buildUserPrompt(input)

	var output ImageConceptOutput
	if err := a.ChatJSON(ctx, imageConceptSystemPrompt, userPrompt, &output); err != nil {
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
