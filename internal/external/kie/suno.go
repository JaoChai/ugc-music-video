package kie

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Suno model constants
const (
	ModelV3_5     = "V3_5"
	ModelV4       = "V4"
	ModelV4_5     = "V4_5"
	ModelV4_5Plus = "V4_5PLUS"
	ModelV5       = "V5"
)

// Task status constants
const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
)

// SunoClient represents a client for the KIE Suno API
type SunoClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// GenerateRequest represents the request body for generating music
type GenerateRequest struct {
	Prompt       string `json:"prompt"`
	CustomMode   bool   `json:"customMode"`
	Instrumental bool   `json:"instrumental"`
	Model        string `json:"model"`
	Style        string `json:"style,omitempty"`
	Title        string `json:"title,omitempty"`
	CallBackUrl  string `json:"callBackUrl,omitempty"`
}

// GenerateResponse represents the response from the generate endpoint
type GenerateResponse struct {
	Code int `json:"code"`
	Data struct {
		TaskId string `json:"taskId"`
	} `json:"data"`
}

// TaskResponse represents the response from the get task endpoint
type TaskResponse struct {
	Code int `json:"code"`
	Data struct {
		Status string     `json:"status"`
		Songs  []SongData `json:"songs"`
	} `json:"data"`
}

// SongData represents information about a generated song
type SongData struct {
	Id       string  `json:"id"`
	AudioUrl string  `json:"audioUrl"`
	Title    string  `json:"title"`
	Duration float64 `json:"duration"`
}

// NewSunoClient creates a new SunoClient with the given API key and base URL
func NewSunoClient(apiKey, baseURL string) *SunoClient {
	return &SunoClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Generate sends a music generation request and returns the task ID
func (c *SunoClient) Generate(ctx context.Context, req GenerateRequest) (string, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/generate", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var generateResp GenerateResponse
	if err := json.Unmarshal(respBody, &generateResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if generateResp.Code != 0 {
		return "", fmt.Errorf("API returned error code %d", generateResp.Code)
	}

	return generateResp.Data.TaskId, nil
}

// GetTask retrieves the status and results of a generation task
func (c *SunoClient) GetTask(ctx context.Context, taskId string) (*TaskResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/task/"+taskId, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var taskResp TaskResponse
	if err := json.Unmarshal(respBody, &taskResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if taskResp.Code != 0 {
		return nil, fmt.Errorf("API returned error code %d", taskResp.Code)
	}

	return &taskResp, nil
}

// WaitForCompletion polls the task status until it's completed or times out
func (c *SunoClient) WaitForCompletion(ctx context.Context, taskId string, timeout time.Duration) (*TaskResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for task completion: %w", ctx.Err())
		case <-ticker.C:
			taskResp, err := c.GetTask(ctx, taskId)
			if err != nil {
				return nil, fmt.Errorf("failed to get task status: %w", err)
			}

			switch taskResp.Data.Status {
			case StatusCompleted:
				return taskResp, nil
			case StatusFailed:
				return taskResp, fmt.Errorf("task failed")
			case StatusPending, StatusProcessing:
				// Continue polling
				continue
			default:
				return nil, fmt.Errorf("unknown task status: %s", taskResp.Data.Status)
			}
		}
	}
}
