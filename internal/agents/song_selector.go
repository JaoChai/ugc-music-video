// Package agents provides AI agents for content generation.
package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jaochai/ugc/internal/external/openrouter"
	"go.uber.org/zap"
)


// SongCandidate represents a song candidate from Suno.
type SongCandidate struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Duration float64 `json:"duration"`
	AudioURL string  `json:"audio_url"`
}

// SongSelectorInput is the input for the song selector agent.
type SongSelectorInput struct {
	OriginalConcept string          `json:"original_concept"`
	Songs           []SongCandidate `json:"songs"`
}

// SongSelectorOutput is the output from the song selector agent.
type SongSelectorOutput struct {
	SelectedSongID string `json:"selectedSongId"`
	Reasoning      string `json:"reasoning"`
}

// SongSelectorAgent selects the best song from candidates based on the original concept.
type SongSelectorAgent struct {
	*BaseAgent
	customPrompt *string
}

// NewSongSelectorAgent creates a new SongSelectorAgent.
func NewSongSelectorAgent(llmClient *openrouter.Client, model string, logger *zap.Logger) *SongSelectorAgent {
	return &SongSelectorAgent{
		BaseAgent:    NewBaseAgent(llmClient, model, logger),
		customPrompt: nil,
	}
}

// NewSongSelectorAgentWithPrompt creates a new SongSelectorAgent with a custom system prompt.
func NewSongSelectorAgentWithPrompt(llmClient *openrouter.Client, model string, logger *zap.Logger, customPrompt *string) *SongSelectorAgent {
	return &SongSelectorAgent{
		BaseAgent:    NewBaseAgent(llmClient, model, logger),
		customPrompt: customPrompt,
	}
}

// getSystemPrompt returns the system prompt for the song selector agent.
func (a *SongSelectorAgent) getSystemPrompt() string {
	if a.customPrompt != nil && *a.customPrompt != "" {
		return *a.customPrompt
	}
	return DefaultSongSelectorPrompt
}

// Select chooses the best song from the candidates based on the original concept.
func (a *SongSelectorAgent) Select(ctx context.Context, input SongSelectorInput) (*SongSelectorOutput, error) {
	if len(input.Songs) == 0 {
		return nil, fmt.Errorf("no song candidates provided")
	}

	// If only one song, return it directly
	if len(input.Songs) == 1 {
		a.Logger().Info("only one song candidate, selecting it automatically",
			zap.String("song_id", input.Songs[0].ID),
			zap.String("title", input.Songs[0].Title),
		)
		return &SongSelectorOutput{
			SelectedSongID: input.Songs[0].ID,
			Reasoning:      "Only one song candidate available, selected automatically.",
		}, nil
	}

	// Build user prompt with song candidates
	userPrompt := a.buildUserPrompt(input)

	a.Logger().Debug("sending song selection request to LLM",
		zap.String("concept", input.OriginalConcept),
		zap.Int("candidate_count", len(input.Songs)),
	)

	// Call LLM
	response, err := a.LLMClient().ChatWithModel(ctx, a.Model(), a.getSystemPrompt(), userPrompt)
	if err != nil {
		a.Logger().Error("failed to call LLM for song selection",
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to call LLM: %w", err)
	}

	// Parse response
	output, err := a.parseResponse(response)
	if err != nil {
		a.Logger().Error("failed to parse LLM response",
			zap.Error(err),
			zap.String("response", response),
		)
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	// Validate selected song ID exists in candidates
	if !a.isValidSongID(output.SelectedSongID, input.Songs) {
		a.Logger().Error("LLM selected invalid song ID",
			zap.String("selected_id", output.SelectedSongID),
		)
		return nil, fmt.Errorf("selected song ID %q not found in candidates", output.SelectedSongID)
	}

	a.Logger().Info("song selected successfully",
		zap.String("selected_id", output.SelectedSongID),
		zap.String("reasoning", output.Reasoning),
	)

	return output, nil
}

// buildUserPrompt creates the user prompt with song candidates.
func (a *SongSelectorAgent) buildUserPrompt(input SongSelectorInput) string {
	var sb strings.Builder

	sb.WriteString("Original concept: ")
	sb.WriteString(input.OriginalConcept)
	sb.WriteString("\n\nSong candidates:\n")

	for _, song := range input.Songs {
		sb.WriteString(fmt.Sprintf("- ID: %s, Title: %q, Duration: %.1f seconds\n",
			song.ID, song.Title, song.Duration))
	}

	sb.WriteString("\nSelect the best song and explain your reasoning.")

	return sb.String()
}

// parseResponse parses the LLM response into SongSelectorOutput.
func (a *SongSelectorAgent) parseResponse(response string) (*SongSelectorOutput, error) {
	// Clean up response - remove markdown code blocks if present
	cleaned := strings.TrimSpace(response)
	if strings.HasPrefix(cleaned, "```json") {
		cleaned = strings.TrimPrefix(cleaned, "```json")
	} else if strings.HasPrefix(cleaned, "```") {
		cleaned = strings.TrimPrefix(cleaned, "```")
	}
	if strings.HasSuffix(cleaned, "```") {
		cleaned = strings.TrimSuffix(cleaned, "```")
	}
	cleaned = strings.TrimSpace(cleaned)

	var output SongSelectorOutput
	if err := json.Unmarshal([]byte(cleaned), &output); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	if output.SelectedSongID == "" {
		return nil, fmt.Errorf("selectedSongId is empty in response")
	}

	return &output, nil
}

// isValidSongID checks if the given ID exists in the song candidates.
func (a *SongSelectorAgent) isValidSongID(id string, songs []SongCandidate) bool {
	for _, song := range songs {
		if song.ID == id {
			return true
		}
	}
	return false
}
