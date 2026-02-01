// Package agents provides base agent functionality for LLM-based tasks.
package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/jaochai/ugc/internal/external/openrouter"
	"go.uber.org/zap"
)

// JSONOutputInstructions is a common prompt suffix for requesting JSON output.
const JSONOutputInstructions = "Respond with valid JSON only. No markdown, no explanation."

// Agent defines the interface for all agents.
type Agent interface {
	Execute(ctx context.Context, input interface{}) (interface{}, error)
}

// BaseAgent provides common functionality for LLM-based agents.
type BaseAgent struct {
	llmClient *openrouter.Client
	model     string
	logger    *zap.Logger
}

// NewBaseAgent creates a new BaseAgent instance.
func NewBaseAgent(llmClient *openrouter.Client, model string, logger *zap.Logger) *BaseAgent {
	return &BaseAgent{
		llmClient: llmClient,
		model:     model,
		logger:    logger,
	}
}

// LLMClient returns the OpenRouter client.
func (b *BaseAgent) LLMClient() *openrouter.Client {
	return b.llmClient
}

// Model returns the model name.
func (b *BaseAgent) Model() string {
	return b.model
}

// Logger returns the logger instance.
func (b *BaseAgent) Logger() *zap.Logger {
	return b.logger
}

// Chat sends a chat request with system and user prompts and returns the response content.
func (b *BaseAgent) Chat(ctx context.Context, systemPrompt string, userPrompt string) (string, error) {
	b.logger.Debug("sending chat request",
		zap.String("model", b.model),
		zap.Int("system_prompt_len", len(systemPrompt)),
		zap.Int("user_prompt_len", len(userPrompt)),
	)

	response, err := b.llmClient.ChatWithModel(ctx, b.model, systemPrompt, userPrompt)
	if err != nil {
		b.logger.Error("chat request failed", zap.Error(err))
		return "", fmt.Errorf("chat request failed: %w", err)
	}

	b.logger.Debug("chat request succeeded", zap.Int("response_len", len(response)))
	return response, nil
}

// ChatJSON sends a chat request and parses the JSON response into the result struct.
// It automatically appends JSONOutputInstructions to the system prompt.
func (b *BaseAgent) ChatJSON(ctx context.Context, systemPrompt string, userPrompt string, result interface{}) error {
	// Append JSON output instructions to system prompt
	fullSystemPrompt := systemPrompt + "\n\n" + JSONOutputInstructions

	response, err := b.Chat(ctx, fullSystemPrompt, userPrompt)
	if err != nil {
		return err
	}

	if err := b.ParseJSONFromResponse(response, result); err != nil {
		b.logger.Error("failed to parse JSON from response",
			zap.Error(err),
			zap.String("response", truncateString(response, 500)),
		)
		return fmt.Errorf("failed to parse JSON from response: %w", err)
	}

	return nil
}

// ParseJSONFromResponse extracts and parses JSON from an LLM response.
// It supports both raw JSON and JSON wrapped in markdown code blocks.
func (b *BaseAgent) ParseJSONFromResponse(response string, result interface{}) error {
	// Try to extract JSON from markdown code blocks first
	jsonStr := extractJSONFromMarkdown(response)

	// If no markdown block found, try the raw response
	if jsonStr == "" {
		jsonStr = strings.TrimSpace(response)
	}

	// Attempt to parse the JSON
	if err := json.Unmarshal([]byte(jsonStr), result); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return nil
}

// extractJSONFromMarkdown extracts JSON content from markdown code blocks.
// It handles ```json, ``` (plain), and tries to find JSON objects/arrays.
func extractJSONFromMarkdown(response string) string {
	// Pattern 1: ```json ... ```
	jsonBlockPattern := regexp.MustCompile("(?s)```json\\s*\\n?(.*?)\\n?```")
	if matches := jsonBlockPattern.FindStringSubmatch(response); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Pattern 2: ``` ... ``` (generic code block)
	genericBlockPattern := regexp.MustCompile("(?s)```\\s*\\n?(.*?)\\n?```")
	if matches := genericBlockPattern.FindStringSubmatch(response); len(matches) > 1 {
		content := strings.TrimSpace(matches[1])
		// Verify it looks like JSON
		if isJSONLike(content) {
			return content
		}
	}

	// Pattern 3: Find raw JSON object or array in the response
	response = strings.TrimSpace(response)
	if isJSONLike(response) {
		return response
	}

	return ""
}

// isJSONLike checks if a string appears to be a JSON object or array.
func isJSONLike(s string) bool {
	s = strings.TrimSpace(s)
	return (strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")) ||
		(strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]"))
}

// truncateString truncates a string to maxLen characters, adding "..." if truncated.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
