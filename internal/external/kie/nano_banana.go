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

const (
	DefaultBaseURL     = "https://api.kie.ai"
	ModelNanoBananaPro = "nano-banana-pro"

	// Aspect ratios
	AspectRatio16x9 = "16:9"
	AspectRatio9x16 = "9:16"
	AspectRatio1x1  = "1:1"
	AspectRatio4x3  = "4:3"
	AspectRatio3x4  = "3:4"

	// Resolutions
	Resolution1K = "1K"
	Resolution2K = "2K"

	// Output formats
	FormatPNG  = "png"
	FormatJPG  = "jpg"
	FormatWEBP = "webp"

	// Polling configuration
	DefaultPollInterval = 3 * time.Second
	DefaultTimeout      = 5 * time.Minute
)

// NanoBananaClient is the client for KIE NanoBanana Pro API
type NanoBananaClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NanoInput represents the input parameters for image generation
type NanoInput struct {
	Prompt       string `json:"prompt"`
	AspectRatio  string `json:"aspect_ratio"`
	Resolution   string `json:"resolution"`
	OutputFormat string `json:"output_format"`
}

// CreateTaskRequest represents the request body for creating a task
type CreateTaskRequest struct {
	Model       string    `json:"model"`
	CallBackUrl string    `json:"callBackUrl,omitempty"`
	Input       NanoInput `json:"input"`
}

// CreateTaskResponse represents the response from creating a task
type CreateTaskResponse struct {
	Code int `json:"code"`
	Data struct {
		TaskId string `json:"taskId"`
	} `json:"data"`
}

// TaskStatusResponse represents the response from getting task status
type TaskStatusResponse struct {
	Code int `json:"code"`
	Data struct {
		Status string `json:"status"`
		Output struct {
			ImageUrl string `json:"imageUrl"`
		} `json:"output"`
	} `json:"data"`
}

// APIError represents an error response from the API
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("KIE API error (status %d): %s", e.StatusCode, e.Message)
}

// NewNanoBananaClient creates a new NanoBanana Pro API client
func NewNanoBananaClient(apiKey, baseURL string) *NanoBananaClient {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	return &NanoBananaClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateTask creates a new image generation task
func (c *NanoBananaClient) CreateTask(ctx context.Context, req CreateTaskRequest) (string, error) {
	// Ensure model is set
	if req.Model == "" {
		req.Model = ModelNanoBananaPro
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/jobs/createTask", bytes.NewReader(body))
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

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	var createResp CreateTaskResponse
	if err := json.Unmarshal(respBody, &createResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if createResp.Data.TaskId == "" {
		return "", fmt.Errorf("empty task ID in response")
	}

	return createResp.Data.TaskId, nil
}

// GetTask retrieves the status of a task
func (c *NanoBananaClient) GetTask(ctx context.Context, taskId string) (*TaskStatusResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/jobs/"+taskId, nil)
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
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	var statusResp TaskStatusResponse
	if err := json.Unmarshal(respBody, &statusResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &statusResp, nil
}

// WaitForCompletion polls the task status until it's completed or the timeout is reached
func (c *NanoBananaClient) WaitForCompletion(ctx context.Context, taskId string, timeout time.Duration) (*TaskStatusResponse, error) {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(DefaultPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for task completion: %w", ctx.Err())
		case <-ticker.C:
			status, err := c.GetTask(ctx, taskId)
			if err != nil {
				return nil, fmt.Errorf("failed to get task status: %w", err)
			}

			switch status.Data.Status {
			case StatusCompleted:
				return status, nil
			case StatusFailed:
				return status, fmt.Errorf("task failed")
			case StatusPending, StatusProcessing:
				// Continue polling
				continue
			default:
				return nil, fmt.Errorf("unknown task status: %s", status.Data.Status)
			}
		}
	}
}

// GenerateImage is a convenience method that creates a task, waits for completion, and returns the image URL
func (c *NanoBananaClient) GenerateImage(ctx context.Context, prompt string, aspectRatio string) (string, error) {
	if aspectRatio == "" {
		aspectRatio = AspectRatio16x9
	}

	req := CreateTaskRequest{
		Model: ModelNanoBananaPro,
		Input: NanoInput{
			Prompt:       prompt,
			AspectRatio:  aspectRatio,
			Resolution:   Resolution1K,
			OutputFormat: FormatPNG,
		},
	}

	taskId, err := c.CreateTask(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to create task: %w", err)
	}

	status, err := c.WaitForCompletion(ctx, taskId, DefaultTimeout)
	if err != nil {
		return "", fmt.Errorf("failed to wait for completion: %w", err)
	}

	if status.Data.Output.ImageUrl == "" {
		return "", fmt.Errorf("empty image URL in response")
	}

	return status.Data.Output.ImageUrl, nil
}
