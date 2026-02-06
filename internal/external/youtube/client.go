// Package youtube provides a client for the YouTube Data API v3.
package youtube

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// UploadInput holds the parameters for uploading a video to YouTube.
type UploadInput struct {
	Title       string
	Description string
	VideoReader io.Reader
}

// UploadResult holds the result of a successful YouTube upload.
type UploadResult struct {
	VideoID  string
	VideoURL string
}

// Client wraps the YouTube Data API operations.
type Client struct {
	oauthConfig *oauth2.Config
	httpClient  *http.Client
	logger      *zap.Logger
}

// NewClient creates a new YouTube API client.
func NewClient(clientID, clientSecret, redirectURI string, logger *zap.Logger) *Client {
	cfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
		Scopes:       []string{"https://www.googleapis.com/auth/youtube.upload"},
		Endpoint:     google.Endpoint,
	}

	return &Client{
		oauthConfig: cfg,
		httpClient:  &http.Client{Timeout: 30 * time.Minute},
		logger:      logger,
	}
}

// GetAuthURL generates the OAuth2 consent URL with a state parameter for CSRF protection.
func (c *Client) GetAuthURL(state string) string {
	return c.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent"))
}

// ExchangeCode exchanges an authorization code for OAuth2 tokens.
// Returns the refresh token for long-term storage.
func (c *Client) ExchangeCode(ctx context.Context, code string) (string, error) {
	token, err := c.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return "", fmt.Errorf("failed to exchange code: %w", err)
	}

	if token.RefreshToken == "" {
		return "", fmt.Errorf("no refresh token received; user may need to re-authorize")
	}

	return token.RefreshToken, nil
}

// UploadVideo uploads a video to YouTube using a stored refresh token.
// Privacy is set to unlisted.
func (c *Client) UploadVideo(ctx context.Context, refreshToken string, input UploadInput) (*UploadResult, error) {
	// Create token source from refresh token
	token := &oauth2.Token{RefreshToken: refreshToken}
	tokenSource := c.oauthConfig.TokenSource(ctx, token)

	// Create YouTube service
	svc, err := youtube.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube service: %w", err)
	}

	// Build video metadata
	video := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       input.Title,
			Description: input.Description,
			CategoryId:  "10", // Music category
		},
		Status: &youtube.VideoStatus{
			PrivacyStatus: "unlisted",
		},
	}

	// Upload video (resumable upload is handled by the library)
	call := svc.Videos.Insert([]string{"snippet", "status"}, video)
	call.Media(input.VideoReader)

	resp, err := call.Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to upload video: %w", err)
	}

	c.logger.Info("video uploaded to YouTube",
		zap.String("video_id", resp.Id),
		zap.String("title", input.Title),
	)

	return &UploadResult{
		VideoID:  resp.Id,
		VideoURL: fmt.Sprintf("https://www.youtube.com/watch?v=%s", resp.Id),
	}, nil
}

// RevokeToken revokes the given refresh token.
func (c *Client) RevokeToken(ctx context.Context, refreshToken string) error {
	revokeURL := fmt.Sprintf("https://oauth2.googleapis.com/revoke?token=%s", url.QueryEscape(refreshToken))

	req, err := http.NewRequestWithContext(ctx, "POST", revokeURL, strings.NewReader(""))
	if err != nil {
		return fmt.Errorf("failed to create revoke request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("revoke failed with status %d: %s", resp.StatusCode, string(body))
	}

	c.logger.Info("YouTube token revoked")
	return nil
}
